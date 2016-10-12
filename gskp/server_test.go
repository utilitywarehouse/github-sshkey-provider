package gskp

import (
	"bytes"
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

func startNewTestServer() *Server {
	testKeyCache = NewKeyCache("none", "", 5*time.Second)
	testKeyCache.collector = testKeyCollector

	h, _ := NewServer(testKeyCache)

	h.mux.HandleFunc("/long_operation", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		fmt.Fprintf(w, "this was a long operation")
	})

	go h.Start(":35432", 10)
	time.Sleep(100 * time.Millisecond)

	return h
}

func testGetResponse(t *testing.T, endpoint string, expectedResponse string) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:35432/%s", endpoint))
	if err != nil {
		t.Fatalf("Error when trying to GET the %s endpoint: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error when reading the response from the %s endpoint: %v", endpoint, err)
	}

	if !bytes.Equal(body, []byte(expectedResponse)) {
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
	"status":         `{"git_sha":"","image":"","status":"ok"}`,
	"long_operation": `this was a long operation`,
}

func TestServer_endpoints(t *testing.T) {
	h := startNewTestServer()
	defer h.Stop(10 * time.Second)

	for endpoint, expected := range testEndpointsMap {
		testGetResponse(t, endpoint, expected)
	}
}

var testUnsupportedMethodsList = map[string]string{
	"POST":    `{"error":"invalid method"}`,
	"PUT":     `{"error":"invalid method"}`,
	"DELETE":  `{"error":"invalid method"}`,
	"HEAD":    "",
	"OPTIONS": `{"error":"invalid method"}`,
	"CONNECT": `{"error":"invalid method"}`,
	"TRACE":   `{"error":"invalid method"}`,
}

func TestServer_unsupportedMethod(t *testing.T) {
	h := startNewTestServer()
	defer h.Stop(10 * time.Second)

	for endpoint := range testEndpointsMap {
		for method, expected := range testUnsupportedMethodsList {
			test405Response(t, method, endpoint, expected)
		}
	}
}

func TestServer_connectionDrain(t *testing.T) {
	h := startNewTestServer()
	time.Sleep(100 * time.Millisecond)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		testGetResponse(t, "long_operation", "this was a long operation")
	}()
	time.Sleep(100 * time.Millisecond)

	h.Stop(10 * time.Second)

	wg.Wait()
}

func TestServer_keys(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	h := startNewTestServer()
	defer h.Stop(time.Second)

	dataExpected := `{"keys":[{"login":"user","id":999999,"name":"User Name","keys":"ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string"}]}`

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// this go func will test longpolling
	go func() {
		defer wg.Done()
		testGetResponse(t, "keys?team=Owners", dataExpected)
	}()

	// this is to test the init functionality of the endpoint
	// it will also unblock the request in the go func above
	testGetResponse(t, "keys?init=true&team=Owners", dataExpected)

	wg.Wait()
}

func TestServer_keys_erros(t *testing.T) {
	h := startNewTestServer()
	defer h.Stop(time.Second)

	testGetResponse(t, "keys?init=true", `{"error":"invalid team value"}`)
	testGetResponse(t, "keys?init=0&team=none", `{"error":"invalid init value"}`)
}
