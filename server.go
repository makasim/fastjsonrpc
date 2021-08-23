package fastjsonrpc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"sync"

	"github.com/valyala/fastjson"
)

var errMissingParams = errors.New("jsonrpc: request body missing params")

type serverCodec struct {
	enc *json.Encoder
	c   io.Closer

	scan *bufio.Scanner
	pp   *fastjson.ParserPool

	mutex sync.Mutex // protects seq, pending
	seq   uint64

	req     *serverRequest
	pending map[uint64]*serverRequest
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		enc:     json.NewEncoder(conn),
		scan:    bufio.NewScanner(conn),
		pp:      &fastjson.ParserPool{},
		c:       conn,
		pending: make(map[uint64]*serverRequest),
	}
}

type serverRequest struct {
	Method string
	Params *json.RawMessage
	Id     *json.RawMessage
}

func (r *serverRequest) reset() {
	r.Method = ""
	*r.Params = (*r.Params)[:0]
	*r.Id = (*r.Id)[:0]
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	if !c.scan.Scan() {
		if err := c.scan.Err(); err != nil {
			return err
		}
		return fmt.Errorf("scan false")
	}

	p := c.pp.Get()
	defer c.pp.Put(p)
	v, err := p.ParseBytes(c.scan.Bytes())
	if err != nil {
		return fmt.Errorf("read request header: %s", err)
	}

	c.req = acquireServerRequest()
	c.req.Method = string(v.GetStringBytes("method"))

	paramsV := v.Get("params", "0")
	if paramsV.Exists() {
		*c.req.Params = paramsV.MarshalTo(*c.req.Params)
	}
	idV := v.Get("id")
	if idV.Exists() {
		*c.req.Id = idV.MarshalTo(*c.req.Id)
	}

	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.req
	r.ServiceMethod = c.req.Method
	r.Seq = c.seq
	c.mutex.Unlock()
	return nil
}

func (c *serverCodec) ReadRequestBody(x interface{}) error {
	defer func() {
		c.req = nil
	}()

	if x == nil {
		return nil
	}

	if len(*c.req.Params) == 0 {
		return errMissingParams
	}

	x1, ok := x.(*json.RawMessage)
	if !ok {
		return fmt.Errorf("params must be *json.RawMessage")
	}

	*x1 = *c.req.Params
	return nil
}

var null = json.RawMessage([]byte("null"))

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	c.mutex.Lock()
	req, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	defer releaseServerRequest(req)

	if len(*req.Id) == 0 {
		req.Id = &null
	}

	resp := serverResponse{Id: req.Id}
	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = r.Error
	}

	return c.enc.Encode(resp)
}

func (c *serverCodec) Close() error {
	return c.c.Close()
}

func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}
