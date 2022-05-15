# DirectCache

[![Build Status](https://github.com/qianbin/directcache/workflows/test/badge.svg)](https://github.com/qianbin/directcache/actions)
[![GoDoc](https://godoc.org/github.com/qianbin/directcache?status.svg)](http://godoc.org/github.com/qianbin/directcache)
[![Go Report](https://goreportcard.com/badge/github.com/qianbin/directcache)](https://goreportcard.com/report/github.com/qianbin/directcache)


A high performance GC-free cache library for Go-lang.

### Features

- high performance
- no GC overhead
- LRU & custom eviction policy
- zero-copy access
- small per-entry space overhead

### Installation

It requires Go 1.12 or newer.

```bash
go get github.com/qianbin/directcache
```

### Usage

```go
// capacity will be set to directcache.MinCapacity
c := directcache.New(0)

key := []byte("DirectCache")
val := []byte("is awesome,")

c.Set(key, val)
got, ok := c.Get(key)

fmt.Println(string(key), string(got), ok)
// Output: DirectCache is awesome, true
```

zero-copy access 

```go
c.AdvGet(
    key, 
    func(val []byte){        
        // consume the value
    },
    true, // peek, which means the recently-used flag won't be added to the accessed entry.
)
```

custom eviction policy
```go
shouldEvict := func(key, val []byte, recentlyUsed bool) bool {
    if recentlyUsed {
        return false
    }
    // custom rules...
    return true
}

c.SetEvictionPolicy(shouldEvict)
```

## License

The MIT License
