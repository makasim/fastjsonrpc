package fastjsonrpc

import (
	"encoding/json"
	"sync"
)

var serverRequestPool = &sync.Pool{}

func acquireServerRequest() *serverRequest {
	req := serverRequestPool.Get()
	if req == nil {
		return &serverRequest{
			Params: &json.RawMessage{},
			Id:     &json.RawMessage{},
		}
	}

	return req.(*serverRequest)
}

func releaseServerRequest(req *serverRequest) {
	req.reset()
	serverRequestPool.Put(req)
}
