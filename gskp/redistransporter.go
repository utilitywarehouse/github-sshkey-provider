package gskp

import (
	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

// RedisTransporter implements the transporter interface using a redis backend.
type RedisTransporter struct {
	*RedisClient
	channel            string
	pubsubConn         redis.PubSubConn
	listenerEnabled    bool
	listenerSubscribed bool
	listenerLastError  error
}

// NewRedisTransporter returns an instantiated Redis transporter struct.
func NewRedisTransporter(host string, password string, channel string) *RedisTransporter {
	return &RedisTransporter{
		NewRedisClient(host, password),
		channel,
		redis.PubSubConn{},
		true,
		false,
		nil,
	}
}

// Publish opens a new connection to the redis host, publishes a message to a
// channel and closes the connection.
func (t *RedisTransporter) Publish(message string) error {
	if err := t.Connect(true); err != nil {
		return err
	}
	defer t.Disconnect()

	n, err := redis.Int(t.Conn.Do("PUBLISH", t.channel, message))
	if err != nil {
		return err
	}

	simplelog.Infof("Published message to %d clients", n)

	return nil
}

// Listen instructs the RedisTransporter to subscribe to a redis channel and
// start listening for messages.
func (t *RedisTransporter) Listen(callback func(string) error) error {
	t.listenerEnabled = true
	t.listenerLastError = nil

	if err := t.Connect(true); err != nil {
		return err
	}
	defer t.Disconnect()

	for t.listenerEnabled {
		t.pubsubConn = redis.PubSubConn{Conn: t.Conn}

		if err := t.pubsubConn.Subscribe(t.channel); err != nil {
			return err
		}

	PubSubReceiveLoop:
		for {
			switch v := t.pubsubConn.Receive().(type) {
			case redis.Message:
				simplelog.Infof("Received new message")

				if err := callback(string(v.Data)); err != nil {
					if err == ErrListenerDisconnect {
						return nil
					}

					return err
				}
			case redis.Subscription:
				if v.Kind == "unsubscribe" {
					simplelog.Infof("Stopped listening for messages")

					t.listenerSubscribed = false

					return nil
				} else if v.Kind == "subscribe" {
					simplelog.Infof("Started listening for new messages")

					t.listenerSubscribed = true
					t.listenerLastError = nil
				}
			case error:
				simplelog.Infof("Error occurred on listener: %v", v)

				if t.listenerLastError == v {
					simplelog.Errorf("Error occurred twice in a row, giving up")
					return v
				}

				t.listenerLastError = v

				if err := t.Reconnect(0); err != nil {
					return err
				}

				break PubSubReceiveLoop
			}
		}
	}

	return nil
}

// StopListening instructs the listener to stop listening and return.
func (t *RedisTransporter) StopListening() error {
	t.listenerEnabled = false

	if t.listenerSubscribed {
		return t.pubsubConn.Unsubscribe(t.channel)
	}

	return nil
}
