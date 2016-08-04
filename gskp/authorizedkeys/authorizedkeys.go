package authorizedkeys

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/collector"
)

const (
	snippetBeginSeparator = `# BEGIN: github_sshkey_provider`
	snippetEndSeparator   = `# END: github_sshkey_provider`
	snippetTemplate       = `
{{ range $index, $user := . -}}
# SSH keys for {{ $user.Login }} (
{{- if $user.Name -}}
    {{ $user.Name }}
{{- else -}}
    unknown name
{{- end }})
{{ $user.Keys }}
{{ end -}}`
)

// GenerateSnippet returns a string containing an OpenSSH-compatible
// authorized_keys file snippet, based on a list of UserInfo structs
func GenerateSnippet(ui collector.UserInfoList) (string, error) {
	t := template.New("authorized_keys")
	t, err := t.Parse(snippetTemplate)
	if err != nil {
		return "", nil
	}

	var output bytes.Buffer
	if err := t.Execute(&output, ui); err != nil {
		return "", nil
	}

	return strings.Join([]string{snippetBeginSeparator, output.String(), snippetEndSeparator}, "\n"), nil
}
