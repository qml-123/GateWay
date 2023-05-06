package rpc

import (
	"fmt"
	"log"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/qml-123/GateWay/kitex_gen/es_log/logservice"
	"github.com/qml-123/GateWay/model"
)

var (
	r          discovery.Resolver
	app_client client.Client
	log_client client.Client
	m          map[string]client.Client
)

func InitClient(conf *model.Conf) (err error) {
	// init consul
	err = initConsulClient(conf.GetConsulAddRess())
	if err != nil {
		return
	}

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

func initKitexClients() (err error) {
	option := client.WithResolver(r)
	m = make(map[string]client.Client)

	{
		log_client, err = client.NewClient(logservice.NewServiceInfo(), option, client.WithDestService(model.LogServiceName))
		if err != nil {
			log.Println("NewClient log filed")
			return
		}
		m[model.LogServiceName] = log_client
	}

	{
		//app_client, err = client.NewClient(appservice.NewServiceInfo(), option, client.WithDestService(model.LogServiceName))
		//if err != nil {
		//	log.Println("NewClient app filed")
		//	return
		//}
		//m[model.AppServiceName] = app_client
	}

	return nil
}

func GetClient(serviceName string) (client.Client, error) {
	if _, ok := m[serviceName]; !ok {
		return nil, fmt.Errorf("no service: %s", serviceName)
	}
	return m[serviceName], nil
}
