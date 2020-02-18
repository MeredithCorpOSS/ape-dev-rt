package backends

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	rtAWS "github.com/TimeIncOSS/ape-dev-rt/aws"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestIsReady(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	isReady, err := s3.IsReady(s)
	if err != nil {
		t.Fatal(err)
	}
	if !isReady {
		t.Fatal("Expected S3 backend to be ready")
	}
}

func TestSaveAndGetSlot(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	insertedSlotData := &schema.SlotData{
		SchemaVersion:           1,
		SlotId:                  "BLUE",
		IsActive:                true,
		LastDeploymentStartTime: timestamp,
	}
	err = s3.SaveSlot(s, "FindingUmar", "BLUE", insertedSlotData)
	if err != nil {
		t.Fatal(err)
	}

	receivedSlotData, err := s3.GetSlot(s, "FindingUmar", "BLUE")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*insertedSlotData, *receivedSlotData) {
		t.Fatalf("Expected slot data to match.\nInserted: %v\nReceived: %v",
			*insertedSlotData, *receivedSlotData)
	}
}

func TestSaveAndDeleteSlot(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	insertedSlotData := &schema.SlotData{
		SchemaVersion:           0,
		SlotId:                  "BLUE",
		IsActive:                true,
		LastDeploymentStartTime: timestamp,
	}
	err = s3.SaveSlot(s, "FindingUmar", "BLUE", insertedSlotData)
	if err != nil {
		t.Fatal(err)
	}

	err = s3.DeleteSlot(s, "FindingUmar", "BLUE")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s3.GetSlot(s, "FindingUmar", "BLUE")
	if err == nil {
		t.Fatal("Expected 404/error when getting deleted slot, got none")
	}
	if _, ok := err.(*SlotNotFound); ok {
		// Slot expected to be gone
		return
	}
	t.Fatal(err)
}

func TestSaveAndGetDeployment_single(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	insertedDeploymentData := &schema.DeploymentData{
		RTVersion:     "1.0",
		SchemaVersion: 1,
		StartTime:     timestamp,
	}
	uniqueID := "1234567890"
	err = s3.SaveDeployment(s, "BloodyHell", "NEW", uniqueID, insertedDeploymentData)
	if err != nil {
		t.Fatal(err)
	}

	receivedDeploymentData, err := s3.GetDeployment(s, "BloodyHell", "NEW", uniqueID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*insertedDeploymentData, *receivedDeploymentData) {
		t.Fatalf("Expected deployment data to match.\nInserted: %#v\nReceived: %#v",
			*insertedDeploymentData, *receivedDeploymentData)
	}
}

func TestSaveAndListSlots(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	for i := 100; i < 145; i++ {
		slotId := fmt.Sprintf("RANDOM_%d", i)
		slotData := &schema.SlotData{
			SchemaVersion:           1,
			IsActive:                true,
			LastDeploymentStartTime: timestamp,
		}
		err := s3.SaveSlot(s, "CookieMonster", slotId, slotData)
		if err != nil {
			t.Fatal(err)
		}
	}

	slots, err := s3.ListSlots(s, "CookieMonster")
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 45 {
		t.Fatalf("Expected exactly 45 slots, returned %d.", len(slots))
	}
	expectedSlotData := schema.SlotData{
		SchemaVersion:           1,
		IsActive:                true,
		SlotId:                  "RANDOM_143",
		LastDeploymentStartTime: timestamp,
	}
	if !reflect.DeepEqual(*slots[43], expectedSlotData) {
		t.Fatalf("Expected slot data to match.\nGiven: %#v\nExpected: %#v",
			*slots[43], expectedSlotData)
	}
}

func TestGetSlot_notFound(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	_, err = s3.GetSlot(s, "my-special-app", "newslot")
	if err == nil {
		t.Fatal("Expected error when getting non-existent slot")
	}
	_, ok := err.(*SlotNotFound)
	if !ok {
		t.Fatalf("Expected SlotNotFound error, received: %s", err)
	}
}

func TestGetApplication_notFound(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	_, err = s3.GetApplication(s, "my-special-app")
	if err == nil {
		t.Fatal("Expected error when getting non-existent slot")
	}
	_, ok := err.(*AppNotFound)
	if !ok {
		t.Fatalf("Expected AppNotFound error, received: %s", err)
	}
}

func TestSaveAndGetApplication(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	data := &schema.ApplicationData{
		UseCentralGitRepo:    false,
		IsActive:             true,
		InfraOutputs:         map[string]string{"colour": "blue"},
		LastRtVersion:        "1.2.3",
		LastTerraformVersion: "0.7.2",
		LastDeploymentTime:   timestamp,
		LastInfraChangeTime:  timestamp,
	}
	err = s3.SaveApplication(s, "brandnewapp", data)
	if err != nil {
		t.Fatalf("Unexpected error when saving app: %s", err)
	}
	app, err := s3.GetApplication(s, "brandnewapp")
	if err != nil {
		t.Fatalf("Unexpected error when getting app: %s", err)
	}
	if app == nil {
		t.Fatal("Expected app data, received nil instead.")
	}
	app.Name = data.Name
	if !reflect.DeepEqual(*app, *data) {
		t.Fatalf("App data don't match.\nExpected: %#v\nReceived: %#v\n", *data, *app)
	}
}

