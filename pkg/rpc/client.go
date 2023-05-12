package rpc

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/qml-123/GateWay/common"
	"github.com/qml-123/GateWay/model"
)

type BaseClient struct {
	_c     client.Client
	c      interface{}
	Method map[string]func() interface{}
}

var (
	r       discovery.Resolver
	clients map[string]baseClient
)

type baseClient interface {
	initClient() error
	getHandler(method string) (func(c context.Context, ctx *app.RequestContext), error)
}

func InitClient(conf *common.Conf) (err error) {
	// init consul
	err = initConsulClient(conf.ConsulAddRess)
	if err != nil {
		return
	}

	clients = make(map[string]baseClient)
	return initKitexClients()
}

func initConsulClient(addr string) (err error) {
	r, err = consul.NewConsulResolver(addr)
	if err != nil {
		log.Printf("NewConsulResolver failed")
		return
	}
	return nil
}

func initKitexClients() error {

	{ // log
		clients[model.LogServiceName] = newlogClient()
		clients[model.LogServiceName].initClient()
	}

	{ // app
		clients[model.AppServiceName] = newappClient()
		clients[model.AppServiceName].initClient()
	}

	return nil
}

func GetHandler(service, method string) (func(c context.Context, ctx *app.RequestContext), error) {
	if _, ok := clients[service]; !ok {
		return nil, fmt.Errorf("no service")
	}

	return clients[service].getHandler(method)
}
