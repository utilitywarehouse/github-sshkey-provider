package simplecache

import (
	"github.com/garyburd/redigo/redis"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

// Redis implements a redis-backed simplecache.
type Redis struct {
	*gskp.RedisClient
	database string
}

// NewRedis returns an instantiated Redis simplecache.
func NewRedis(host string, password string, database string) *Redis {
	return &Redis{
		gskp.NewRedisClient(host, password),
		database,
	}
}

// Set will set a key-value pair in the cache.
func (c *Redis) Set(key string, value string) error {
	if err := c.Connect(false); err != nil {
		return err
	}
	defer c.Disconnect()

	_, err := redis.String(c.Conn.Do("SELECT", c.database))
	if err != nil {
		return err
	}
	simplelog.Debugf("Selected DB %s", c.database)

	v, err := redis.String(c.Conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		return err
	}
	if err == nil && v == value {
		simplelog.Debugf("Value for key '%s' in cache is the same, will not set", key)
		return ErrValueHasNotChanged
	}

	_, err = redis.String(c.Conn.Do("SET", key, value))
	if err != nil {
		return err
	}

	simplelog.Infof("Updated value for key '%s' in the cache", key)

	return nil
}

// Get will retrieve a value pair from the cache.
func (c *Redis) Get(key string) (string, error) {
	if err := c.Connect(false); err != nil {
		return "", err
	}
	defer c.Disconnect()

	_, err := redis.String(c.Conn.Do("SELECT", c.database))
	if err != nil {
		return "", err
	}
	simplelog.Debugf("Selected DB %s", c.database)

	v, err := redis.String(c.Conn.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		return "", err
	}

	if err == redis.ErrNil {
		simplelog.Debugf("Key '%s' does not exist, returning empty value", key)
		return "", nil
	}

	return v, nil
}

// Flush will clear the entire cache.
func (c *Redis) Flush() error {
	if err := c.Connect(false); err != nil {
		return err
	}
	defer c.Disconnect()

	_, err := redis.String(c.Conn.Do("SELECT", c.database))
	if err != nil {
		return err
	}
	simplelog.Debugf("Selected DB %s", c.database)

	_, err = redis.String(c.Conn.Do("FLUSHDB"))
	if err != nil {
		return err
	}

	return nil
}
