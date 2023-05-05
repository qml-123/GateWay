package main

import (
	"context"
	app "github.com/qml-123/GateWay/kitex_gen/app"
	es_log "github.com/qml-123/GateWay/kitex_gen/es_log"
)

// LogServiceImpl implements the last service interface defined in the IDL.
type LogServiceImpl struct{}

// Search implements the LogServiceImpl interface.
func (s *LogServiceImpl) Search(ctx context.Context, req *es_log.SearchRequest) (resp *es_log.SearchResponse, err error) {
	// TODO: Your code here...
	return
}

// Ping implements the AppServiceImpl interface.
func (s *AppServiceImpl) Ping(ctx context.Context, req *app.PingRequest) (resp *app.PingResponse, err error) {
	// TODO: Your code here...
	return
}
