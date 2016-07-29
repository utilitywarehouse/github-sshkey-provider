package authorizedkeys

import (
	"testing"

	"github.com/utilitywarehouse/github-sshkey-provider/pkg/collector"
)

func TestGenerateAuthorizedKeysFile(t *testing.T) {
	ui := []collector.UserInfo{
		collector.UserInfo{
			Login: "user00",
			ID:    999998,
			Name:  "User Zero",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user00",
		},
		collector.UserInfo{
			Login: "user01",
			ID:    999999,
			Name:  "User One",
			Keys:  "ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user01",
		},
	}

	akExpected := `
# BEGIN: github_key_provider

# SSH keys for user00 (User Zero)
ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user00

# SSH keys for user01 (User One)
ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string_for_user01

# END: github_key_provider
`
	ak, err := GenerateSnippet(ui)

	if err != nil {
		t.Errorf("GenerateAuthorizedKeysFile returned error: %v", err)
	}

	if ak != akExpected {
		t.Errorf("GenerateAuthorizedKeysFile returned unexpected value: %v", ak)
	}
}
