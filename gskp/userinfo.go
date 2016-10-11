package gskp

// UserInfo is a struct that contains information about a GitHub user,
// including Login Name and SSH Keys.
type UserInfo struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Keys  string `json:"keys"`
}
