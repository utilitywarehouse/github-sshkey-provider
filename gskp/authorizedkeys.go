package gskp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
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

// ReadAndStripFile will read an authorized_keys file and strip any portions
// between the separators.
func (authorizedKeys) ReadAndStripFile(filename string) (string, error) {
	input, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer input.Close()

	return AuthorizedKeys.stripFile(input)
}

func (authorizedKeys) stripFile(file io.Reader) (string, error) {
	ret := []string{}

	scanner := bufio.NewScanner(file)

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
