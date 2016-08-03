package collector

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

	// testKeyCollector is the KeyCollector being tested
	testKeyCollector *KeyCollector

	// testGithubClient is the GitHub client being tested (used in testKeyCollector)
	testGithubClient *github.Client

	// testHTTPClient is the HTTP client being tested (used in testKeyCollector)
	testHTTPClient *http.Client
)

// mockSetup sets up the test HTTP server along with a test KeyCollector instance
// configured to talk to that test server
func mockSetup() {
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testGithubClient = github.NewClient(nil)
	u, _ := url.Parse(testServer.URL)
	testGithubClient.BaseURL = u
	testGithubClient.UploadURL = u

	testKeyCollector = &KeyCollector{
		githubClient:  testGithubClient,
		httpClient:    &http.Client{},
		githubKeysURL: fmt.Sprintf("%s/%%s.keys", u),
	}
}

// mockTeardown closes the test HTTP server
func mockTeardown() {
	testServer.Close()
}

func mockInstallHandlers(handlers []string) {
	for _, h := range handlers {
		switch h {
		case "orgTeams":
			testMux.HandleFunc("/orgs/none/teams", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{"name": "Owners", "id": 888888}]`)
			})
		case "userKeys":
			testMux.HandleFunc("/user.keys", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string`)
			})
		case "userInfo":
			testMux.HandleFunc("/user/999999", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"id": 999999, "name": "User Name"}`)
			})
		case "teamUserList":
			// cannot be specific on the path for OrganizationListTeamMembersOptions
			// because it contains URL parameters which the mux won't handle properly
			// eg. "teams/888888/members?per_page=100&role=all"
			// all the other endpoints have been configured previously and so this
			// generic catch-all handler will only ever respond to the desired call
			testMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{"login": "user", "id": 999999}]`)
			})
		}
	}
}

func TestKeyCollector_GetTeamMemberInfo(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	miExpected := UserInfoList{
		UserInfo{
			Login: "user",
			ID:    999999,
			Name:  "User Name",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string",
		},
	}

	mi, err := testKeyCollector.GetTeamMemberInfo("none", "Owners")
	if err != nil {
		t.Fatalf("KeyCollector.GetTeamMemberInfo returned an error: %v", err)
	}

	if !reflect.DeepEqual(mi, miExpected) {
		t.Errorf("KeyCollector.GetTeamMemberInfo returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_GetTeamMemberInfo_getTeamIDError(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	mi, err := testKeyCollector.GetTeamMemberInfo("none", "Owners")
	if err == nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo should have returned an error")
	}

	if mi != nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_GetTeamMemberInfo_getTeamIDUnknown(t *testing.T) {
	mockSetup()
	testMux.HandleFunc("/orgs/none/teams", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{
                "name": "Owners",
                "id": 888888
            }]`)
	})
	defer mockTeardown()

	mi, err := testKeyCollector.GetTeamMemberInfo("none", "UnkownTeam")
	if err == nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo should have returned an error")
	}

	if mi != nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_GetTeamMemberInfo_getTeamMembersError(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams"})
	defer mockTeardown()

	mi, err := testKeyCollector.GetTeamMemberInfo("none", "Owners")
	if err == nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo should have returned an error")
	}

	if mi != nil {
		t.Errorf("KeyCollector.GetTeamMemberInfo returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_getUserKeys_error(t *testing.T) {
	mi, err := testKeyCollector.getUserKeys("")
	if err == nil {
		t.Errorf("KeyCollector.getUserKeys should have returned an error")
	}

	if mi != "" {
		t.Errorf("KeyCollector.getUserKeys returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_getUserKeys_notFound(t *testing.T) {
	mockSetup()
	defer mockTeardown()

	testMux.HandleFunc("/user.keys", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `Not Found`)
	})

	mi, err := testKeyCollector.getUserKeys("user")
	if err == nil {
		t.Errorf("KeyCollector.getUserKeys should have returned an error")
	}

	if err != ErrGithubKeysNotFound {
		t.Errorf("KeyCollector.getUserKeys returned an unexpected error: %v", err)
	}

	if mi != "" {
		t.Errorf("KeyCollector.getUserKeys returned unexpected value: %v", mi)
	}
}

func TestKeyCollector_setUserName_error(t *testing.T) {
	ui := UserInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	uiExpected := UserInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	testKeyCollector.setUserName(&ui)

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("KeyCollector.setUserName returned unexpected value: %v", ui)
	}
}

func TestKeyCollector_setUserKeys_error(t *testing.T) {
	ui := UserInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	uiExpected := UserInfo{
		Login: "user",
		ID:    999999,
		Name:  "unknown name",
		Keys:  "",
	}

	testKeyCollector.setUserKeys(&ui)

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("KeyCollector.setUserKeys returned unexpected value: %v", ui)
	}
}
