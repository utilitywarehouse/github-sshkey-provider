package keycollector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

var (
	// testServer is a test HTTP server used to provide mock API responses
	testServer *httptest.Server

	// testMux is the HTTP request multiplexer used with the test server
	testMux *http.ServeMux

	// testGithubClient is the GitHub client being tested
	testGithubClient *github.Client

	// testHTTPClient is the HTTP client being tested
	testHTTPClient *http.Client

	// testKeyCollector is the KeyCollector being tested
	testKeyCollector *KeyCollector
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func mockSetup() {
	// test server
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	// github client configured to use test server
	testGithubClient = github.NewClient(nil)
	u, _ := url.Parse(testServer.URL)
	testGithubClient.BaseURL = u
	testGithubClient.UploadURL = u

	// keycollector client configured to use test server
	testKeyCollector = &KeyCollector{
		githubClient:  testGithubClient,
		httpClient:    &http.Client{},
		githubKeysURL: fmt.Sprintf("%s/%%s.keys", u),
	}
}

// teardown closes the test HTTP server.
func mockTeardown() {
	testServer.Close()
}

func TestKeyCollector_getTeamID(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	testMux.HandleFunc("/orgs/none/teams", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{
            "name": "Owners",
            "id": 888888
        }]`)
	})

	id, err := testKeyCollector.getTeamID("none", "Owners")
	if err != nil {
		t.Errorf("KeyCollector.getTeamID returned error: %v", err)
	}

	if id != 888888 {
		t.Errorf("KeyCollector.getTeamID returned ID %d, wanted 888888", id)
	}
}

func TestKeyCollector_getUserKeys(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	testMux.HandleFunc("/user.keys", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string`)
	})

	keys, err := testKeyCollector.getUserKeys("user")
	if err != nil {
		t.Errorf("KeyCollector.getUserKeys returned error: %v", err)
	}

	if keys != "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string" {
		t.Errorf("KeyCollector.getUserKeys returned unexpected value: %s", keys)
	}
}

func TestKeyCollector_setUserName(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	testMux.HandleFunc("/user/999999", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
            "id": 999999,
            "name": "User Name"
        }`)
	})

	ui := userInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	uiExpected := userInfo{
		Login: "user",
		ID:    999999,
		Name:  "User Name",
		Keys:  "",
	}

	testKeyCollector.setUserName(&ui)

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("KeyCollector.setUserName returned unexpected value: %v", ui)
	}
}

func TestKeyCollector_setUserKeys(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	testMux.HandleFunc("/user.keys", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string`)
	})

	ui := userInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	uiExpected := userInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string",
	}

	testKeyCollector.setUserKeys(&ui)

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("KeyCollector.setUserKeys returned unexpected value: %v", ui)
	}
}

func TestKeyCollector_getTeamMembersInfo(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	// cannot be very specific on the path here because it contains parameters,
	// eg. "teams/888888/members?per_page=100&role=all" and so the mux will
	// have to handle everything ("/")
	testMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{
            "login": "user",
            "id": 999999
        }]`)
	})

	miExpected := []userInfo{
		userInfo{
			Login: "user",
			ID:    999999,
			Name:  "unknown name",
			Keys:  "",
		},
	}

	mi, err := testKeyCollector.getTeamMembersInfo(888888)
	if err != nil {
		t.Errorf("KeyCollector.getTeamMembersInfo returned error: %v", err)
	}

	if !reflect.DeepEqual(mi, miExpected) {
		t.Errorf("KeyCollector.getTeamMembersInfo returned unexpected value: %v", mi)
	}
}
