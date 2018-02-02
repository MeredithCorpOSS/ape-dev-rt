package backends

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	rtAWS "github.com/TimeInc/ape-dev-rt/aws"
	"github.com/TimeInc/ape-dev-rt/deploymentstate/schema"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	s3_appPrefix    = "%s/"
	s3_appObjectKey = "%s/%s/APPLICATION.json"

	s3_slotObjectPrefix = "%s/%s/SLOT-"
	s3_slotObjectSuffix = ".json"
	s3_slotKey          = "%s/%s/SLOT-%s%s"

	s3_deploymentsPrefix       = "%s/%s/DEPLOYMENT-"
	s3_deploymentPerSlotPrefix = "%s/%s/DEPLOYMENT-%s-"
	s3_deploymentKey           = "%s/%s/DEPLOYMENT-%s-%s%s"
	s3_deploymentKeySuffix     = ".json"

	defaultContentType = "application/json"
	defaultAcl         = "bucket-owner-read"
)

type S3 struct{}

type S3Config struct {
	s3conn *awsS3.S3

	ApiCallerArn string
	Prefix       string
	Bucket       string
	Region       string
}

func (s3 *S3) Configure(config map[string]interface{}) (interface{}, error) {
	cfg := S3Config{}

	if v, ok := config["prefix"]; ok {
		cfg.Prefix = strings.TrimSuffix(v.(string), "/")
	} else {
		return nil, fmt.Errorf("Unable to find `prefix` in config")
	}

	if v, ok := config["bucket"]; ok {
		cfg.Bucket = v.(string)
	} else {
		return nil, fmt.Errorf("Unable to find `bucket` in config")
	}

	if v, ok := config["region"]; ok {
		cfg.Region = v.(string)
	} else {
		return nil, fmt.Errorf("Unable to find `region` in config")
	}

	profile, ok := config["profile"]
	if !ok {
		profile = ""
	}

	creds := rtAWS.CredentialsProvider(profile.(string))

	sess := session.New(&aws.Config{
		Credentials: creds,
		Region:      aws.String(cfg.Region),
	})
	cfg.s3conn = awsS3.New(sess)
	stsconn := sts.New(sess)

	out, err := stsconn.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("Failed getting AWS API caller identity: %s", err)
	}
	cfg.ApiCallerArn = *out.Arn

	return &cfg, nil
}

// This will allow us to print some useful warnings to user,
// e.g. "Hey, backend %s doesn't support locking, it is your responsibility
// to let your colleagues know you're deploying!"
func (s3 *S3) SupportsWriteLock() bool {
	return false
}

func (s3 *S3) IsReady(meta interface{}) (bool, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn

	// This won't catch all permission related issues
	// (user may have Get, but not Put perm)
	// There isn't way to check it w/out actually executing Put :/
	input := awsS3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(cfg.Prefix),
	}
	_, err := conn.GetObject(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchKey" {
			log.Printf("[DEBUG] Ignoring 404 (NoSuchKey) error, assuming S3 backend is ready")
			return true, nil
		}
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "AccessDenied" {
			return false, fmt.Errorf("Failed accessing %s (bucket): %s (prefix). %s",
				cfg.Bucket, cfg.Prefix, awsErr.Error())
		}

		return false, err
	}

	return true, nil
}

func (s3 *S3) GetApplication(meta interface{}, appName string) (*schema.ApplicationData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	log.Printf("[DEBUG] prefix = %s, app name = %s", cfg.Prefix, appName)
	key := s3.buildAppKey(cfg.Prefix, appName)

	input := awsS3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	}
	log.Printf("[DEBUG] Getting application %q from S3: %s", appName, input)
	out, err := conn.GetObject(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchKey" {
			return nil, &AppNotFound{AppName: appName, OriginalErr: err}
		}
		return nil, fmt.Errorf("Failed to get application %q from S3: %s", appName, err)
	}
	log.Printf("[DEBUG] Received application %q from S3: %q (Etag: %s, VersionId: %#v)",
		appName, key, *out.ETag, out.VersionId)

	data, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}

	app := &schema.ApplicationData{}
	err = app.FromJSON(data)
	if err != nil {
		return nil, err
	}
	app.Name = appName

	return app, nil
}

