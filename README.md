eurekache
----

[![GoDoc][1]][2] [![License: MIT][3]][4] [![Release][5]][6] [![Build Status][7]][8] [![Co decov Coverage][11]][12] [![Go Report Card][13]][14] [![Code Climate][19]][20] [![BCH compliance][21]][22]

[1]: https://godoc.org/github.com/evalphobia/eurekache?status.svg
[2]: https://godoc.org/github.com/evalphobia/eurekache
[3]: https://img.shields.io/badge/License-MIT-blue.svg
[4]: LICENSE.md
[5]: https://img.shields.io/github/release/evalphobia/eurekache.svg
[6]: https://github.com/evalphobia/eurekache/releases/latest
[7]: https://travis-ci.org/evalphobia/eurekache.svg?branch=master
[8]: https://travis-ci.org/evalphobia/eurekache
[9]: https://coveralls.io/repos/evalphobia/eurekache/badge.svg?branch=master&service=github
[10]: https://coveralls.io/github/evalphobia/eurekache?branch=master
[11]: https://codecov.io/github/evalphobia/eurekache/coverage.svg?branch=master
[12]: https://codecov.io/github/evalphobia/eurekache?branch=master
[13]: https://goreportcard.com/badge/github.com/evalphobia/eurekache
[14]: https://goreportcard.com/report/github.com/evalphobia/eurekache
[15]: https://img.shields.io/github/downloads/evalphobia/eurekache/total.svg?maxAge=1800
[16]: https://github.com/evalphobia/eurekache/releases
[17]: https://img.shields.io/github/stars/evalphobia/eurekache.svg
[18]: https://github.com/evalphobia/eurekache/stargazers
[19]: https://codeclimate.com/github/evalphobia/eurekache/badges/gpa.svg
[20]: https://codeclimate.com/github/evalphobia/eurekache
[21]: https://bettercodehub.com/edge/badge/evalphobia/eurekache?branch=master
[22]: https://bettercodehub.com/


eurekache is cache library, implementing multiple cache source and fallback system.

# Supported Cache Type

- on memory
    - max size
    - expire ttl
- Redis

# Installation

Install eurekache and required packages using `go get` command:

```bash
$ go get github.com/evalphobia/eurekache
```


# Usage

## caches

### Memory cache

```go
// create on-memory cache
maxCacheItemSize := 100 // max allocated caches
expiredTTL := 5 * 60 * 1000 // 5 minutes (millisecond)

mc := memorycache.NewCacheTTL(maxCacheItemSize)
mc.SetTTL(expiredTTL)

cache := eurekache.New()
cache.SetCacheSources([]cache{mc})
```

### Redis cache

```go
import redigo "github.com/garyburd/redigo/redis"

// create redis cache
redisHost := "127.0.0.1:6379"
expiredTTL := 5 * 60 * 1000 // 5 minutes (millisecond)
keyPrefix := "myapp:" // added key prefix before set on redis
dbNumber := 1 // redis db number

pool := &redigo.Pool{
    Dial: func() (redigo.Conn, error) {
        return redigo.Dial("tcp", redisHost)
    },
}

rc := rediscache.NewRedisCache(pool)
rc.SetTTL(expiredTTL)
rc.SetPrefix(keyPrefix)
rc.Select(dbNumber)

cache := eurekache.New()
cache.SetCacheSources([]cache{rc})
```

### Multiple cache

```go
// search cache using from 1st cache to last cache by index order
cacheSources := []cache{mc, rc}

cache := eurekache.New()
cache.SetCacheSources(cacheSources)
```

## Set data

```go
cache := eurekache.New()
cache.SetCacheSources([]cache{mc, rc})


// save data to all of caches with default TTL
// when TTL=0, cache is not expired
cache.Set("key", "value")

// save data and cache lives on 24 hours
cache.SetExpire("key", "value", 24 * 60 * 60 * 1000)
```

Eurekache uses `encoding/gob` internally, you register your own types before use it.

```go
type MyType struct {
    Data interface{}
}

func init() {
    gob.Register(&MyType{})
}
```


## Get data

```go
cache := eurekache.New()
cache.SetCacheSources([]cache{mc, rc})

var ok bool // is cache existed or not

// pass pointer value; type must be equal
var stringValue string
ok = cache.Get("key", &stringValue)

// return interface value
var result interface{}
result, ok = cache.GetInterface("key")
stringValue, ok = result.(string)

// return []byte encoded by gob
var b []byte
b, ok := cache.GetGobBytes("key")
dec := gob.NewDecoder(bytes.NewBuffer(b))
err = dec.Decode(&stringValue)
```

# Contribution

Thanks!

Before create pull request, check the codes using below commands:

```bash
$ go vet
$ gofmt -s -l .
$ golint
```

And test on your local machine:

```bash
# install assertion library
$ go get github.com/stretchr/testify/assert

# you need to install and run redis-server before running test
$ go test ./...
```
