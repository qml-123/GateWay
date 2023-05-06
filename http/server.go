package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/GateWay/rpc"
)

func NewServer(conf *model.Conf, port int) *server.Hertz {
	srv := server.Default(server.WithHostPorts(":" + fmt.Sprintf("%d", port)))

	var err error
	if err = rpc.InitClient(conf); err != nil {
		panic(err)
	}

	if err = registerFunc(srv, conf); err != nil {
		panic(err)
	}

	return srv
}

func registerFunc(srv *server.Hertz, conf *model.Conf) error {
	for _, api := range conf.GetApi() {
		service_name := api.GetName()
		client, err := rpc.GetClient(service_name)
		if err != nil {
			return err
		}

		for _, method := range api.GetMethods() {
			rpc_func_name := method.GetRpcFunction()
			http_method := method.GetHttpMethod()
			http_path := method.GetHttpPath()

			if http_method == "Post" {
				srv.POST(http_path, func(c context.Context, ctx *app.RequestContext) {
					var req interface{}
					err = json.NewDecoder(ctx.GetRequest().BodyStream()).Decode(req)
					if err != nil {
						ctx.AbortWithStatus(http.StatusInternalServerError)
						return
					}
					var resp interface{}
					rpcCtx := context.Background()
					err = client.Call(rpcCtx, rpc_func_name, req, resp)
					if err != nil {
						ctx.AbortWithStatus(http.StatusInternalServerError)
						return
					}
					logger.Info("http_resp: %v", resp)
					ctx.JSON(http.StatusOK, resp)
				})
			} else if http_path == "Get" {
				srv.GET(http_path, func(c context.Context, ctx *app.RequestContext) {

				})
			} else {
				return fmt.Errorf("not suport mehod: %s", http_method)
			}
		}
	}
	return nil
}
