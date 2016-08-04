package simplelog

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	// DebugEnabled determines whether Debug-level log entries will be printed.
	DebugEnabled = false

	clock = time.Now
)

// MockClock sets the timestamp to a fixed value. It is meant to be used in
// tests.
func MockClock() {
	clock = func() time.Time { return time.Date(2016, 10, 1, 18, 20, 10, 123, time.Local) }
}

type logEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Debugf prints an Debug-level JSON formatted log entry to STDOUT.
func Debugf(message string, args ...interface{}) {
	if DebugEnabled {
		printLogMessage("debug", message, args...)
	}
}

// Infof prints an Info-level JSON formatted log entry to STDOUT.
func Infof(message string, args ...interface{}) {
	printLogMessage("info", message, args...)
}

// Errorf prints an Error-level JSON formatted log entry to STDOUT.
func Errorf(message string, args ...interface{}) {
	printLogMessage("error", message, args...)
}

func printLogMessage(level string, message string, args ...interface{}) {
	le := logEntry{
		Timestamp: clock(),
		Level:     level,
		Message:   fmt.Sprintf(message, args...),
	}

	output, _ := json.Marshal(le)

	fmt.Println(string(output))
}
