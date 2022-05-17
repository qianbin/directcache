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

### Benchmarks

```bash
$ cd benches
$ go test -benchmem -run=^$ -bench ^Benchmark benches -benchtime=5s
goos: linux
goarch: amd64
pkg: benches
cpu: Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz
BenchmarkGet/DirectCache-8  38625438    153.4 ns/op 156.46 MB/s 16 B/op 1 allocs/op
BenchmarkGet/FreeCache-8    26123972    230.3 ns/op 104.19 MB/s 39 B/op 1 allocs/op
BenchmarkGet/FastCache-8    38518879    158.9 ns/op 151.03 MB/s 16 B/op 1 allocs/op
BenchmarkSet/DirectCache-8  28563682    356.2 ns/op  67.37 MB/s  4 B/op 0 allocs/op
BenchmarkSet/FreeCache-8    27578604    325.1 ns/op  73.83 MB/s  2 B/op 0 allocs/op
BenchmarkSet/FastCache-8    30179632    215.8 ns/op 111.21 MB/s  7 B/op 0 allocs/op
PASS
ok  	benches	46.751s
```

```bash
$ go test -timeout 30s -run ^TestHitrate$ benches -v
=== RUN   TestHitrate
=== RUN   TestHitrate/DirectCache
    benches_test.go:143: hits: 739035	misses: 260965	hitrate: 73.90%
=== RUN   TestHitrate/FreeCache
    benches_test.go:143: hits: 727308	misses: 272692	hitrate: 72.73%
=== RUN   TestHitrate/FastCache
    benches_test.go:143: hits: 696096	misses: 303904	hitrate: 69.61%
--- PASS: TestHitrate (1.97s)
    --- PASS: TestHitrate/DirectCache (0.58s)
    --- PASS: TestHitrate/FreeCache (0.66s)
    --- PASS: TestHitrate/FastCache (0.72s)
PASS
ok  	benches	1.973s
```

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
