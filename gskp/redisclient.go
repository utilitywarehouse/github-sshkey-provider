package gskp

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

const (
	// RedisDefaultConnectTimeoutMilliseconds specifies the connect timeout.
	RedisDefaultConnectTimeoutMilliseconds = 1000
	// RedisDefaultReconnectBackoffMilliseconds specifies the time duration
	// that the client will wait before trying to re-connect to the redis host.
	RedisDefaultReconnectBackoffMilliseconds = 5000
	// RedisDefaultReconnectAttempts specifies the number of times that the
	// client will try to reconnect to the redis host.
	RedisDefaultReconnectAttempts = 3
)

var (
	// ErrRedisReconnectTriesExhausted is returned when there have been multiple
	// unsuccessful attempts to establish a connection to redis and the client
	// has decided to quit.
	ErrRedisReconnectTriesExhausted = errors.New("Exhausted reconnect attempts to redis")
)

// RedisClient provides a wrapper around redis functionality.
type RedisClient struct {
	Host                         string
	Password                     string
	Conn                         redis.Conn
	ConnectTimeoutMilliseconds   uint
	ReconnectBackoffMilliseconds uint
	ReconnectAttempts            uint
}

// NewRedisClient returns an instantiated RedisClient struct.
func NewRedisClient(host string, password string) *RedisClient {
	return &RedisClient{
		Host:                         host,
		Password:                     password,
		ConnectTimeoutMilliseconds:   RedisDefaultConnectTimeoutMilliseconds,
		ReconnectBackoffMilliseconds: RedisDefaultReconnectBackoffMilliseconds,
		ReconnectAttempts:            RedisDefaultReconnectAttempts,
	}
}

// Connect tries to establish a connection to the redis host. It will attempt
// to connect a number of times before giving up.
func (r *RedisClient) Connect() error {
	var conn redis.Conn
	var err error
	var reconnectAttemptCount uint

	for conn == nil {
		conn, err = redis.Dial("tcp", r.Host, r.getDialOptions()...)

		if err != nil {
			simplelog.Infof("Error connecting to redis: %v", err)

			if reconnectAttemptCount < r.ReconnectAttempts {
				backoffDuration := (reconnectAttemptCount + 1) * r.ReconnectBackoffMilliseconds

				simplelog.Infof("Reconnecting in %d milliseconds", backoffDuration)
				time.Sleep(time.Duration(backoffDuration) * time.Millisecond)

				reconnectAttemptCount++
			} else {
				simplelog.Infof("Will not attempt to reconnect: exhausted attempts")
				return ErrRedisReconnectTriesExhausted
			}
		}
	}

	r.Conn = conn

	simplelog.Infof("Connected to redis at %s", r.Host)

	return nil
}

// Disconnect tries to close the connection to the redis host.
func (r *RedisClient) Disconnect() error {
	if r.Conn != nil {
		if err := r.Conn.Close(); err != nil {
			return err
		}

		simplelog.Infof("Disconnected from redis at %s", r.Host)
	}

	return nil
}

// Reconnect attempts to disconnect from redis and connect again, after waiting
// for a number of milliseconds.
func (r *RedisClient) Reconnect(backoffDuration uint) error {
	simplelog.Infof("Will try to reconnect to redis at %s", r.Host)
	if err := r.Disconnect(); err != nil {
		return err
	}

	if backoffDuration == 0 {
		backoffDuration = r.ReconnectBackoffMilliseconds
	}

	simplelog.Infof("Trying to connect again in %d milliseconds", r.ReconnectBackoffMilliseconds)
	time.Sleep(time.Duration(backoffDuration) * time.Millisecond)

	if err := r.Connect(); err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) getDialOptions() []redis.DialOption {
	ret := []redis.DialOption{
		redis.DialConnectTimeout(time.Duration(r.ConnectTimeoutMilliseconds) * time.Millisecond),
	}

	if r.Password != "" {
		ret = append(ret, redis.DialPassword(r.Password))
	}

	return ret
}