func (s3 *S3) SaveApplication(meta interface{}, name string, app *schema.ApplicationData) error {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildAppKey(cfg.Prefix, name)

	appDataInBytes, err := app.ToJSON()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Saving app data into S3. Bucket: %q, Key: %q", cfg.Bucket, key)
	input := awsS3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(appDataInBytes),
		ContentType: aws.String(defaultContentType),
		ACL:         aws.String(defaultAcl),
	}
	out, err := conn.PutObject(&input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Written app data to S3: %q (Etag: %s, VersionId: %#v)",
		key, *out.ETag, out.VersionId)

	return nil
}

func (s3 *S3) ListApplications(meta interface{}) ([]*schema.ApplicationData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	prefix := s3.buildAppPrefix(cfg.Prefix)

	var apps = make([]*schema.ApplicationData, 0)
	var pageNum = 0
	paginateFunc := func(page *awsS3.ListObjectsOutput, lastPage bool) bool {
		objects := page.Contents
		log.Printf("[DEBUG] Ranging over %d apps (page %d):\n%s\n\n\n", len(objects), pageNum, objects)
		for _, o := range objects {
			log.Printf("[DEBUG] Pulling app from S3: %q (%d) w/ Etag %q", *o.Key, *o.Size, *o.ETag)
			appName, ok := s3.getAppKey(*o.Key, prefix)
			if !ok {
				continue
			}
			input := awsS3.GetObjectInput{
				Bucket: aws.String(cfg.Bucket),
				Key:    o.Key,
			}
			log.Printf("[DEBUG] Pulling S3 object: %s", input)
			out, err := conn.GetObject(&input)
			if err != nil {
				log.Printf("[ERROR] Failed to download %q from S3: %s", *o.Key, err)
				return false
			}

			data, err := ioutil.ReadAll(out.Body)
			if err != nil {
				log.Printf("[ERROR] Failed to read bytes of %q: %s", *o.Key, err)
				return false
			}

			app := &schema.ApplicationData{}
			err = app.FromJSON(data)
			if err != nil {
				log.Printf("[ERROR] Failed to unmarshal %q: %s", *o.Key, err)
				return false
			}

			app.Name = appName
			apps = append(apps, app)
		}
		pageNum++
		return !lastPage
	}

	input := awsS3.ListObjectsInput{
		Bucket: aws.String(cfg.Bucket),
		Prefix: aws.String(prefix),
	}
	log.Printf("[DEBUG] Listing apps in S3: %s", input)
	err := conn.ListObjectsPages(&input, paginateFunc)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (s3 *S3) ListSlots(meta interface{}, appName string) ([]*schema.SlotData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	prefix := s3.buildSlotPrefix(cfg.Prefix, appName)

	var slots = make([]*schema.SlotData, 0)
	var pageNum = 0
	paginateFunc := func(page *awsS3.ListObjectsOutput, lastPage bool) bool {
		objects := page.Contents
		log.Printf("[DEBUG] Ranging over %d slots (page %d)", len(objects), pageNum)
		for _, o := range objects {
			log.Printf("[DEBUG] Pulling slot from S3: %q (%d) w/ Etag %q", *o.Key, *o.Size, *o.ETag)
			input := awsS3.GetObjectInput{
				Bucket: aws.String(cfg.Bucket),
				Key:    o.Key,
			}
			out, err := conn.GetObject(&input)
			if err != nil {
				log.Printf("[ERROR] Failed to download %q from S3: %s", *o.Key, err)
				return false
			}

			data, err := ioutil.ReadAll(out.Body)
			if err != nil {
				log.Printf("[ERROR] Failed to read bytes of %q: %s", *o.Key, err)
				return false
			}

			slot := &schema.SlotData{}
			err = slot.FromJSON(data)
			if err != nil {
				log.Printf("[ERROR] Failed to unmarshal %q: %s", *o.Key, err)
				return false
			}

			slotId := strings.TrimPrefix(*o.Key, prefix)
			slotId = strings.TrimSuffix(slotId, s3_slotObjectSuffix)
			slot.SlotId = slotId
			slots = append(slots, slot)
		}
		pageNum++
		return !lastPage
	}

	input := awsS3.ListObjectsInput{
		Bucket: aws.String(cfg.Bucket),
		Prefix: aws.String(prefix),
	}
	log.Printf("[DEBUG] Listing slots in S3: %s", input)
	err := conn.ListObjectsPages(&input, paginateFunc)
	if err != nil {
		return nil, err
	}

	return slots, nil
}

func (s3 *S3) SaveSlot(meta interface{}, appName, slotId string, slot *schema.SlotData) error {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildSlotKey(cfg.Prefix, appName, slotId)

	slotDataInBytes, err := slot.ToJSON()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Saving slot data into S3. Bucket: %q, Key: %q", cfg.Bucket, key)
	input := awsS3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(slotDataInBytes),
		ContentType: aws.String(defaultContentType),
		ACL:         aws.String(defaultAcl),
	}
	out, err := conn.PutObject(&input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Written slot data to S3: %q (Etag: %s, VersionId: %#v)",
		key, *out.ETag, out.VersionId)

	return nil
}

func (s3 *S3) DeleteSlot(meta interface{}, appName, slotId string) error {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildSlotKey(cfg.Prefix, appName, slotId)

	input := awsS3.DeleteObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	}
	log.Printf("[DEBUG] Deleting slot from S3: %s", input)
	_, err := conn.DeleteObject(&input)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleted slot from S3: %q", key)
	return nil
}

