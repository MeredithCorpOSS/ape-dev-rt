package aws

import (
	"fmt"
	"os"
	"strings"
	"time"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/go-cleanhttp"
)

var DefaultRegion = "us-east-1"

type User struct {
	AccountID string
	Name      string
	UserID    string
	Arn       string
}

type AWS struct {
	Region *string

	autoscalingConn *autoscaling.AutoScaling
	ec2Conn         *ec2.EC2
	elbConn         *elb.ELB
	s3Conn          *s3.S3
	stsConn         *sts.STS
}

func NewAWS(profileName string, region string) *AWS {
	if profileName == "" {
		profileName = "default"
	}
	if region == "" {
		region = DefaultRegion
	}
	creds := CredentialsProvider(profileName)
	os.Setenv("AWS_PROFILE", profileName) // Terraform uses aws-sdk-go
	config := &awsSDK.Config{
		Credentials: creds,
		Region:      awsSDK.String(region),
	}

	sess := session.New(config)

	return &AWS{
		Region: awsSDK.String(region),

		autoscalingConn: autoscaling.New(sess),
		ec2Conn:         ec2.New(sess),
		elbConn:         elb.New(sess),
		stsConn:         sts.New(sess),
		s3Conn:          s3.New(sess),
	}
}

func (a *AWS) User() (*User, error) {
	svc := a.stsConn

	resp, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("Failed to verify AWS identity via STS: %s", err)
	}

	return &User{*resp.Account, parseUsernameFromARN(*resp.Arn), *resp.UserId, *resp.Arn}, nil
}

func CredentialsProvider(awsCredentialProfile string) *credentials.Credentials {
	provider := []credentials.Provider{
		&credentials.SharedCredentialsProvider{
			Filename: "",
			Profile:  awsCredentialProfile,
		},
		&credentials.EnvProvider{},
	}

	// Check if we're in AWS, if we are then we
	// want to try and use instance profiles
	client := cleanhttp.DefaultClient()
	client.Timeout = 100 * time.Millisecond
	cfg := &awsSDK.Config{
		HTTPClient: client,
	}

	metadataClient := ec2metadata.New(session.New(cfg))
	if metadataClient.Available() {
		provider = append(provider, &ec2rolecreds.EC2RoleProvider{
			Client: metadataClient,
		})
	}

	return credentials.NewChainCredentials(provider)
}

func parseAccountIdFromARN(arn string) string {
	arnParts := strings.Split(arn, ":")
	return arnParts[4]
}

func parseUsernameFromARN(arn string) string {
	arnParts := strings.Split(arn, ":")
	userNameParts := strings.Split(arnParts[len(arnParts)-1], "/")
	return userNameParts[1]
}
