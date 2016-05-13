eurekache redis
====

# Requirements

Depends on the [garyburd/redigo](https://github.com/garyburd/redigo) library for Redis client.


# Installation

Install eurekache and required packages using `go get` command:

```bash
$ go get github.com/garyburd/redigo
```


# Usage

```go
// create redis cache
redisHost := "127.0.0.1:6379"
expiredTTL := 5 * 60 * 1000 // 5 minutes (millisecond)
keyPrefix := "myapp:" // added key prefix before set on redis
dbNumber := 1 // redis db number

pool := &redis.Pool{
    Dial: func() (redis.Conn, error) {
        return redis.Dial("tcp", redisHost)
    },
}

rc := NewRedisCache(pool)
rc.SetTTL(expiredTTL)
rc.SetPrefix(keyPrefix)
rc.Select(dbNumber)

cache := eurekache.New()
cache.SetCacheSources([]cache{rc})
```