func (s3 *S3) GetSlot(meta interface{}, appName, slotId string) (*schema.SlotData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildSlotKey(cfg.Prefix, appName, slotId)

	input := awsS3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	}
	log.Printf("[DEBUG] Getting slot from S3: %s", input)
	out, err := conn.GetObject(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchKey" {
			return nil, &SlotNotFound{SlotName: slotId, OriginalErr: err}
		}
		return nil, err
	}
	log.Printf("[DEBUG] Received slot from S3: %q (Etag: %s, VersionId: %#v)",
		key, *out.ETag, out.VersionId)

	data, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}

	slot := &schema.SlotData{}
	err = slot.FromJSON(data)
	if err != nil {
		return nil, err
	}
	slot.SlotId = slotId

	return slot, nil
}

func (s3 *S3) ListSortedDeploymentsForSlotId(meta interface{}, appName, slotId string, limitPerSlot int) ([]*schema.DeploymentData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	prefix := s3.buildDeploymentPerSlotKey(cfg.Prefix, appName, slotId)

	input := awsS3.ListObjectsInput{
		Bucket: aws.String(cfg.Bucket),
		Prefix: aws.String(prefix),
	}

	var deployments = make([]*schema.DeploymentData, 0)
	var pageNum = 0
	paginateFunc := func(page *awsS3.ListObjectsOutput, lastPage bool) bool {
		objects := page.Contents
		log.Printf("[DEBUG] Ranging over %d deployments (page %d)", len(objects), pageNum)
		for _, o := range objects {
			log.Printf("[DEBUG] Pulling deployment from S3: %q (%d) w/ Etag %q", *o.Key, *o.Size, *o.ETag)
			input := awsS3.GetObjectInput{
				Bucket: aws.String(cfg.Bucket),
				Key:    o.Key,
			}
			out, err := conn.GetObject(&input)
			if err != nil {
				log.Printf("[ERROR] Failed to download %q from S3: %s", *o.Key, err)
				return false
			}

			data, err := ioutil.ReadAll(out.Body)
			if err != nil {
				log.Printf("[ERROR] Failed to read bytes of %q: %s", *o.Key, err)
				return false
			}

			deployment := &schema.DeploymentData{}
			err = deployment.FromJSON(data)
			if err != nil {
				log.Printf("[ERROR] Failed to unmarshal %q: %s", *o.Key, err)
				return false
			}
			deploymentId := strings.TrimPrefix(*o.Key, prefix)
			deploymentId = strings.TrimSuffix(deploymentId, s3_deploymentKeySuffix)
			deployment.DeploymentId = deploymentId
			deployments = append(deployments, deployment)

			if len(deployments) == limitPerSlot {
				return false
			}
		}
		pageNum++

		return !lastPage
	}

	log.Printf("[DEBUG] Listing deployments in S3: %s", input)
	err := conn.ListObjectsPages(&input, paginateFunc)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (s3 *S3) SaveDeployment(meta interface{}, appName, slotId, deploymentId string, data *schema.DeploymentData) error {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildDeploymentKey(cfg.Prefix, appName, slotId, deploymentId)

	deploymentDataInBytes, err := data.ToJSON()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Saving deployment data into S3. Bucket: %q, Key: %q", cfg.Bucket, key)
	input := awsS3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(deploymentDataInBytes),
		ContentType: aws.String(defaultContentType),
		ACL:         aws.String(defaultAcl),
	}
	out, err := conn.PutObject(&input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Written deployment data to S3: %q (Etag: %s, VersionId: %#v)",
		key, *out.ETag, out.VersionId)

	return nil
}

func (s3 *S3) GetDeployment(meta interface{}, appName, slotId, deploymentId string) (*schema.DeploymentData, error) {
	cfg := meta.(*S3Config)
	conn := cfg.s3conn
	key := s3.buildDeploymentKey(cfg.Prefix, appName, slotId, deploymentId)

	input := awsS3.GetObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String(key),
	}
	log.Printf("[DEBUG] Getting deployment from S3: %s", input)
	out, err := conn.GetObject(&input)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Received deployment from S3: %q (Etag: %s, VersionId: %#v)",
		key, *out.ETag, out.VersionId)

	data, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}

	deployment := &schema.DeploymentData{}
	err = deployment.FromJSON(data)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func (s3 *S3) getAppKey(key, s3Prefix string) (string, bool) {
	exp := fmt.Sprintf(s3_appObjectKey, strings.TrimSuffix(s3Prefix, "/"), "([^/]+)")
	re := regexp.MustCompile(exp)
	matches := re.FindStringSubmatch(key)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], true
}

func (s3 *S3) buildAppPrefix(s3Prefix string) string {
	return fmt.Sprintf(s3_appPrefix, s3Prefix)
}

func (s3 *S3) buildAppKey(s3Prefix, appName string) string {
	appName = strings.Trim(appName, "/")
	return fmt.Sprintf(s3_appObjectKey, s3Prefix, appName)
}

func (s3 *S3) buildSlotPrefix(s3Prefix, appName string) string {
	return fmt.Sprintf(s3_slotObjectPrefix, s3Prefix, appName)
}

func (s3 *S3) buildSlotKey(s3Prefix, appName, slotId string) string {
	return fmt.Sprintf(s3_slotKey, s3Prefix, appName, slotId, s3_slotObjectSuffix)
}

func (s3 *S3) buildDeploymentPerSlotKey(s3Prefix, appName, slotId string) string {
	return fmt.Sprintf(s3_deploymentPerSlotPrefix, s3Prefix, appName, slotId)
}

func (s3 *S3) buildDeploymentKey(s3Prefix, appName, slotId, deploymentId string) string {
	return fmt.Sprintf(s3_deploymentKey, s3Prefix, appName, slotId, deploymentId, s3_deploymentKeySuffix)
}
