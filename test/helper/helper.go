package helper

import (
	"bytes"
	"encoding/gob"

	"github.com/garyburd/redigo/redis"

	"github.com/evalphobia/eurekache"
)

var testRedisHost = "127.0.0.1:6379"

func TestGetPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", testRedisHost)
		},
	}
}

func TestGobItem(v interface{}) []byte {
	item := &eurekache.Item{}
	item.Init()
	item.Value = v

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(item)
	return buf.Bytes()
}
