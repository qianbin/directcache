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
BenchmarkGet/DirectCache-8              39112850     153.4 ns/op     156.47 MB/s
BenchmarkGet/FreeCache-8                26317232     226.3 ns/op     106.07 MB/s
BenchmarkGet/FastCache-8                37878956     159.6 ns/op     150.35 MB/s
BenchmarkParallelGet/DirectCache-8     184576632     32.78 ns/op     732.18 MB/s
BenchmarkParallelGet/FreeCache-8       171217690     35.01 ns/op     685.54 MB/s
BenchmarkParallelGet/FastCache-8       178660434     32.59 ns/op     736.36 MB/s
BenchmarkSet/DirectCache-8              27788277     364.0 ns/op      65.94 MB/s
BenchmarkSet/FreeCache-8                26571054     331.6 ns/op      72.38 MB/s
BenchmarkSet/FastCache-8                28562538     219.9 ns/op     109.15 MB/s
BenchmarkParallelSet/DirectCache-8      62751762     103.5 ns/op     231.94 MB/s
BenchmarkParallelSet/FreeCache-8        62635524     105.2 ns/op     228.10 MB/s
BenchmarkParallelSet/FastCache-8        60270451     96.31 ns/op     249.20 MB/s
BenchmarkParallelSetGet/DirectCache-8   31439042     219.8 ns/op     109.20 MB/s
BenchmarkParallelSetGet/FreeCache-8     24580938     274.6 ns/op      87.41 MB/s
BenchmarkParallelSetGet/FastCache-8     31697812     185.8 ns/op     129.18 MB/s
PASS
ok  	benches	116.088s
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
