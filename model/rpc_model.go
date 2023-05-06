package model

import "context"

type RPCService struct {
	Name    string      `json:"name"`
	Methods []RPCMethod `json:"methods"`
}

type RPCMethod struct {
	Name       string `json:"name"`
	HTTPMethod string `json:"http_method"`
	HTTPPath   string `json:"http_path"`
	NewRequest func() (interface{}, error)
	Call       func(ctx context.Context, request interface{}) (interface{}, error)
}
