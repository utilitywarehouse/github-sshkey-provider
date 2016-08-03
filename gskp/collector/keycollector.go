package collector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
	"golang.org/x/oauth2"
)

var (
	// ErrGithubKeysNotFound is returned when Github responds with "Not Found"
	// when trying to get a user's keys.
	ErrGithubKeysNotFound = errors.New("Response was 'Not Found'")

	// ErrTeamNotFound is returned when a team cannot be found in the list of
	// the organization's teams.
	ErrTeamNotFound = errors.New("Team was not found in the organization")

	defaultGithubKeysURL = "https://github.com/%s.keys"
)

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
	tc := oauth2.NewClient(
		oauth2.NoContext,
		oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubAccessToken},
		),
	)

	return &KeyCollector{
		githubClient:  github.NewClient(tc),
		httpClient:    &http.Client{},
		githubKeysURL: defaultGithubKeysURL,
	}
}

// GetTeamMemberInfo returns a slice of UserInfo structs, which contains
// information on the users that belong to the specified GitHub team.
func (k *KeyCollector) GetTeamMemberInfo(teamID int) (UserInfoList, error) {
	memberInfo, err := k.getTeamMembers(teamID)
	if err != nil {
		return nil, err
	}

	for i := range memberInfo {
		k.setUserName(&memberInfo[i])
		k.setUserKeys(&memberInfo[i])
	}

	return memberInfo, nil
}

// GetTeamID finds the GitHub team id, based on the organization and team
// names.
func (k *KeyCollector) GetTeamID(organizationName string, teamName string) (int, error) {
	simplelog.Debugf("Fetching list of teams for organization '%s'", organizationName)

	orgTeams, _, err := k.githubClient.Organizations.ListTeams(organizationName, nil)
	if err != nil {
		return -1, err
	}

	for _, team := range orgTeams {
		if *team.Name == teamName {
			simplelog.Debugf("Team '%s' with id %d found in organization '%s'", teamName, *team.ID, organizationName)

			return *team.ID, nil
		}
	}

	return -1, ErrTeamNotFound
}

func (k *KeyCollector) getTeamMembers(teamID int) (UserInfoList, error) {
	memberInfo := UserInfoList{}

	ltmOpts := &github.OrganizationListTeamMembersOptions{
		Role: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 100,
		},
	}

	simplelog.Debugf("Fetching a list of users in team with ID %d", teamID)

	for {
		teamMembers, resp, err := k.githubClient.Organizations.ListTeamMembers(teamID, ltmOpts)
		if err != nil {
			return nil, err
		}

		for _, tm := range teamMembers {
			ui := UserInfo{
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

func (k *KeyCollector) setUserName(ui *UserInfo) {
	user, _, err := k.githubClient.Users.GetByID(ui.ID)
	if err != nil {
		simplelog.Infof("Could not fetch details for user '%s': %v", ui.Login, err)
	} else {
		ui.Name = *user.Name
	}
}

func (k *KeyCollector) setUserKeys(ui *UserInfo) {
	keys, err := k.getUserKeys(ui.Login)
	if err != nil {
		simplelog.Infof("Could not fetch keys for user '%s': %s", ui.Login, err.Error())
	} else {
		ui.Keys = keys
	}
}

func (k *KeyCollector) getUserKeys(userLogin string) (string, error) {
	// Instead of using github.Users.ListKeys() which calls the GitHub API and is
	// a throttled request, we simply fetch them from the public URL that is
	// provided by GitHub.
	simplelog.Debugf("Fetching keys for user '%s'", userLogin)

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
		simplelog.Debugf("Github responed with 'Not Found' when looking for the keys of user '%s'", userLogin)
		return "", ErrGithubKeysNotFound
	}

	return keys, nil
}
