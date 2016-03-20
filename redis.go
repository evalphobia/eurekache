package eurekache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

var (
	errNilPool    = errors.New("redis.Pool is nil")
	errClosedConn = errors.New("redis.Conn is closed")
)

// RedisCache is a cache source for Redis and contains redis.Pool
type RedisCache struct {
	pool       *redis.Pool
	dbno       string
	prefix     string
	defaultTTL int64
}

// NewRedisCache returns initialized RedisCache with given redis.Pool
func NewRedisCache(pool *redis.Pool) *RedisCache {
	return &RedisCache{
		pool: pool,
		dbno: "0",
	}
}

// SetTTL sets default TTL (milliseconds)
func (c *RedisCache) SetTTL(ttl int64) {
	c.defaultTTL = ttl
}

// SetPrefix sets the prefix used for adding prefix into key name
func (c *RedisCache) SetPrefix(prefix string) {
	c.prefix = prefix
}

// Select sets db number for redis-server
func (c *RedisCache) Select(num int) {
	c.dbno = strconv.Itoa(num)
}

// Get searches cache by given key from redis and returns flag of cache is existed or not.
// when cache hit, data is assigned.
func (c *RedisCache) Get(key string, data interface{}) bool {
	b, ok := c.GetGobBytes(key)
	if !ok {
		return false
	}

	dec := gob.NewDecoder(bytes.NewBuffer(b))
	dec.Decode(data)
	return true
}

// GetInterface searches cache by given key from redis and returns interface value.
func (c *RedisCache) GetInterface(key string) (interface{}, bool) {
	item, ok := c.getGobItem(key)
	if !ok {
		return nil, false
	}

	return item.Value, true
}

// GetGobBytes searches cache by given key from redis and returns gob-encoded value.
func (c *RedisCache) GetGobBytes(key string) ([]byte, bool) {
	item, ok := c.getGobItem(key)
	switch {
	case !ok:
		return nil, false
	case item.Value == nil:
		return nil, false
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item.Value)
	if err != nil {
		return nil, false
	}

	return buf.Bytes(), true
}

// getGobItem searches cache by given key from redis and returns Item data
func (c *RedisCache) getGobItem(key string) (*Item, bool) {
	conn, err := c.conn()
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	data, err := conn.Do("GET", c.prefix+key)
	if err != nil {
		return nil, false
	}

	b, err := redis.Bytes(data, err)
	if err != nil {
		return nil, false
	}

	var item Item
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err = dec.Decode(&item)
	if err != nil {
		return nil, false
	}

	return &item, true
}

// Set sets data into redis. data is wrapped by gob-encoded Item
func (c *RedisCache) Set(key string, data interface{}) error {
	return c.SetExpire(key, data, c.defaultTTL)
}

// SetExpire sets data into redis with TTL. data is wrapped by gob-encoded Item
func (c *RedisCache) SetExpire(key string, data interface{}, ttl int64) error {
	conn, err := c.conn()
	if err != nil {
		return err
	}
	defer conn.Close()

	if data == nil {
		_, err = conn.Do("DEL", c.prefix+key)
		return err
	}

	item := Item{}
	item.init()
	item.SetExpire(ttl)
	item.Value = data

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(item)
	if err != nil {
		return err
	}

	switch {
	case ttl < 1:
		_, err = conn.Do("SET", c.prefix+key, buf.Bytes())
	default:
		ttl = ttl / 1000 // convert milli-sec to sec
		_, err = conn.Do("SETEX", c.prefix+key, ttl, buf.Bytes())
	}

	if err != nil {
		return err
	}
	return nil
}

// conn returns redis.Conn created from redis.Pool
func (c *RedisCache) conn() (redis.Conn, error) {
	if c.pool == nil {
		return nil, errNilPool
	}

	conn := c.pool.Get()
	err := conn.Err()
	if err != nil {
		return nil, err
	}

	_, err = conn.Do("SELECT", c.dbno)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
