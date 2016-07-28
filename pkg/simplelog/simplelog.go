package simplelog

import (
	"encoding/json"
	"fmt"
)

var (
	// DebugEnabled determines whether Debug-level log entries will be printed
	DebugEnabled = true
)

type logEntry struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
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

func printLogMessage(kind string, message string, args ...interface{}) error {
	le := logEntry{
		Kind:    kind,
		Message: fmt.Sprintf(message, args...),
	}

	output, _ := json.Marshal(le)

	fmt.Println(string(output))

	return nil
}
