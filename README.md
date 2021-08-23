# fastjsonrpc

Fast JSON RPC 1.0 codec for Go's net\rpc server.
This is a fork of [net/rpc/jsonrpc](https://pkg.go.dev/net/rpc/jsonrpc@go1.17)

It works only with function signatures like: 

```
func (t *FooService) FooRaw(_ json.RawMessage, res *json.RawMessage) error {
```

Usage:
```
var conn *net.Conn

s := rpc.NewServer()
s.ServeCodec(fastjsonrpc.NewServerCodec(conn))
```

## Benchmarks

```
$go test -run=none -bench=. ./...
goos: darwin
goarch: amd64
pkg: github.com/makasim/fastjsonrpc
cpu: Intel(R) Core(TM) i7-4980HQ CPU @ 2.80GHz
BenchmarkServerFoo-8               	  141896	      7726 ns/op	    2207 B/op	      45 allocs/op
BenchmarkServerFooRaw-8            	  162300	      6758 ns/op	     745 B/op	      17 allocs/op
BenchmarkServerFooRawFastCodec-8   	  191630	      5577 ns/op	     145 B/op	       6 allocs/op
PASS
ok  	github.com/makasim/fastjsonrpc	3.899s
```