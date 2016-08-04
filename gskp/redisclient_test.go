package gskp

import (
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func ExampleRedisClient_Connect_reconnecErrorNoAddress() {
	simplelog.MockClock()

	rc := NewRedisClient("", "")
	rc.ConnectTimeoutMilliseconds = 100
	rc.ReconnectBackoffMilliseconds = 100
	rc.ReconnectAttempts = 2
	rc.Connect(true)

	// Output:
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error connecting to redis: dial tcp: missing address"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Reconnecting in 100 milliseconds"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error connecting to redis: dial tcp: missing address"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Reconnecting in 200 milliseconds"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error connecting to redis: dial tcp: missing address"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Will not attempt to reconnect: exhausted attempts"}
}
