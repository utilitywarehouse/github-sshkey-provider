package gskp

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestUserInfo_Marshal(t *testing.T) {
	ui := []UserInfo{
		UserInfo{
			Login: "user0",
			ID:    999999,
			Name:  "User Zero",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user0",
		},
		UserInfo{
			Login: "user1",
			ID:    999998,
			Name:  "User One",
			Keys:  "ssh-dsa this_will_be_a_really_really_really_long_ssh_key_string_for_user0",
		},
	}

	jsonTextExpected := `[{"login":"user0","id":999999,"name":"User Zero","keys":"ssh-rsa this_will_be_a_really_` +
		`really_really_long_ssh_key_string_for_user0"},{"login":"user1","id":999998,` +
		`"name":"User One","keys":"ssh-dsa this_will_be_a_really_really_really_long_ssh_key_string_` +
		`for_user0"}]`

	jsonText, err := json.Marshal(ui)
	if err != nil {
		t.Fatalf("[]UserInfo marshal returned error: %v", err)
	}

	if string(jsonText) != jsonTextExpected {
		t.Errorf("[]UserInfo marshal returned unexpected results: %s", jsonText)
	}
}

func TestUserInfo_Unmarshal(t *testing.T) {
	uiExpected := []UserInfo{
		UserInfo{
			Login: "user0",
			ID:    999999,
			Name:  "User Zero",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user0",
		},
		UserInfo{
			Login: "user1",
			ID:    999998,
			Name:  "User One",
			Keys:  "ssh-dsa this_will_be_a_really_really_really_long_ssh_key_string_for_user0",
		},
	}

	jsonText := `[{"login":"user0","id":999999,"name":"User Zero","keys":"ssh-rsa this_will_be_a_really_` +
		`really_really_long_ssh_key_string_for_user0"},{"login":"user1","id":999998,` +
		`"name":"User One","keys":"ssh-dsa this_will_be_a_really_really_really_long_ssh_key_string_` +
		`for_user0"}]`

	ui := []UserInfo{}
	err := json.Unmarshal([]byte(jsonText), &ui)
	if err != nil {
		t.Fatalf("[]UserInfo unmarshal returned error: %v", err)
	}

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("[]UserInfo unmarshal returned unexpected results: %v", ui)
	}
}

func TestUserInfo_Unmarshal_error(t *testing.T) {
	jsonText := `["login":"user0","id":999999,"name":"User Zero","keys":""]`

	ui := []UserInfo{}
	err := json.Unmarshal([]byte(jsonText), &ui)
	if err == nil {
		t.Errorf("[]UserInfo unmarshal should have returned an error")
	}
}
