package simplelog

import "time"

func mockClock() {
	clock = func() time.Time { return time.Date(2016, 10, 1, 18, 20, 10, 123, time.Local) }
}

func ExampleInfo() {
	mockClock()
	Info("this is an info log message")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"this is an info log message"}
}

func ExampleInfo_withArgs() {
	Info("this is an info log message with a string argument: %s", "argument_value")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"this is an info log message with a string argument: argument_value"}
}

func ExampleDebug() {
	mockClock()
	Debug("this is a debug log message")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"this is a debug log message"}
}

func ExampleDebug_withArgs() {
	mockClock()
	Debug("this is a debug log message with a string argument: %s", "argument_value")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"this is a debug log message with a string argument: argument_value"}
}

func ExampleDebug_suppressed() {
	DebugEnabled = false
	Debug("this is a debug log message that should be suppressed")
	// Output:
}
