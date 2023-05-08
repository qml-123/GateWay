package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
	app_service "github.com/qml-123/AppService/kitex_gen/app"
	"github.com/qml-123/AppService/kitex_gen/app/appservice"
	"github.com/qml-123/GateWay/error_code"
	"github.com/qml-123/GateWay/model"
)

type appClient struct {
	app_service_client appservice.Client
	m                  map[string]func(c context.Context, ctx *app.RequestContext)
}

func newappClient() *appClient {
	return &appClient{
		m: make(map[string]func(c context.Context, ctx *app.RequestContext)),
	}
}

func (cli *appClient) initClient() error {
	option := client.WithResolver(r)
	var err error
	cli.app_service_client, err = appservice.NewClient(model.AppServiceName, option)
	if err != nil {
		log.Println("NewClient log filed")
		return err
	}
	cli.m["Ping"] = cli.PingHandler
	return nil
}

func (cli *appClient) getHandler(method string) (func(c context.Context, ctx *app.RequestContext), error) {
	if _, ok := cli.m[method]; !ok {
		return nil, fmt.Errorf("no method")
	}
	return cli.m[method], nil
}

func (cli *appClient) PingHandler(c context.Context, ctx *app.RequestContext) {
	resp, err := cli.app_service_client.Ping(c, &app_service.PingRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}
