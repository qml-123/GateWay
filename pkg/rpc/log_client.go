package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/app_log/error_code"
	"github.com/qml-123/app_log/logger"
	"github.com/qml-123/es_log/kitex_gen/es_log"
	"github.com/qml-123/es_log/kitex_gen/es_log/logservice"
)

type logClient struct {
	log_service_client logservice.Client
	m                  map[string]func(c context.Context, ctx *app.RequestContext)
}

func newlogClient() *logClient {
	return &logClient{
		m: make(map[string]func(c context.Context, ctx *app.RequestContext)),
	}
}

func (cli *logClient) initClient() error {
	option := client.WithResolver(r)
	var err error
	cli.log_service_client, err = logservice.NewClient(model.LogServiceName, option)
	if err != nil {
		log.Println("NewClient log filed")
		return err
	}
	cli.m["Search"] = cli.SearchHandler
	return nil
}

func (cli *logClient) getHandler(method string) (func(c context.Context, ctx *app.RequestContext), error) {
	if _, ok := cli.m[method]; !ok {
		return nil, fmt.Errorf("no method")
	}
	return cli.m[method], nil
}

func (cli *logClient) SearchHandler(c context.Context, ctx *app.RequestContext) {
	req := es_log.NewSearchRequest()
	if err := json.Unmarshal(ctx.Request.Body(), req); err != nil {
		logger.Warn(c, "[Search] req unmarshal failed, err: %v", err)
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	resp, err := cli.log_service_client.Search(c, req)
	if err != nil {
		logger.Error(c, "[Search] call Search failed, err: %v", err)
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}
