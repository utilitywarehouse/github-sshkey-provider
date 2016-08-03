// +build integration

package simplecache

import (
	"testing"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

const (
	redisIntegrationTestsDB = "11"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestRedis(t *testing.T) {
	rc := NewRedis(":6379", "", redisIntegrationTestsDB)
	rc.Flush()

	testValue1 := "mB32AHgcgi76g0KmmGr2YJMA4w50jNJyAb6IbknwiUcbNQlUw29XO2541xF2KLHR"
	testValue2 := "Ab6IbknwiUcbNQlUw29XO2541xF2KLHRmB32AHgcgi76g0KmmGr2YJMA4w50jNJy"

	rv0, err := rc.Get("test_key")
	if err != nil {
		t.Errorf("Redis.Get returned an error: %v", err)
	}
	if rv0 != "" {
		t.Errorf("Redis.Get returned an unexpected value: %s", rv0)
	}

	err = rc.Set("test_key", testValue1)
	if err != nil {
		t.Errorf("Redis.Set returned an error: %v", err)
	}

	rv1, err := rc.Get("test_key")
	if err != nil {
		t.Errorf("Redis.Get returned an error: %v", err)
	}
	if rv1 != testValue1 {
		t.Errorf("Redis.Get returned an unexpected value: %s", rv1)
	}

	err = rc.Set("test_key", testValue2)
	if err != nil {
		t.Errorf("Redis.Set returned an error: %v", err)
	}

	rv2, err := rc.Get("test_key")
	if err != nil {
		t.Errorf("Redis.Get returned an error: %v", err)
	}
	if rv2 != testValue2 {
		t.Errorf("Redis.Get returned an unexpected value: %s", rv2)
	}

	err = rc.Set("test_key", testValue2)
	if err != ErrValueHasNotChanged {
		t.Errorf("Redis.Set should have returned an ErrValueHasNotChanged error")
	}
}
