package arukas

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	}))
	defer server.Close()
	os.Setenv("ARUKAS_JSON_API_URL", server.URL)

	client := NewClientWithOsExitOnErr()
	client.Timeout = 500 * time.Millisecond
	err := client.Get(nil, "/")
	if err == nil || !strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
		t.Error("Client doesn't timeout properly")
	}
}
