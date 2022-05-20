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
$ go test -run=^$ -bench ^Benchmark benches -benchtime=5s
goos: linux
goarch: amd64
pkg: benches
cpu: Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz
BenchmarkGet/DirectCache-8               39089787     154.7 ns/op    155.16 MB/s
BenchmarkGet/FreeCache-8                 26036332     228.4 ns/op    105.08 MB/s
BenchmarkGet/FastCache-8                 38764371     158.6 ns/op    151.28 MB/s
BenchmarkParallelGet/DirectCache-8      161248044     37.18 ns/op    645.52 MB/s
BenchmarkParallelGet/FreeCache-8        100000000     52.67 ns/op    455.68 MB/s
BenchmarkParallelGet/FastCache-8        180940428     33.24 ns/op    722.07 MB/s
BenchmarkSet/DirectCache-8               27090906     367.3 ns/op     65.35 MB/s
BenchmarkSet/FreeCache-8                 26872202     334.9 ns/op     71.67 MB/s
BenchmarkSet/FastCache-8                 28796907     222.3 ns/op    107.96 MB/s
BenchmarkParallelSet/DirectCache-8       83409436     84.57 ns/op    283.79 MB/s
BenchmarkParallelSet/FreeCache-8         73930858     85.63 ns/op    280.28 MB/s
BenchmarkParallelSet/FastCache-8         83161243     86.25 ns/op    278.25 MB/s
BenchmarkParallelSetGet/DirectCache-8    32223900     213.9 ns/op    112.22 MB/s
BenchmarkParallelSetGet/FreeCache-8      23440317     290.7 ns/op     82.57 MB/s
BenchmarkParallelSetGet/FastCache-8      33250729     178.9 ns/op    134.19 MB/s
PASS
ok  	benches	114.304s
```

```bash
$ go test -timeout 30s -run ^TestHitrate$ benches -v
=== RUN   TestHitrate
=== RUN   TestHitrate/DirectCache
    benches_test.go:164: hits: 739035	misses: 260965	hitrate: 73.90%
=== RUN   TestHitrate/DirectCache(custom_policy)
    benches_test.go:164: hits: 780808	misses: 219192	hitrate: 78.08%
=== RUN   TestHitrate/FreeCache
    benches_test.go:164: hits: 727308	misses: 272692	hitrate: 72.73%
=== RUN   TestHitrate/FastCache
    benches_test.go:164: hits: 696096	misses: 303904	hitrate: 69.61%
--- PASS: TestHitrate (2.72s)
    --- PASS: TestHitrate/DirectCache (0.58s)
    --- PASS: TestHitrate/DirectCache(custom_policy) (0.75s)
    --- PASS: TestHitrate/FreeCache (0.67s)
    --- PASS: TestHitrate/FastCache (0.72s)
PASS
ok  	benches	2.722s
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
shouldEvict := func(entry directcache.Entry) bool {
    if entry.RecentlyUsed() {
        return false
    }
    // custom rules...
    return true
}

c.SetEvictionPolicy(shouldEvict)
```

## License

The MIT License
