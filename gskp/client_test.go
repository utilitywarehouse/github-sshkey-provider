package gskp

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

var (
	testClient *Client
)

func init() {
	simplelog.DebugEnabled = true

	testClient, _ = NewClient("http://localhost:35432", 1)
}

func TestClient_GetKeys(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	h := startNewTestServer()
	defer h.Stop(time.Second)

	var dataExpected []UserInfo
	json.Unmarshal([]byte(`[{"login":"user","id":999999,"name":"User Name","keys":"ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string"}]`), &dataExpected)

	data, err := testClient.GetKeys("Owners")
	if err != nil {
		t.Fatalf("Client.GetKeys returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(data, dataExpected) {
		t.Errorf("Client.GetKeys returned unexpected value: %v", data)
	}
}

func TestClient_GetKeys_error(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	h := startNewTestServer()
	defer h.Stop(time.Second)

	_, err := testClient.GetKeys("invalid")
	if err != ErrClientUnexpected {
		t.Fatalf("Client.GetKeys returned unexpected error: %v", err)
	}
}

func TestClient_PollForKeys_timeout(t *testing.T) {
	h := startNewTestServer()
	defer h.Stop(time.Second)

	_, err := testClient.PollForKeys("Owners")
	if err != ErrClientPollTimeout {
		t.Fatalf("Client.GetKeys returned unexpected error, was expecting timeout: %v", err)
	}
}
