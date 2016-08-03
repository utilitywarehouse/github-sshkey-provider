package collector

import (
	"reflect"
	"testing"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestUserInfoList_Marshal(t *testing.T) {
	ui := UserInfoList{
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

	jsonText, err := ui.Marshal()
	if err != nil {
		t.Fatalf("UserInfoList.Marshal returned error: %v", err)
	}

	if jsonText != jsonTextExpected {
		t.Errorf("UserInfoList.Marshal returned unexpected results: %s", jsonText)
	}
}

func TestUserInfoList_Unmarshal(t *testing.T) {
	uiExpected := UserInfoList{
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

	ui := UserInfoList{}
	err := ui.Unmarshal(jsonText)
	if err != nil {
		t.Fatalf("UserInfoList.Unmarshal returned error: %v", err)
	}

	if !reflect.DeepEqual(ui, uiExpected) {
		t.Errorf("UserInfoList.Unmarshal returned unexpected results: %v", ui)
	}
}

func TestUserInfoList_Unmarshal_error(t *testing.T) {
	jsonText := `["login":"user0","id":999999,"name":"User Zero","keys":""]`

	ui := UserInfoList{}
	err := ui.Unmarshal(jsonText)
	if err == nil {
		t.Errorf("UserInfoList.Unmarshal should have returned an error")
	}
}
