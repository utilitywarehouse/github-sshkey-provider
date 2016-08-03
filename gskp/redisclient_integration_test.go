// +build integration

package gskp

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

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
	if err := rc.Connect(); err != nil {
		t.Fatalf("RedisClient.Connect returned an error: %v", err)
	}

	if err := rc.Reconnect(); err != nil {
		t.Fatalf("RedisClient.Connect returned an error: %v", err)
	}

	expectedValue := "u3h5rEj-wl1SVm%i45JLt3$khjZv348j_Fdk"

	rs, err := rc.Conn.Do("SET", "test_key", expectedValue)
	if err != nil {
		t.Fatalf("Got an error on SET: %v", err)
	}
	if rs == nil {
		t.Fatalf("Got an error on SET, return value is nil")
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
