package gskp

import (
	"testing"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestAuthorizedKeys_GenerateAuthorizedKeysFile(t *testing.T) {
	ui := UserInfoList{
		UserInfo{
			Login: "user00",
			ID:    999998,
			Name:  "User Zero",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user00",
		},
		UserInfo{
			Login: "user01",
			ID:    999999,
			Name:  "User One",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user01",
		},
	}

	akExpected := `# BEGIN: github_sshkey_provider

# SSH keys for user00 (User Zero)
ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user00
# SSH keys for user01 (User One)
ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user01

# END: github_sshkey_provider`

	ak, err := AuthorizedKeys.GenerateSnippet(ui)

	if err != nil {
		t.Fatalf("GenerateAuthorizedKeysFile returned error: %v", err)
	}

	if ak != akExpected {
		t.Errorf("GenerateAuthorizedKeysFile returned unexpected value: %v", ak)
	}
}

var stripTestsWithoutErrors = []struct {
	Input    string
	Expected string
}{
	{ // test #0
		`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
sample snippet line 01
# END: github_sshkey_provider
sample line 01`,
		`this_is_a_test_authorized_keys_file
sample line 00
sample line 01`,
	},
	{ // test #1
		`this_is_a_test_authorized_keys_file
sample line 00
sample line 01`,
		`this_is_a_test_authorized_keys_file
sample line 00
sample line 01`,
	},
	{ // test #2
		`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
sample snippet line 01
# END: github_sshkey_provider`,
		`this_is_a_test_authorized_keys_file
sample line 00`,
	},
	{ // test #3
		`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
sample snippet line 01`,
		`this_is_a_test_authorized_keys_file
sample line 00`,
	},
	{ // test #4
		`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
sample snippet line 01
# END: github_sshkey_provider
sample line 01
`,
		`this_is_a_test_authorized_keys_file
sample line 00
sample line 01`,
	},
	{ // test #5
		`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
sample snippet line 01
# END: github_sshkey_provider
sample line 01

`,
		`this_is_a_test_authorized_keys_file
sample line 00
sample line 01`,
	},
}

var stripTestsWithErrors = []string{
	// test #0
	`this_is_a_test_authorized_keys_file
sample line 00
sample snippet line 00
sample snippet line 01
# END: github_sshkey_provider`,
	// test #1
	`this_is_a_test_authorized_keys_file
sample line 00
# BEGIN: github_sshkey_provider
sample snippet line 00
# BEGIN: github_sshkey_provider
sample snippet line 01
# END: github_sshkey_provider`,
}

func TestAuthorizedKeys_stripFile(t *testing.T) {
	for i, test := range stripTestsWithoutErrors {
		out, err := AuthorizedKeys.stripFile(test.Input)
		if err != nil {
			t.Fatalf("AuthorizedKeys.stripFile returned an error for test #%d: %v", i, err)
		}

		if out != test.Expected {
			t.Errorf("AuthorizedKeys.stripFile returned unexpected output for test #%d: %s", i, out)
		}
	}

	for i, testInput := range stripTestsWithErrors {
		_, err := AuthorizedKeys.stripFile(testInput)
		if err == nil {
			t.Errorf("AuthorizedKeys.stripFile should have returned an error for test #%d", i)
		}
	}
}
