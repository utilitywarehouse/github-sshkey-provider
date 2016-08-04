package gskp

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"text/template"
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

	// ErrAuthorizedKeysFileMalformed is returned when the authorized_keys file
	// that is being read is found to be malformed.
	ErrAuthorizedKeysFileMalformed = errors.New("The authorized_keys file is malformed")
)

type authorizedKeys struct{}

// GenerateSnippet returns a string containing an snippet compatible with
// OpenSSH authorized_keys format, based on a list of UserInfo structs.
func (authorizedKeys) GenerateSnippet(ui UserInfoList) (string, error) {
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

// Update will read an authorized_keys, strip any portions managed by this
// service (identified by the separators) and append the provided snippet
// at the end.
func (authorizedKeys) Update(filename string, snippet string) error {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	strippedContents, err := AuthorizedKeys.stripFile(string(fileContents))
	if err != nil {
		return err
	}

	output := []byte(strings.Join([]string{strippedContents, snippet}, "\n\n"))

	return ioutil.WriteFile(filename, output, 0600)
}

func (authorizedKeys) stripFile(fileContents string) (string, error) {
	ret := []string{}

	scanner := bufio.NewScanner(strings.NewReader(fileContents))

	readingSnippetLines := false
	for scanner.Scan() {
		line := scanner.Text()

		if line == snippetBeginSeparator {
			if readingSnippetLines {
				return "", ErrAuthorizedKeysFileMalformed
			}

			readingSnippetLines = true
			continue
		} else if line == snippetEndSeparator {
			if !readingSnippetLines {
				return "", ErrAuthorizedKeysFileMalformed
			}

			readingSnippetLines = false
			continue
		}

		if !readingSnippetLines {
			ret = append(ret, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(ret, "\n"), nil
}
