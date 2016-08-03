// +build integration

package transporter

import (
	"sync"
	"testing"
	"time"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

func init() {
	simplelog.DebugEnabled = true
}

func TestRedis(t *testing.T) {
	expectedMessage := `,1C73Fyxt[To|BOx7ixztgie\]Za@2h'GC-n'mQ_~rMO>u::^_}~O"(|Sk9&))<W`

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := NewRedis(":6379", "", "test_channel").Listen(func(msg string) error {
			if msg != expectedMessage {
				t.Fatalf("Redis.Listen received unexpected message: %s", msg)
			}
			return ErrListenerDisconnect
		}); err != nil {
			t.Fatalf("Redis.Listen returned an error: %v", err)

		}
	}()

	time.Sleep(100 * time.Millisecond)

	if err := NewRedis(":6379", "", "test_channel").Publish(expectedMessage); err != nil {
		t.Fatalf("Redis.Publish returned an error: %v", err)
	}

	wg.Wait()
}

func TestRedis_stopListening(t *testing.T) {
	l := NewRedis(":6379", "", "test_channel")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := l.Listen(func(msg string) error { return nil }); err != nil {
			t.Fatalf("Redis.Listen returned an error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	l.StopListening()

	wg.Wait()
}

func ExampleRedis_Listen_wrongPassword() {
	simplelog.MockClock()

	l := NewRedis(":6380", "", "test_channel")
	l.ReconnectBackoffMilliseconds = 100
	l.ReconnectAttempts = 2

	l.Listen(func(m string) error { return nil })

	// Output:
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Connected to redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error occured on listener: NOAUTH Authentication required."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Will try to reconnect to redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Disconnected from redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Trying to connect again in 100 milliseconds"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Connected to redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error occured on listener: NOAUTH Authentication required."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Will try to reconnect to redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Disconnected from redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Trying to connect again in 200 milliseconds"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Connected to redis at :6380"}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Error occured on listener: NOAUTH Authentication required."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Giving up trying to reconnect."}
	// {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"Disconnected from redis at :6380"}
}
