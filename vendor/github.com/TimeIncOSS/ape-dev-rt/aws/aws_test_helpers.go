package aws

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

type MockedAWSInput struct {
	Region          string
	AutoscalingSess *session.Session
	Ec2Sess         *session.Session
	ElbSess         *session.Session
	StsSess         *session.Session
	S3Sess          *session.Session
}

func MockedAWS(input *MockedAWSInput) *AWS {
	a := &AWS{
		Region: awsSDK.String(input.Region),
	}
	if input.AutoscalingSess != nil {
		a.autoscalingConn = autoscaling.New(input.AutoscalingSess)
	}
	if input.Ec2Sess != nil {
		a.ec2Conn = ec2.New(input.Ec2Sess)
	}
	if input.ElbSess != nil {
		a.elbConn = elb.New(input.ElbSess)
	}
	if input.StsSess != nil {
		a.stsConn = sts.New(input.StsSess)
	}
	if input.S3Sess != nil {
		a.s3Conn = s3.New(input.S3Sess)
	}

	return a
}

func GetMockedAwsSession(apiRoutes []*MockRoute, region string) (*session.Session, func()) {
	url, closeFunc := mockServer(apiRoutes)

	sess := session.New(&awsSDK.Config{
		Region:      awsSDK.String(region),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
	})

	return sess.Copy(&awsSDK.Config{
		Endpoint:         awsSDK.String(url),
		S3ForcePathStyle: awsSDK.Bool(true), // This makes mocking easier
	}), closeFunc
}

func mockServer(routes []*MockRoute) (string, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Some parts of AWS SDK (S3) make RequestURI full URL (strangely)
		uri := strings.TrimPrefix(r.RequestURI, "http://"+r.Host)

		log.Printf("[DEBUG] Mocked server received request to %q", uri)

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		reqBody := buf.String()

		log.Printf("[DEBUG] Mocked server received body: %q", reqBody)

		mr, err := lookupMockRoute(routes, uri, reqBody, r)
		if err != nil {
			w.WriteHeader(400)
			log.Printf("[DEBUG] Responding HTTP 400: %s, known: %#v", err, routes)
			fmt.Fprintln(w, "<ErrorResponse><Error><Code>UnknownRequest</Code></Error></ErrorResponse>")
			return
		}

		resp := mr.Response
		w.WriteHeader(resp.Code)

		for k, v := range resp.HeaderMap {
			w.Header().Set(k, v)
		}
		if resp.Body != "" {
			fmt.Fprintln(w, resp.Body)
		}
	}))

	log.Printf("[DEBUG] Created new mock server: %s", ts.URL)

	return ts.URL, ts.Close
}

func lookupMockRoute(routes []*MockRoute, uri, body string, req *http.Request) (*MockRoute, error) {
	for _, route := range routes {
		r := *route
		if r.ExpectedURI == uri && r.ExpectedRequestBody == body &&
			(r.ExpectedMethod == "" || r.ExpectedMethod == req.Method) {
			log.Printf("[DEBUG] Mocked server matched...")
			return route, nil
		}
	}
	return nil, fmt.Errorf("Mock route not found")

}

type MockRoute struct {
	ExpectedURI         string
	ExpectedMethod      string
	ExpectedRequestBody string
	Response            MockResponse
}

type MockResponse struct {
	Code      int
	HeaderMap map[string]string
	Body      string
}
