package transport

// Transporter is an interface that is used for communication between
// components of the application
type Transporter interface {
	Publish(string)
	Listen(func(string) error)
	StopListening()
}
