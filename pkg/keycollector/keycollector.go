package keycollector

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/simplelog"
	"golang.org/x/oauth2"
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

type userInfo struct {
	Login string
	ID    int
	Name  string
	Keys  string
}

// KeyCollector fetches public SSH keys from Github and generates an OpenSSH
// compatible authorized_keys snippet. The keys are selected based on Team
// membership.
type KeyCollector struct {
	githubClient  *github.Client
	httpClient    *http.Client
	githubKeysURL string
}

// NewKeyCollector returns an instantiated KeyCollector
func NewKeyCollector(githubAccessToken string) *KeyCollector {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return &KeyCollector{
		githubClient:  github.NewClient(tc),
		httpClient:    &http.Client{},
		githubKeysURL: "https://github.com/%s.keys",
	}
}

// GetTeamMemberAuthorizedKeys returns a snippet of SSH keys that can be
// inserted in an authorized_keys file in order to provide SSH access to
// members of the specified Team of a GitHub organization.
func (k *KeyCollector) GetTeamMemberAuthorizedKeys(organizationName string, teamName string) (string, error) {
	results, err := k.getTeamMembers(organizationName, teamName)
	if err != nil {
		return "", err
	}

	t := template.New("authorized_keys")
	t, err = t.Parse(authorizedKeysTemplate)
	if err != nil {
		return "", nil
	}

	var output bytes.Buffer
	t.Execute(&output, results)

	return output.String(), nil
}

func (k *KeyCollector) getTeamMembers(organizationName string, teamName string) ([]userInfo, error) {
	teamID, err := k.getTeamID(organizationName, teamName)
	if err != nil {
		return nil, err
	}

	memberInfo, err := k.getTeamMembersInfo(teamID)
	if err != nil {
		return nil, err
	}

	for _, mi := range memberInfo {
		k.setUserName(&mi)
		k.setUserKeys(&mi)
	}

	return memberInfo, nil
}

func (k *KeyCollector) getTeamMembersInfo(teamID int) ([]userInfo, error) {
	memberInfo := []userInfo{}

	ltmOpts := &github.OrganizationListTeamMembersOptions{
		Role: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	for {
		teamMembers, resp, err := k.githubClient.Organizations.ListTeamMembers(teamID, ltmOpts)
		if err != nil {
			return nil, err
		}

		for _, tm := range teamMembers {
			ui := userInfo{
				Login: *tm.Login,
				ID:    *tm.ID,
				Name:  "unknown name",
				Keys:  "",
			}

			memberInfo = append(memberInfo, ui)
		}

		if resp.NextPage == 0 {
			break
		}
	}

	return memberInfo, nil
}

func (k *KeyCollector) setUserName(ui *userInfo) {
	user, _, err := k.githubClient.Users.GetByID(ui.ID)
	if err != nil {
		simplelog.Info("Could not fetch details for user '%s': %v", ui.Login, err)
	} else {
		ui.Name = *user.Name
	}
}

func (k *KeyCollector) setUserKeys(ui *userInfo) {
	keys, err := k.getUserKeys(ui.Login)
	if err != nil {
		simplelog.Info("Could not fetch keys for user '%s': %s", ui.Login, err.Error())
	} else {
		ui.Keys = keys
	}
}

func (k *KeyCollector) getUserKeys(userLogin string) (string, error) {
	// Instead of using github.Users.ListKeys() which calls the GitHub API and is
	// a throttled request, we simply fetch them from the public URL that is
	// provided by GitHub.
	simplelog.Info("Fetching keys for user '%s'", userLogin)

	response, err := k.httpClient.Get(fmt.Sprintf(k.githubKeysURL, userLogin))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	keys := strings.TrimSpace(string(body))
	if keys == "Not Found" {
		return "", errors.New("Response was 'Not Found'")
	}

	return keys, nil
}

func (k *KeyCollector) getTeamID(organizationName string, teamName string) (int, error) {
	simplelog.Debug("Fetching list of teams for organization '%s'", organizationName)

	orgTeams, _, err := k.githubClient.Organizations.ListTeams(organizationName, nil)
	if err != nil {
		return -1, err
	}

	for _, team := range orgTeams {
		if *team.Name == teamName {
			simplelog.Debug("Team '%s' with id %d found in organization '%s'", teamName, *team.ID, organizationName)

			return *team.ID, nil
		}
	}

	return -1, fmt.Errorf("Could not find team '%s' in organization '%s'", teamName, organizationName)
}
