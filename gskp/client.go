package gskp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

var (
	clientHeaders = map[string]string{
		"User-Agent": "gskp/0.4",
	}

	// ErrClientPollTimeout is returned when a polling request times out.
	ErrClientPollTimeout = errors.New("longpolling timed out")
	// ErrClientUnexpected is returned when the collector replies with an
	// unexpected error.
	ErrClientUnexpected = errors.New("unexpected error")
	// ErrClientEmptyCollectorBaseURL is returned if trying to create a new
	// Client with an empty base URL.
	ErrClientEmptyCollectorBaseURL = errors.New("collectorBaseURL cannot be empty")
)

// Client is used by the agent to make requests to the collector service.
type Client struct {
	collectorBaseURL string
	timeoutSeconds   int64
	client           *http.Client
}

// NewClient creates and returns a new Client with the provided configuration.
func NewClient(collectorBaseURL string, timeoutSeconds int64) (*Client, error) {
	if collectorBaseURL == "" {
		return nil, ErrClientEmptyCollectorBaseURL
	}

	return &Client{
		collectorBaseURL: collectorBaseURL,
		timeoutSeconds:   timeoutSeconds,
		client:           &http.Client{},
	}, nil
}

// GetKeys requests the list of SSH keys from the collector.
func (c *Client) GetKeys(teamName string) ([]UserInfo, error) {
	return c.requestKeys(teamName, false)
}

// PollForKeys starts a longpoll request to watch for updates on the SSH keys.
func (c *Client) PollForKeys(teamName string) ([]UserInfo, error) {
	return c.requestKeys(teamName, true)
}

func (c *Client) requestKeys(teamName string, pollForChanges bool) ([]UserInfo, error) {
	u, err := url.Parse(c.collectorBaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "keys")

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	for k, v := range clientHeaders {
		req.Header.Set(k, v)
	}

	q := req.URL.Query()
	q.Add("team", teamName)
	if !pollForChanges {
		q.Add("init", "true")
	}
	if c.timeoutSeconds > 0 {
		q.Add("timeout", strconv.FormatInt(c.timeoutSeconds, 10))
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if bytes.Equal(body, serverUnexpectedError.Marshal()) {
		return nil, ErrClientUnexpected
	}

	if bytes.Equal(body, serverLongpollTimeout.Marshal()) {
		return nil, ErrClientPollTimeout
	}

	data := map[string][]UserInfo{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data["keys"], nil
}
