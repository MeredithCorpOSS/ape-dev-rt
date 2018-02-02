package ipinfo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeMyIPResponse = `{
	"ip": "12.34.56.78",
	"hostname": "google-proxy-12-34-56-78.google.com",
	"city": null,
	"country": "EU",
	"loc": "12.0000,3.0000",
	"org": "AS15169 Google Inc."
}`

var fakeForeignIPResponse = `{
	"ip": "8.8.8.8",
	"hostname": "google-public-dns-a.google.com",
	"city": "Mountain View",
	"region": "California",
	"country": "US",
	"loc": "37.3860,-122.0838",
	"org": "AS15169 Google Inc.",
	"postal": "94040"
}`

func mockServer(mockResponse string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mockResponse)
	}))

	return server
}

func TestForeignIP(t *testing.T) {
	server := mockServer(fakeForeignIPResponse)
	defer server.Close()

	ipinfoURI = server.URL

	testIP := "8.8.8.8"
	info, err := ForeignIP(testIP)

	if err != nil {
		t.Errorf(`ForeignIP("%s") error %s`, testIP, err)
	}

	if info.IP != testIP {
		t.Errorf(`ForeignIP("%s") expected %s got %s`, testIP, testIP, info.IP)
	}
}

func TestMyIP(t *testing.T) {
	server := mockServer(fakeForeignIPResponse)
	defer server.Close()

	ipinfoURI = server.URL

	_, err := MyIP()

	if err != nil {
		t.Errorf(`MyIP() error %s`, err)
	}
}
