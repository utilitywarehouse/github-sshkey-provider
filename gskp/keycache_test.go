package gskp

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

var (
	// testKeyCache is the KeyCache being tested
	testKeyCache *KeyCache
)

func init() {
	simplelog.DebugEnabled = true
}

func TestKeyCache_Get_new(t *testing.T) {
	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	testKeyCache = NewKeyCache("none", "", 5*time.Second)
	testKeyCache.collector = testKeyCollector

	dataExpected := []byte(`[{"login":"user","id":999999,"name":"User Name","keys":"ssh-rsa this_will_be_a_really_really_really_long_ssh_key_string"}]`)

	data, err := testKeyCache.Get("Owners")
	if err != nil {
		t.Fatalf("KeyCache.Get returned an error: %v", err)
	}

	if !bytes.Equal(data, dataExpected) {
		t.Errorf("KeyCache.Get returned unexpected value: %v", string(data))
	}
}

func ExampleKeyCache_Get_twice() {
	simplelog.MockClock(true)
	defer simplelog.MockClock(false)

	mockSetup()
	mockInstallHandlers([]string{"orgTeams", "userKeys", "userInfo", "teamUserList"})
	defer mockTeardown()

	testKeyCache = NewKeyCache("none", "", 5*time.Second)
	testKeyCache.collector = testKeyCollector

	_, err := testKeyCache.Get("Owners")
	if err != nil {
		simplelog.Errorf("KeyCache.Get returned an error: %v", err)
	}

	_, err = testKeyCache.Get("Owners")
	if err != nil {
		simplelog.Errorf("KeyCache.Get returned an error: %v", err)
	}

	// Output:
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"keys not found in cache, updating..."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching list of teams for organization 'none'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Team 'Owners' with id 888888 found in organization 'none'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching a list of users in team with ID 888888"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching keys for user 'user'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"GitHub API Limits: 0 / 0 until 0001-01-01 00:00:00 +0000 UTC"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"found recent keys in the cache"}
}

func ExampleKeyCache_Get_late() {
	simplelog.MockClock(true)
	defer simplelog.MockClock(false)

	mockSetup()
	mockInstallHandlers([]string{"userKeys", "userInfo", "teamUserList"})
	// custom handler here to add delay
	testMux.HandleFunc("/orgs/none/teams", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, `[{"name": "Owners", "id": 888888}]`)
	})
	defer mockTeardown()

	testKeyCache = NewKeyCache("none", "", 5*time.Second)
	testKeyCache.collector = testKeyCollector

	go func() {
		_, err := testKeyCache.Get("Owners")
		if err != nil {
			simplelog.Errorf("KeyCache.Get returned an error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	_, err := testKeyCache.Get("Owners")
	if err != nil {
		simplelog.Errorf("KeyCache.Get returned an error: %v", err)
	}

	// Output:
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"keys not found in cache, updating..."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching list of teams for organization 'none'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"keys not found in cache, updating..."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Team 'Owners' with id 888888 found in organization 'none'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching a list of users in team with ID 888888"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"Fetching keys for user 'user'"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"GitHub API Limits: 0 / 0 until 0001-01-01 00:00:00 +0000 UTC"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"keys are already up to date, won't update"}
}
