package gskp

import "encoding/json"

// UserInfo is a struct that contains information about a GitHub user,
// including Login Name and SSH Keys.
type UserInfo struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Keys  string `json:"keys"`
}

// UserInfoList is simply a collection of UserInfo structs.
type UserInfoList []UserInfo

// Marshal returns a string containing the serialised JSON version of the object.
func (ui *UserInfoList) Marshal() (string, error) {
	jsonText, err := json.Marshal(ui)
	if err != nil {
		return "", err
	}

	return string(jsonText), err
}

// Unmarshal takes JSON input and unserialises it into the struct.
func (ui *UserInfoList) Unmarshal(data string) error {
	if err := json.Unmarshal([]byte(data), ui); err != nil {
		return err
	}

	return nil
}
