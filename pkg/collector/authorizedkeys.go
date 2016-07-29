package collector

import (
	"bytes"
	"text/template"
)

const (
	authorizedKeysTemplate = `
# BEGIN: github_key_provider

{{ range $index, $user := . -}}
# SSH keys for {{ $user.Login }} (
{{- if $user.Name -}}
    {{ $user.Name }}
{{- else -}}
    unknown name
{{- end }})
{{ $user.Keys }}

{{ end -}}

# END: github_key_provider
`
)

// GenerateAuthorizedKeysFile returns a string containing an OpenSSH-compatible
// authorized_keys file snippet, based on a list of UserInfo structs
func GenerateAuthorizedKeysFile(ui []UserInfo) (string, error) {
	t := template.New("authorized_keys")
	t, err := t.Parse(authorizedKeysTemplate)
	if err != nil {
		return "", nil
	}

	var output bytes.Buffer
	t.Execute(&output, ui)

	return output.String(), nil
}