func TestListApplications(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")

	// Save 10 apps
	for i := 0; i < 10; i++ {
		err := s3.SaveApplication(s, fmt.Sprintf("rt-test-%d", i), &schema.ApplicationData{
			UseCentralGitRepo:    false,
			IsActive:             true,
			InfraOutputs:         map[string]string{"order": fmt.Sprintf("%d", i)},
			LastRtVersion:        "1.2.4",
			LastTerraformVersion: "0.7.1",
			LastDeploymentTime:   timestamp,
			LastInfraChangeTime:  timestamp,
		})
		if err != nil {
			t.Fatalf("Unexpected error when saving app: %s", err)
		}
	}

	// Retrieve 10 apps
	apps, err := s3.ListApplications(s)
	if err != nil {
		t.Fatalf("Unexpected error when listing apps: %s", err)
	}
	expectedLen := 10
	if l := len(apps); l != expectedLen {
		t.Fatalf("Expected %d apps, %d received.", expectedLen, l)
	}
}

func TestSaveAndListDeployments(t *testing.T) {
	s, setUp, tearDown, err := testAccS3Setup()
	if err != nil {
		t.Skip(err)
	}
	err = setUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDown()

	s3 := &S3{}
	slotId := "RANDOMACCTEST"
	timestamp, _ := time.Parse(time.RFC1123, "Wed, 30 Mar 2016 15:04:05 BST")
	for i := 100; i < 120; i++ {
		uniqueID := fmt.Sprintf("1234567%d", i)
		dData := &schema.DeploymentData{
			SchemaVersion: 1,
			DeployPilot: &schema.DeployPilot{
				AWSApiCaller: "arn:aws:iam::123456789012:user/FishSticker",
			},
			StartTime: timestamp,
			Terraform: &schema.TerraformRun{
				PlanStartTime:  timestamp,
				PlanFinishTime: timestamp,
				StartTime:      timestamp,
				FinishTime:     timestamp,
				Variables: map[string]string{
					"colour": "green",
					"umar":   "FishSticks",
					"index":  fmt.Sprintf("%d", i),
				},
				TerraformVersion: "0.7.0",
			},
		}
		err := s3.SaveDeployment(s, "Bollocks", slotId, uniqueID, dData)
		if err != nil {
			t.Fatal(err)
		}
	}

	deployments, err := s3.ListSortedDeploymentsForSlotId(s, "Bollocks", slotId, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments) != 10 {
		t.Fatalf("Expected exactly 10 deployments, returned %d.", len(deployments))
	}
	expectedDeploymentData := &schema.DeploymentData{
		SchemaVersion: 1,
		DeploymentId:  "1234567106",
		DeployPilot: &schema.DeployPilot{
			AWSApiCaller: "arn:aws:iam::123456789012:user/FishSticker",
		},
		StartTime: timestamp,
		Terraform: &schema.TerraformRun{
			PlanStartTime:  timestamp,
			PlanFinishTime: timestamp,
			StartTime:      timestamp,
			FinishTime:     timestamp,
			Variables: map[string]string{
				"colour": "green",
				"umar":   "FishSticks",
				"index":  "106",
			},
			TerraformVersion: "0.7.0",
		},
	}
	if !reflect.DeepEqual(*deployments[6], *expectedDeploymentData) {
		t.Fatalf("Expected slot data to match.\nGiven:    %#v\nExpected: %#v",
			*deployments[6], *expectedDeploymentData)
	}
}

func testAccS3Setup() (*S3Config, func() error, func(), error) {
	profileName := os.Getenv("RT_ACC_AWS_PROFILE")
	if profileName == "" {
		return nil, nil, nil, fmt.Errorf("Please set AWS profile name (RT_ACC_AWS_PROFILE)")
	}
	creds := rtAWS.CredentialsProvider(profileName)
	sess := session.New(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-west-2"),
	})
	s3conn := s3.New(sess)
	cfg := &S3Config{
		s3conn:       s3conn,
		ApiCallerArn: "arn:aws:iam::123456789012:user/FishSticker",
		Prefix:       "my-account/dev",
		Bucket:       "ti-rt-acc-test-" + generateRandomId(),
		Region:       "us-west-2",
	}

	setUp := func() error {
		log.Printf("[DEBUG] Creating temporary bucket %q", cfg.Bucket)
		_, err := cfg.s3conn.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cfg.Bucket),
		})
		return err
	}

	tearDown := func() {
		log.Printf("[DEBUG] Emptying & deleting temporary bucket %q", cfg.Bucket)
		listInput := s3.ListObjectsInput{
			Bucket: aws.String(cfg.Bucket),
		}
		var keysForDeletion = make([]*s3.ObjectIdentifier, 0)
		err := s3conn.ListObjectsPages(&listInput, func(page *s3.ListObjectsOutput, lastPage bool) bool {
			keys := page.Contents
			for _, k := range keys {
				keysForDeletion = append(keysForDeletion, &s3.ObjectIdentifier{
					Key: k.Key,
				})
			}
			return !lastPage
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(keysForDeletion) > 0 {
			objDelInput := s3.DeleteObjectsInput{
				Bucket: aws.String(cfg.Bucket),
				Delete: &s3.Delete{
					Objects: keysForDeletion,
				},
			}
			log.Printf("[DEBUG] Emptying bucket: %s\n", objDelInput)
			_, err = cfg.s3conn.DeleteObjects(&objDelInput)
			if err != nil {
				log.Fatal(err)
			}
		}

		bucketDelInput := s3.DeleteBucketInput{
			Bucket: aws.String(cfg.Bucket),
		}
		log.Printf("[DEBUG] Deleting bucket: %s\n", bucketDelInput)
		_, err = cfg.s3conn.DeleteBucket(&bucketDelInput)
		if err != nil {
			log.Fatal(err)
		}
	}

	return cfg, setUp, tearDown, nil
}

func generateRandomId() string {
	rInt := rand.Int()
	return fmt.Sprintf("%d", rInt)
}
