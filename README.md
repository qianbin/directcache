# DirectCache

[![Build Status](https://github.com/qianbin/directcache/workflows/test/badge.svg)](https://github.com/qianbin/directcache/actions)
[![GoDoc](https://godoc.org/github.com/qianbin/directcache?status.svg)](http://godoc.org/github.com/qianbin/directcache)
[![Go Report](https://goreportcard.com/badge/github.com/qianbin/directcache)](https://goreportcard.com/report/github.com/qianbin/directcache)


The high performance GC-free cache library for Golang.

### Features

- Fast set/get, scales well with the number of goroutines
- High hit rate, benefit from LRU & custom eviction policies
- GC friendly, almost zero GC overhead
- Zero-copy access

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

dump entries

```go
c.Dump(func(e directcache.Entry) bool {
    // do something to e
    return true    
})
```


### Benchmarks

The performance is compared with [FreeCache](https://github.com/coocood/freecache), [FastCache](https://github.com/VictoriaMetrics/fastcache) and [BigCache](https://github.com/allegro/bigcache). The code of benchmarks can be found under [./benches/](./benches/).

```bash
$ cd benches
$ go test -run=^$ -bench ^Benchmark benches -benchtime=5s
goos: linux
goarch: amd64
pkg: benches
cpu: Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz
BenchmarkGet/DirectCache-8             39183045      154.3 ns/op     155.50 MB/s
BenchmarkGet/FreeCache-8               23686060      253.4 ns/op      94.71 MB/s
BenchmarkGet/FastCache-8               34900104      170.7 ns/op     140.59 MB/s
BenchmarkGet/BigCache-8                27393757      217.9 ns/op     110.13 MB/s

BenchmarkParallelGet/DirectCache-8    168825828      35.39 ns/op     678.14 MB/s
BenchmarkParallelGet/FreeCache-8      100000000      56.37 ns/op     425.77 MB/s
BenchmarkParallelGet/FastCache-8      170344753      35.95 ns/op     667.52 MB/s
BenchmarkParallelGet/BigCache-8       135971068      46.87 ns/op     512.09 MB/s

BenchmarkSet/DirectCache-8             32157844      348.8 ns/op      68.82 MB/s
BenchmarkSet/FreeCache-8               33666162      410.0 ns/op      58.54 MB/s
BenchmarkSet/FastCache-8               28295484      222.3 ns/op     107.98 MB/s
BenchmarkSet/BigCache-8                18730782      471.1 ns/op      50.95 MB/s

BenchmarkParallelSet/DirectCache-8     92843146      80.81 ns/op     296.98 MB/s
BenchmarkParallelSet/FreeCache-8       76079636      99.11 ns/op     242.16 MB/s
BenchmarkParallelSet/FastCache-8       97058389      81.22 ns/op     295.50 MB/s
BenchmarkParallelSet/BigCache-8        68715756      99.88 ns/op     240.30 MB/s

BenchmarkParallelSetGet/DirectCache-8  34167814      202.3 ns/op     118.65 MB/s
BenchmarkParallelSetGet/FreeCache-8    23563779      294.8 ns/op      81.40 MB/s
BenchmarkParallelSetGet/FastCache-8    33523118      185.4 ns/op     129.46 MB/s
BenchmarkParallelSetGet/BigCache-8     25198658      248.0 ns/op      96.77 MB/s
PASS
ok  	benches	172.131s
```

```bash
$ go test -timeout 30s -run ^TestHitrate$ benches -v
=== RUN   TestHitrate
=== RUN   TestHitrate/DirectCache
    benches_test.go:181: hits: 741644	misses: 258356	hitrate: 74.16%
=== RUN   TestHitrate/DirectCache(custom_policy)
    benches_test.go:181: hits: 784456	misses: 215544	hitrate: 78.45%
=== RUN   TestHitrate/FreeCache
    benches_test.go:181: hits: 727308	misses: 272692	hitrate: 72.73%
=== RUN   TestHitrate/FastCache
    benches_test.go:181: hits: 690139	misses: 309861	hitrate: 69.01%
=== RUN   TestHitrate/BigCache
    benches_test.go:181: hits: 697831	misses: 302169	hitrate: 69.78%
--- PASS: TestHitrate (5.22s)
    --- PASS: TestHitrate/DirectCache (0.87s)
    --- PASS: TestHitrate/DirectCache(custom_policy) (1.08s)
    --- PASS: TestHitrate/FreeCache (1.01s)
    --- PASS: TestHitrate/FastCache (1.12s)
    --- PASS: TestHitrate/BigCache (1.14s)
PASS
ok  	benches	5.225s
```

## License

The MIT License
