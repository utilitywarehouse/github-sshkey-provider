package gskp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func startNewTestHTTPServer() *HTTPServer {
	h := NewHTTPServer()

	h.HandleGet("/long_operation", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprintf(w, "this was a long operation")
	})

	go h.Listen(":35432", 10)
	time.Sleep(100 * time.Millisecond)

	return h
}

func testGetResponse(t *testing.T, endpoint string, expectedResponse string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:35432/%s", endpoint))
	if err != nil {
		t.Fatalf("Error when trying to GET the %s endpoint: %v", endpoint, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error when reading the response from the %s endpoint: %v", endpoint, err)
	}
	resp.Body.Close()

	if string(body) != expectedResponse {
		t.Errorf("Unexpected response from the %s endpoint: %s", endpoint, body)
	}
}

func test405Response(t *testing.T, method string, endpoint string, expected string) {
	req, err := http.NewRequest(method, "http://localhost:35432/status", nil)
	if err != nil {
		t.Fatalf("Could not construct a proper %s request", method)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error when trying a %s on the status endpoint: %v", method, err)
	}

	if resp.StatusCode != 405 {
		t.Errorf("Unexpected status code when trying %s on the status endpoint: %d", method, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error when reading the response from a %s on the status endpoint: %v", method, err)
	}
	resp.Body.Close()

	if string(body) != expected {
		t.Errorf("Unexpected response when trying %s on the status endpoint: %s", method, string(body))
	}
}

var testEndpointsMap = map[string]string{
	"status":         `{"status":"ok"}`,
	"long_operation": `this was a long operation`,
}

func TestHTTPServer_endpoints(t *testing.T) {
	h := startNewTestHTTPServer()
	defer h.StopListening(10)

	for endpoint, expected := range testEndpointsMap {
		testGetResponse(t, endpoint, expected)
	}
}

var testUnsupportedMethodsList = map[string]string{
	"POST":    `{"error":"method not allowed"}`,
	"PUT":     `{"error":"method not allowed"}`,
	"DELETE":  `{"error":"method not allowed"}`,
	"HEAD":    "",
	"OPTIONS": `{"error":"method not allowed"}`,
	"CONNECT": `{"error":"method not allowed"}`,
	"TRACE":   `{"error":"method not allowed"}`,
}

func TestHTTPServer_unsupportedMethod(t *testing.T) {
	h := startNewTestHTTPServer()
	defer h.StopListening(10)

	for endpoint := range testEndpointsMap {
		for method, expected := range testUnsupportedMethodsList {
			fmt.Println(method, endpoint, expected)
			test405Response(t, method, endpoint, expected)
		}
	}
}

func TestHTTPServer_connectionDrain(t *testing.T) {
	wg := &sync.WaitGroup{}

	h := startNewTestHTTPServer()
	time.Sleep(100 * time.Millisecond)

	wg.Add(1)
	var body []byte
	go func() {
		defer wg.Done()

		resp, err := http.Get("http://localhost:35432/long_operation")
		if err != nil {
			t.Fatalf("Error when trying to GET the status page: %v", err)
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Error when reading the response from the status page: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	h.StopListening(10)

	wg.Wait()

	if string(body) != `this was a long operation` {
		t.Errorf("Status page contained unexpected response: %s", string(body))
	}
}
