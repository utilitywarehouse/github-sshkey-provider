package gskp

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
	// ErrCouldNotFetchGithubKeys is returned when GitHub responds with a non-202
	// HTTP status code.
	ErrCouldNotFetchGithubKeys = errors.New("Could not fetch SSH keys from Github")

	// ErrTeamNotFound is returned when a team cannot be found in the list of
	// the organization's teams.
	ErrTeamNotFound = errors.New("Team was not found in the organization")

	defaultGithubKeysURL = "https://github.com/%s.keys"
)

// UserInfo is a struct that contains information about a GitHub user,
// including Login Name and SSH Keys.
type UserInfo struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Keys  string `json:"keys"`
}

// KeyCollector fetches user information and their public SSH keys from GitHub.
type KeyCollector struct {
	githubClient  *github.Client
	httpClient    *http.Client
	githubKeysURL string
}

// NewKeyCollector returns an instantiated KeyCollector.
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

// GetTeamMemberInfo returns a slice of UserInfo structs, which contains
// information on the users that belong to the specified GitHub team.
func (k *KeyCollector) GetTeamMemberInfo(teamID int) ([]UserInfo, error) {
	memberInfo := []UserInfo{}

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

			user, _, err := k.githubClient.Users.GetByID(*tm.ID)
			if err != nil {
				simplelog.Infof("Could not fetch details for user '%s': %v", *tm.Login, err)
			} else if user.Name != nil {
				ui.Name = *user.Name
			}

			keys, err := k.getUserKeys(*tm.Login)
			if err != nil {
				simplelog.Infof("Could not fetch keys for user '%s': %v", *tm.Login, err)
			} else {
				ui.Keys = keys
			}

			if ui.Keys == "" {
				simplelog.Infof("No public SSH keys for user '%s'", *tm.Login)
			} else {
				memberInfo = append(memberInfo, ui)
			}
		}

		if resp.NextPage == 0 {
			simplelog.Debugf("GitHub API Limits: %d / %d until %s", resp.Remaining, resp.Limit, resp.Reset)

			break
		}
	}

	return memberInfo, nil
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

	if response.StatusCode != http.StatusOK {
		simplelog.Errorf("Could not fetch keys for user '%s': github returned status code %v", response.StatusCode)
		return "", ErrCouldNotFetchGithubKeys
	}

	return strings.TrimSpace(string(body)), nil
}
