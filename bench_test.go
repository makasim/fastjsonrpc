package fastjsonrpc_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"

	"github.com/makasim/fastjsonrpc"
)

type FooService struct {
}

func (t *FooService) Foo(_ interface{}, res *map[string]interface{}) error {
	res1 := *res
	res1["a"] = "a"
	res1["b"] = "b"
	res1["c"] = "c"
	res1["d"] = "d"
	res1["e"] = "e"
	res1["f"] = "f"
	res1["g"] = "g"
	return nil
}

func (t *FooService) FooRaw(_ json.RawMessage, res *json.RawMessage) error {
	*res = append(*res, `{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e", "f": "f", "g": "g"}`...)
	return nil
}

func BenchmarkServerFoo(bench *testing.B) {
	s := rpc.NewServer()
	if err := s.Register(&FooService{}); err != nil {
		log.Fatalln(err)
	}
	conn1, conn2 := net.Pipe()

	go func() {
		s.ServeCodec(jsonrpc.NewServerCodec(conn1))
	}()

	go func() {
		io.Copy(ioutil.Discard, conn2)
	}()

	msg := []byte(`{"method": "FooService.Foo","id": 123,"params": [{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e", "f": "f", "g": "g"}]}
`)

	bench.ResetTimer()
	bench.ReportAllocs()
	for i := 0; i < bench.N; i++ {
		if _, err := conn2.Write(msg); err != nil {
			log.Fatalln(err)
		}
	}
}

func BenchmarkServerFooRaw(bench *testing.B) {
	s := rpc.NewServer()
	if err := s.Register(&FooService{}); err != nil {
		log.Fatalln(err)
	}
	conn1, conn2 := net.Pipe()

	go func() {
		s.ServeCodec(jsonrpc.NewServerCodec(conn1))
	}()

	go func() {
		io.Copy(ioutil.Discard, conn2)
	}()

	msg := []byte(`{"method": "FooService.FooRaw","id": 123,"params": [{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e", "f": "f", "g": "g"}]}
`)

	bench.ResetTimer()
	bench.ReportAllocs()
	for i := 0; i < bench.N; i++ {
		if _, err := conn2.Write(msg); err != nil {
			log.Fatalln(err)
		}
	}
}

func BenchmarkServerFooRawFastCodec(bench *testing.B) {
	s := rpc.NewServer()
	if err := s.Register(&FooService{}); err != nil {
		log.Fatalln(err)
	}
	conn1, conn2 := net.Pipe()

	go func() {
		s.ServeCodec(fastjsonrpc.NewServerCodec(conn1))
	}()

	go func() {
		io.Copy(ioutil.Discard, conn2)
	}()

	msg := []byte(`{"method": "FooService.Foo","id": 123,"params": [{"a": "a", "b": "b", "c": "c", "d": "d", "e": "e", "f": "f", "g": "g"}]}
`)

	bench.ResetTimer()
	bench.ReportAllocs()
	for i := 0; i < bench.N; i++ {
		if _, err := conn2.Write(msg); err != nil {
			log.Fatalln(err)
		}
	}
}
