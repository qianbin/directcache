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
BenchmarkGet/DirectCache-8               37590464      158.7 ns/op    151.22 MB/s
BenchmarkGet/FreeCache-8                 25786171      230.9 ns/op    103.96 MB/s
BenchmarkGet/FastCache-8                 37254318      159.3 ns/op    150.64 MB/s
BenchmarkGet/BigCache-8                  37842639      186.1 ns/op    128.97 MB/s

BenchmarkParallelGet/DirectCache-8      161808373      36.99 ns/op    648.87 MB/s
BenchmarkParallelGet/FreeCache-8        100000000      53.11 ns/op    451.90 MB/s
BenchmarkParallelGet/FastCache-8        179409560      33.35 ns/op    719.61 MB/s
BenchmarkParallelGet/BigCache-8         134716734      44.69 ns/op    536.99 MB/s

BenchmarkSet/DirectCache-8               25885528      371.9 ns/op     64.54 MB/s
BenchmarkSet/FreeCache-8                 26270774      335.2 ns/op     71.60 MB/s
BenchmarkSet/FastCache-8                 28288302      225.2 ns/op    106.58 MB/s
BenchmarkSet/BigCache-8                  19396027      462.2 ns/op     51.93 MB/s

BenchmarkParallelSet/DirectCache-8       84369208      85.00 ns/op    282.34 MB/s
BenchmarkParallelSet/FreeCache-8         73204508      86.74 ns/op    276.68 MB/s
BenchmarkParallelSet/FastCache-8         82061668      86.49 ns/op    277.50 MB/s
BenchmarkParallelSet/BigCache-8          71246137      96.85 ns/op    247.81 MB/s

BenchmarkParallelSetGet/DirectCache-8    32410581      215.7 ns/op    111.26 MB/s
BenchmarkParallelSetGet/FreeCache-8      23206942      293.7 ns/op     81.71 MB/s
BenchmarkParallelSetGet/FastCache-8      32624253      179.3 ns/op    133.87 MB/s
BenchmarkParallelSetGet/BigCache-8       24936825      247.4 ns/op     97.02 MB/s
PASS
ok  	benches	156.348s
```

```bash
$ go test -timeout 30s -run ^TestHitrate$ benches -v
=== RUN   TestHitrate
=== RUN   TestHitrate/DirectCache
    benches_test.go:175: hits: 739035	misses: 260965	hitrate: 73.90%
=== RUN   TestHitrate/DirectCache(custom_policy)
    benches_test.go:175: hits: 780808	misses: 219192	hitrate: 78.08%
=== RUN   TestHitrate/FreeCache
    benches_test.go:175: hits: 727308	misses: 272692	hitrate: 72.73%
=== RUN   TestHitrate/FastCache
    benches_test.go:175: hits: 696096	misses: 303904	hitrate: 69.61%
=== RUN   TestHitrate/BigCache
    benches_test.go:175: hits: 697831	misses: 302169	hitrate: 69.78%
--- PASS: TestHitrate (3.43s)
    --- PASS: TestHitrate/DirectCache (0.58s)
    --- PASS: TestHitrate/DirectCache(custom_policy) (0.76s)
    --- PASS: TestHitrate/FreeCache (0.67s)
    --- PASS: TestHitrate/FastCache (0.72s)
    --- PASS: TestHitrate/BigCache (0.71s)
PASS
ok  	benches	3.439s
```

## License

The MIT License
