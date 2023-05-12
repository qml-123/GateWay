package http

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/qml-123/GateWay/common"
	"github.com/qml-123/GateWay/pkg/rpc"
)

func NewServer(conf *common.Conf, port int) *server.Hertz {
	srv := server.Default(server.WithHostPorts(":" + fmt.Sprintf("%d", port)))
	srv.Use(CorsMiddleware())

	var err error
	if err = rpc.InitClient(conf); err != nil {
		panic(err)
	}

	if err = registerFunc(srv, conf); err != nil {
		panic(err)
	}

	return srv
}

type rpcRequest struct {
	Params json.RawMessage `json:"params"`
}

func registerFunc(srv *server.Hertz, conf *common.Conf) error {
	for _, api := range conf.Api {
		service_name := api.Name

		for _, method := range api.Methods {
			rpc_func_name := method.RpcFunction
			http_method := method.HttpMethod
			http_path := method.HttpPath

			//reqType, respType, err := getReqAndRespTypes(_client.GetClient(), rpc_func_name)
			//if err != nil {
			//	return err
			//}
			//logger.Infof("reqType: %v, respType: %v", reqType, respType)

			if http_method == "POST" {
				f, err := rpc.GetHandler(service_name, rpc_func_name)
				if err != nil {
					return err
				}
				srv.POST(http_path, f)
			} else if http_path == "Get" {
				srv.GET(http_path, func(c context.Context, ctx *app.RequestContext) {})
			} else {
				return fmt.Errorf("not suport mehod: %s", http_method)
			}
		}
	}
	return nil
}
