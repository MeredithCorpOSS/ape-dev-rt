package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate"
	"github.com/TimeIncOSS/ape-dev-rt/deploymentstate/schema"
	"github.com/TimeIncOSS/ape-dev-rt/hcl"
	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mitchellh/go-homedir"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
)

var boldGreen = chalk.Green.NewStyle().WithTextStyle(chalk.Bold).Style
var boldWhite = chalk.White.NewStyle().WithTextStyle(chalk.Bold).Style
var boldRed = chalk.Red.NewStyle().WithTextStyle(chalk.Bold).Style

type TfState struct {
	Version int
	Serial  int
	Remote  *TfStateRemote
	Modules []*TfStateModule
}

type TfStateRemote struct {
	Type   string
	Config map[string]string
}

type TfStateModule struct {
	Path      []string
	Outputs   map[string]string
	Resources map[string]interface{}
}

func main() {
	app := cli.NewApp()
	app.Name = "Migration Commands"
	app.Usage = "One off tasks"
	app.Version = "v0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "migrate-tf-remote-state",
			Usage: "Migrate the remote tf state",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "env",
					Usage: "Specify the environment you wish to run the migration on",
				},
				cli.StringFlag{
					Name:  "app",
					Usage: "Optional: Specify a single application you wish to migrate",
				},
				cli.StringFlag{
					Name:  "region",
					Usage: "Optional: Defaults to us-east-1",
					Value: "us-east-1",
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "Verbose output",
				},
				cli.StringFlag{
					Name:  "aws-profile",
					Usage: "Optional: Defaults to default",
					Value: "default",
				},
			},
			Action: migrateTfRemoteState(),
			Before: beforeMigration,
		},
		{
			Name:  "rollback-migrated-tf-remote-state",
			Usage: "Roll back the migrated the remote tf state",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "env",
					Usage: "Specify the environment you wish to run the migration on",
				},
				cli.StringFlag{
					Name:  "app",
					Usage: "Optional: Specify a single application you wish to migrate",
				},
				cli.StringFlag{
					Name:  "region",
					Usage: "Optional: Defaults to us-east-1",
					Value: "us-east-1",
				},
				cli.StringFlag{
					Name:  "aws-profile",
					Usage: "Optional: Defaults to default",
					Value: "default",
				},
			},
			Action: rollbackMigratedTfRemoteState(),
			Before: beforeMigration,
		},
		{
			Name:  "prepare-deploymentstate",
			Usage: "Prepopulate the deployment state for existing apps",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "app",
					Usage: "Specify the application you wish to run the migration on",
				},
				cli.StringFlag{
					Name:  "env",
					Usage: "Specify the environment you wish to run the migration on",
				},
				cli.StringFlag{
					Name:  "region",
					Usage: "Optional: Defaults to us-east-1",
					Value: "us-east-1",
				},
				cli.StringFlag{
					Name:  "aws-profile",
					Usage: "Optional: Defaults to default",
					Value: "default",
				},
			},
			Action: prepareDeploymentState(),
			Before: beforeMigration,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[Error] "+err.Error())
	}
}

func listObjectsInS3Bucket(svc *s3.S3, bucket string) ([]*s3.Object, error) {
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: awsSDK.String(bucket),
	}
	objects, err := svc.ListObjectsV2(listObjectsInput)
	if err != nil {
		return nil, err
	}

	return objects.Contents, nil
}

func getStateBucketForEnv(env string) string {
	return fmt.Sprintf("ti-terraform-state-%s", env)
}

func beforeMigration(c *cli.Context) error {
	os.Setenv("AWS_PROFILE", c.String("aws-profile"))

	if len(c.String("env")) == 0 {
		return fmt.Errorf("You must specify an environment")
	}

	region := c.String("region")

	sess := session.New(&awsSDK.Config{Region: awsSDK.String(region)})

	stsSvc := sts.New(sess)
	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	c.App.Metadata = make(map[string]interface{}, 0)
	c.App.Metadata["identity"] = identity

	svc := s3.New(sess, &awsSDK.Config{S3ForcePathStyle: awsSDK.Bool(true)})
	c.App.Metadata["s3conn"] = svc

	return nil
}

