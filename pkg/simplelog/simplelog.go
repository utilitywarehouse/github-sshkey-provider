package simplelog

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	// DebugEnabled determines whether Debug-level log entries will be printed
	DebugEnabled = true

	clock = time.Now
)

type logEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Info prints an Info-level JSON formatted log entry to STDOUT
func Info(message string, args ...interface{}) error {
	return printLogMessage("info", message, args...)
}

// Debug prints an Debug-level JSON formatted log entry to STDOUT
func Debug(message string, args ...interface{}) error {
	if !DebugEnabled {
		return nil
	}

	return printLogMessage("debug", message, args...)
}

func printLogMessage(level string, message string, args ...interface{}) error {
	le := logEntry{
		Timestamp: clock(),
		Level:     level,
		Message:   fmt.Sprintf(message, args...),
	}

	output, _ := json.Marshal(le)

	fmt.Println(string(output))

	return nil
}
