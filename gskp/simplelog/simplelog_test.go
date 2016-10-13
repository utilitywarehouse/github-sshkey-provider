package simplelog

import "errors"

func ExampleInfof() {
	MockClock(true)
	defer MockClock(false)
	Infof("this is an info log message")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"this is an info log message"}
}

func ExampleInfof_arguments() {
	MockClock(true)
	defer MockClock(false)
	Infof("this is an info log message with a string argument: %s", "argument_value")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"info","message":"this is an info log message with a string argument: argument_value"}
}

func ExampleDebugf() {
	MockClock(true)
	defer MockClock(false)
	DebugEnabled = true
	Debugf("this is a debug log message")
	DebugEnabled = false
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"this is a debug log message"}
}

func ExampleDebugf_arguments() {
	MockClock(true)
	defer MockClock(false)
	DebugEnabled = true
	Debugf("this is a debug log message with a numeric argument: %d", 1234567890)
	DebugEnabled = false
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"debug","message":"this is a debug log message with a numeric argument: 1234567890"}
}

func ExampleDebugf_suppressed() {
	MockClock(true)
	defer MockClock(false)
	Debugf("this is a debug log message that should be suppressed")
	// Output:
}

func ExampleErrorf() {
	MockClock(true)
	defer MockClock(false)
	Errorf("this is an error log message")
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"error","message":"this is an error log message"}
}

func ExampleErrorf_arguments() {
	MockClock(true)
	defer MockClock(false)
	Errorf("this is an error log message with an error argument: %v", errors.New("this is an error"))
	// Output: {"timestamp":"2016-10-01T18:20:10.000000123+01:00","level":"error","message":"this is an error log message with an error argument: this is an error"}
}