type objectMoveFn func(sourceKey string) (bool, string)
type remoteBackendKeyMigrator func(origKey string) string

func migrateTfStateFiles(s3conn *s3.S3, bucket string, fn objectMoveFn, rbkmFn remoteBackendKeyMigrator, verbose bool) error {
	// List the objects in the bucket
	objects, err := listObjectsInS3Bucket(s3conn, bucket)
	if err != nil {
		return err
	}

	// We have some objects to migrate
	for _, obj := range objects {
		key := *obj.Key
		matched, destKey := fn(key)
		if !matched {
			if verbose {
				fmt.Printf("Skipping %s ...\n", key)
			}
			continue
		}

		fmt.Printf(boldWhite("Attempting to migrate %s to %s\n"), key, destKey)

		// Migrate the tfstate
		getInput := &s3.GetObjectInput{
			Bucket: awsSDK.String(bucket),
			Key:    obj.Key,
		}
		out, err := s3conn.GetObject(getInput)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(out.Body)
		if err != nil {
			return err
		}
		state := TfState{}
		err = json.Unmarshal(b, &state)
		if err != nil {
			return err
		}
		v, ok := state.Remote.Config["key"]
		if !ok {
			return fmt.Errorf("No remote backend key found when migrating %s", key)
		}

		state.Remote.Config["key"] = rbkmFn(v)
		newB, err := json.Marshal(state)
		if err != nil {
			return err
		}

		putInput := &s3.PutObjectInput{
			Bucket:      awsSDK.String(bucket),
			Key:         awsSDK.String(destKey),
			Body:        bytes.NewReader(newB),
			ACL:         awsSDK.String("bucket-owner-full-control"),
			ContentType: awsSDK.String("application/json"),
		}
		_, err = s3conn.PutObject(putInput)
		if err != nil {
			return fmt.Errorf("Migrating %s to %s failed: %s", *obj.Key, destKey, err)
		}

		fmt.Printf(boldGreen("Migration of %s to %s successful\n"), key, destKey)

		// Delete the original
		fmt.Printf("Now attempting to delete %s\n", key)

		deleteInput := &s3.DeleteObjectInput{
			Bucket: awsSDK.String(bucket),
			Key:    awsSDK.String(key),
		}
		_, err = s3conn.DeleteObject(deleteInput)
		if err != nil {
			return fmt.Errorf("Delete of %s failed: %s", key, err)
		}
		fmt.Printf(boldGreen("Key deleted successfully: %s\n"), key)
	}
	return nil
}

func migrateTfRemoteState() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		identity := c.App.Metadata["identity"].(*sts.GetCallerIdentityOutput)
		s3conn := c.App.Metadata["s3conn"].(*s3.S3)
		bucket := getStateBucketForEnv(c.String("env"))

		return migrateTfStateFiles(s3conn, bucket, func(sourceKey string) (bool, string) {
			re := regexp.MustCompile(*identity.Account + `\/` + c.String("app") + `\/versions\/[a-zA-Z0-9]{7}\.tfstate`)
			if !re.MatchString(sourceKey) {
				return false, ""
			}
			return true, strings.Replace(sourceKey, "/versions/", "/slots/", -1)
		}, func(source string) string {
			return strings.Replace(source, "/versions/", "/slots/", -1)
		}, c.Bool("v"))
	}
}

func rollbackMigratedTfRemoteState() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		identity := c.App.Metadata["identity"].(*sts.GetCallerIdentityOutput)
		s3conn := c.App.Metadata["s3conn"].(*s3.S3)
		bucket := getStateBucketForEnv(c.String("env"))

		return migrateTfStateFiles(s3conn, bucket, func(sourceKey string) (bool, string) {
			re := regexp.MustCompile(*identity.Account + `\/` + c.String("app") + `\/slots\/[a-zA-Z0-9]{7}\.tfstate`)
			if !re.MatchString(sourceKey) {
				return false, ""
			}
			return true, strings.Replace(sourceKey, "/slots/", "/versions/", -1)
		}, func(source string) string {
			return strings.Replace(source, "/slots/", "/versions/", -1)
		}, c.Bool("v"))
	}
}

