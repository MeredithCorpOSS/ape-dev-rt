package aws

import (
	"testing"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
)

func TestParseUsernameFromARN(t *testing.T) {
	sampleARNs := map[string]string{
		"arn:aws:iam::9991112224:user/radek.simko": "radek.simko",
		"arn:aws:iam::0123456789:user/ian":         "ian",
		"arn:aws:iam::0123456789:user/rsimko1016":  "rsimko1016",
	}

	for arn, expectedUsername := range sampleARNs {
		username := parseUsernameFromARN(arn)
		if username != expectedUsername {
			t.Errorf("Parsed username (%s) does not match expected username (%s)",
				username, expectedUsername)
		}
	}
}

func TestUser(t *testing.T) {
	routes := []*MockRoute{
		&MockRoute{
			ExpectedURI:         "/",
			ExpectedRequestBody: "Action=GetCallerIdentity&Version=2011-06-15",
			Response: MockResponse{
				Code: 200,
				Body: test_sts_getCallerIdentityAwsApiBody,
			},
		},
	}
	mockedSession, closeFunc := GetMockedAwsSession(routes, "us-east-1")
	defer closeFunc()

	a := &AWS{
		Region:  awsSDK.String("us-east-1"),
		stsConn: sts.New(mockedSession),
	}

	u, err := a.User()
	if err != nil {
		t.Fatalf("Failed getting current user: %s", err)
	}
	expectedAccountId := "123456789012"
	if u.AccountID != expectedAccountId {
		t.Fatalf("Expected AccountID (%s) doesn't match: %s", expectedAccountId, u.AccountID)
	}
	expectedUserName := "Bob"
	if u.Name != expectedUserName {
		t.Fatalf("Expected Name (%s) doesn't match: %s", expectedUserName, u.Name)
	}
}

var test_sts_getCallerIdentityAwsApiBody = `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
   <Arn>arn:aws:iam::123456789012:user/Bob</Arn>
    <UserId>AKIAI44QH8DHBEXAMPLE</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`
