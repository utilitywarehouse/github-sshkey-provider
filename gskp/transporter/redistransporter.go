package transporter

import (
	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

// Redis implements the transporter interface using a redis backend.
type Redis struct {
	*gskp.RedisClient
	channel                string
	pubsubConn             redis.PubSubConn
	listenerReconnectCount uint
	listenerEnabled        bool
	listenerSubscribed     bool
}

// NewRedis returns an instantiated Redis transporter struct.
func NewRedis(host string, password string, channel string) *Redis {
	return &Redis{
		gskp.NewRedisClient(host, password),
		channel,
		redis.PubSubConn{},
		0,
		true,
		false,
	}
}

// Publish opens a new connection to the redis host, publishes a message to a
// channel and closes the connection.
func (t *Redis) Publish(message string) error {
	if err := t.Connect(); err != nil {
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

// Listen instructs the Redis transporter to subscribe to a redis channel and start
// listening for messages.
func (t *Redis) Listen(callback func(string) error) error {
	t.listenerEnabled = true
	t.listenerReconnectCount = 0

	if err := t.Connect(); err != nil {
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

					t.listenerReconnectCount = 0
				}
			case error:
				simplelog.Infof("Error occured on listener: %v", v)

				if err := t.listenerReconnect(); err != nil {
					return err
				}

				break PubSubReceiveLoop
			}
		}
	}

	return nil
}

// StopListening instructs the listener to stop listening and return.
func (t *Redis) StopListening() error {
	t.listenerEnabled = false

	if t.listenerSubscribed {
		return t.pubsubConn.Unsubscribe(t.channel)
	}

	return nil
}

func (t *Redis) listenerReconnect() error {
	if t.listenerReconnectCount >= t.ReconnectAttempts {
		simplelog.Infof("Giving up trying to reconnect.")

		return gskp.ErrRedisReconnectTriesExhausted
	}

	if err := t.Reconnect((t.listenerReconnectCount + 1) * t.ReconnectBackoffMilliseconds); err != nil {
		return err
	}

	t.listenerReconnectCount++

	return nil
}
