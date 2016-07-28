package simplelog

func ExampleInfo() {
	Info("this is an info log message")
	// Output: {"kind":"info","message":"this is an info log message"}
}

func ExampleInfo_withArgs() {
	Info("this is an info log message with a string argument: %s", "argument_value")
	// Output: {"kind":"info","message":"this is an info log message with a string argument: argument_value"}
}

func ExampleDebug() {
	Debug("this is a debug log message")
	// Output: {"kind":"debug","message":"this is a debug log message"}
}

func ExampleDebug_withArgs() {
	Debug("this is a debug log message with a string argument: %s", "argument_value")
	// Output: {"kind":"debug","message":"this is a debug log message with a string argument: argument_value"}
}

func ExampleDebug_suppressed() {
	DebugEnabled = false
	Debug("this is a debug log message that should be suppressed")
	// Output:
}
