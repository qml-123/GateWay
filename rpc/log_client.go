package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
	"github.com/qml-123/GateWay/error_code"
	"github.com/qml-123/GateWay/kitex_gen/es_log"
	"github.com/qml-123/GateWay/kitex_gen/es_log/logservice"
	"github.com/qml-123/GateWay/model"
)

var(
	log_service_client logservice.Client
	m map[string]func(c context.Context, ctx *app.RequestContext)
)

func initLogClient() error {
	m = make(map[string]func(c context.Context, ctx *app.RequestContext))
	option := client.WithResolver(r)
	var err error
	log_service_client, err = logservice.NewClient(model.LogServiceName, option)
	if err != nil {
		log.Println("NewClient log filed")
		return err
	}
	m["Search"] = SearchHandler
	return nil
}

func getLogMethodHandler(method string) (func(c context.Context, ctx *app.RequestContext), error) {
	if _, ok := m[method]; !ok {
		return nil, fmt.Errorf("no method")
	}
	return m[method], nil
}

func SearchHandler(c context.Context, ctx *app.RequestContext) {
	req := es_log.NewSearchRequest()
	if err := json.Unmarshal(ctx.Request.Body(), req); err != nil {
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	resp, err := log_service_client.Search(c, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": 0,
		"message": "success",
		"data": resp,
	})
}
