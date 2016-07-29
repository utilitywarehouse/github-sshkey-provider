package transport

import (
	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/pkg/simplelog"
)

// RedisTransporter implements the Transporter interface with a redis backend
type RedisTransporter struct {
	host    string
	channel string
	pubsub  redis.PubSubConn
}

// NewRedisTransporter returns an instantiated RedisTransporter struct
func NewRedisTransporter(host string, channel string) *RedisTransporter {
	return &RedisTransporter{
		host:    host,
		channel: channel,
	}
}

// Publish opens a new connection to the redis host, publishes a message to a
// channel and closes the connection
func (t *RedisTransporter) Publish(message string) error {
	conn, err := redis.Dial("tcp", t.host)
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := conn.Do("PUBLISH", t.channel, message)
	if err != nil {
		return err
	}

	simplelog.Info("Published message to %d clients", n)

	return conn.Flush()
}

// Listen instructs the RedisTransporter to subscribe to a redis channel and
// start listening for messages
func (t *RedisTransporter) Listen(callback func(string) error) error {
	if err := t.initPubSubConnection(); err != nil {
		return err
	}
	defer t.pubsub.Close()

	simplelog.Info("Started listening for new messages")

	for {
		switch v := t.pubsub.Receive().(type) {
		case redis.Message:
			simplelog.Info("Received new message")

			if err := callback(string(v.Data)); err != nil {
				return err
			}
		case redis.Subscription:
			if v.Kind == "unsubscribe" {
				simplelog.Info("Stopped listening for messages")

				return nil
			}
		case error:
			return v
		}
	}
}

// StopListening instructs the RedisTransporter to unsubscribe from the channel
// and disconnect from the redis host
func (t *RedisTransporter) StopListening() error {
	return t.pubsub.Unsubscribe(t.channel)
}

func (t *RedisTransporter) initPubSubConnection() error {
	connection, err := redis.Dial("tcp", t.host)
	if err != nil {
		return err
	}

	t.pubsub = redis.PubSubConn{Conn: connection}

	return t.pubsub.Subscribe(t.channel)
}