func prepareDeploymentState() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		identity := c.App.Metadata["identity"].(*sts.GetCallerIdentityOutput)
		s3conn := c.App.Metadata["s3conn"].(*s3.S3)
		bucket := getStateBucketForEnv(c.String("env"))

		cfgPath, _ := homedir.Expand("~/.rt/deployment-state.hcl.tpl")
		cfg, _, err := hcl.LoadConfigFromPath(c.String("env"), *identity.Account, cfgPath)
		if err != nil {
			return err
		}

		ds, err := deploymentstate.New(cfg.DeploymentState)
		if err != nil {
			return err
		}

		objects, err := listObjectsInS3Bucket(s3conn, bucket)
		if err != nil {
			return err
		}

		infraStateRe := regexp.MustCompile(*identity.Account + `\/([^/]+)/terraform.tfstate`)
		versionStateRe := regexp.MustCompile(*identity.Account + `\/([^/]+)/versions/([a-fA-F0-9]+).tfstate`)
		for _, obj := range objects {
			key := *obj.Key
			if infraStateRe.MatchString(key) {
				matches := infraStateRe.FindStringSubmatch(key)
				appName := matches[1]
				if appName == "shared-services" {
					continue
				}
				if c.String("app") != "" && c.String("app") != appName {
					continue
				}

				out, err := s3conn.GetObject(&s3.GetObjectInput{
					Bucket: awsSDK.String(bucket),
					Key:    awsSDK.String(key),
				})
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(out.Body)
				if err != nil {
					return err
				}
				state := TfState{}
				err = json.Unmarshal(b, &state)
				if err != nil {
					return err
				}

				numOfResources := 0
				rootOutputs := make(map[string]string, 0)
				for _, m := range state.Modules {
					if len(m.Path) == 1 && m.Path[0] == "root" {
						rootOutputs = m.Outputs
					}
					numOfResources += len(m.Resources)
				}

				if numOfResources > 0 {
					newApp := &schema.ApplicationData{
						UseCentralGitRepo:    true,
						LastRtVersion:        "0.4.7",
						LastTerraformVersion: "0.6.16",
						IsActive:             true,
						LastInfraChangeTime:  *obj.LastModified,
						InfraOutputs:         rootOutputs,
					}

					fmt.Printf("Saving application %q (state %d)...", appName, state.Version)
					err = ds.SaveApplication(appName, newApp)
					if err != nil {
						return err
					}
					fmt.Println(" DONE")
					rootOutputs = map[string]string{}
				}
			}
			if versionStateRe.MatchString(key) {
				matches := versionStateRe.FindStringSubmatch(key)
				appName := matches[1]
				slotId := matches[2]

				if appName == "shared-services" {
					continue
				}
				if c.String("app") != "" && c.String("app") != appName {
					continue
				}

				out, err := s3conn.GetObject(&s3.GetObjectInput{
					Bucket: awsSDK.String(bucket),
					Key:    awsSDK.String(key),
				})
				if err != nil {
					return err
				}
				b, err := ioutil.ReadAll(out.Body)
				if err != nil {
					return err
				}
				state := TfState{}
				err = json.Unmarshal(b, &state)
				if err != nil {
					return err
				}

				numOfResources := 0
				rootOutputs := make(map[string]string, 0)
				for _, m := range state.Modules {
					if len(m.Path) == 1 && m.Path[0] == "root" {
						rootOutputs = m.Outputs
					}
					numOfResources += len(m.Resources)
				}

				if numOfResources > 0 {
					fmt.Printf("Saving slot for %q: %q (state %d)...", appName, slotId, state.Version)
					tfVariables := map[string]string{
						"app_name":    appName,
						"app_version": slotId,
						"environment": c.String("env"),
					}
					zeroTime, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
					if err != nil {
						return err
					}
					data, err := ds.BeginDeployment(appName, slotId, false, nil, zeroTime, tfVariables)
					if err != nil {
						return err
					}
					data.RTVersion = "0.4.7"
					err = ds.FinishDeployment(appName, slotId, data.DeploymentId, true, data, &schema.FinishedTerraformRun{
						FinishTime: *obj.LastModified,
						Outputs:    rootOutputs,
					})
					if err != nil {
						return err
					}
					fmt.Println(" DONE")
				}
				rootOutputs = map[string]string{}
			}
		}

		return nil
	}
}
