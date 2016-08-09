package gskp

import "errors"

var (
	// ErrListenerDisconnect is an error that, if returned by the listener
	// callback, will cause the listener to disconnect, without returning an
	// error to the caller.
	ErrListenerDisconnect = errors.New("Listener callback has requested to disconnect.")
)

// Transporter to be implemented for providing communication capabilities
// between the components of the application.
type Transporter interface {
	Publish(string) error
	Listen(func(string) error) error
	StopListening() error
}
