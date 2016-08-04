// +build integration

package gskp

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestRedisClient_noPassword(t *testing.T) {
	rc := NewRedisClient(":6379", "")
	rc.ReconnectBackoffMilliseconds = 100

	testRedisClient(t, rc)
}

func TestRedisClient_withPassword(t *testing.T) {
	rc := NewRedisClient(":6380", "password")
	rc.ReconnectBackoffMilliseconds = 100

	testRedisClient(t, rc)
}

func testRedisClient(t *testing.T, rc *RedisClient) {
	if err := rc.Connect(false); err != nil {
		t.Fatalf("RedisClient.Connect returned an error: %v", err)
	}

	if err := rc.Reconnect(0); err != nil {
		t.Fatalf("RedisClient.Connect returned an error: %v", err)
	}

	if _, err := redis.String(rc.Conn.Do("FLUSHDB")); err != nil {
		t.Fatalf("Got an error on FLUSHDB: %v", err)
	}

	expectedValue := "u3h5rEj-wl1SVm%i45JLt3$khjZv348j_Fdk"

	_, err := redis.String(rc.Conn.Do("SET", "test_key", expectedValue))
	if err != nil {
		t.Fatalf("Got an error on SET: %v", err)
	}

	rg, err := redis.String(rc.Conn.Do("GET", "test_key"))
	if err != nil {
		t.Fatalf("Got an error on GET: %v", err)
	}
	if rg != expectedValue {
		t.Errorf("Got an error on GET, return value is not the expected: %s", expectedValue)
	}

	re, err := redis.Int64(rc.Conn.Do("TTL", "test_key"))
	if err != nil {
		t.Fatalf("Got an error on TTL: %v", err)
	}
	if re != -1 {
		t.Errorf("Got a TTL of %d seconds, was expecting -1", re)
	}
}
