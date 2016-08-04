package gskp

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

var (
	// AuthorizedKeys provides various functions related to the manipulation
	// of OpenSSH-compatible authorized_keys files.
	AuthorizedKeys authorizedKeys
)

type authorizedKeys struct{}

// GenerateSnippet returns a string containing an snippet compatible with
// OpenSSH authorized_keys format, based on a list of UserInfo structs.
func (authorizedKeys) GenerateSnippet(ui collector.UserInfoList) (string, error) {
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
