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

type RedisCache struct {
	pool       *redis.Pool
	dbno       string
	prefix     string
	defaultTTL int64
}

func NewRedisCache(pool *redis.Pool) *RedisCache {
	return &RedisCache{
		pool: pool,
		dbno: "0",
	}
}

func (c *RedisCache) SetPrefix(prefix string) {
	c.prefix = prefix
}

func (c *RedisCache) Select(num int) {
	c.dbno = strconv.Itoa(num)
}

func (c *RedisCache) Get(key string, data interface{}) bool {
	b, ok := c.GetGobByte(key)
	if !ok {
		return false
	}

	dec := gob.NewDecoder(bytes.NewBuffer(b))
	dec.Decode(data)

	return true
}

func (c *RedisCache) GetInterface(key string) (interface{}, bool) {
	item, ok := c.getGobItem(key)
	if !ok {
		return nil, false
	}

	return item.Value, true
}

func (c *RedisCache) GetGobByte(key string) ([]byte, bool) {
	item, ok := c.getGobItem(key)
	if !ok {
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

func (c *RedisCache) Set(key string, data interface{}) error {
	return c.SetExpire(key, data, c.defaultTTL)
}

func (c *RedisCache) SetExpire(key string, data interface{}, ttl int64) error {
	conn, err := c.conn()
	if err != nil {
		return err
	}
	defer conn.Close()

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
