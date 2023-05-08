package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hashicorp/consul/api"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/qml-123/GateWay/common"
	"github.com/qml-123/GateWay/http"
)

const (
	configPath = "config/services.json"
)

func main() {
	conf, err := common.GetJsonFromFile(configPath)
	if err != nil {
		panic(err)
	}
	server := http.NewServer(conf, conf.ListenPort)
	addr, _ := net.ResolveTCPAddr("tcp", conf.ListenIp+":"+fmt.Sprintf("%d", conf.ListenPort))
	if err = initConsul(addr, conf); err != nil {
		panic(err)
	}

	defer closeConsul(addr, conf)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func initConsul(addr net.Addr, conf *common.Conf) error {
	//r, err := consul.NewConsulRegister("127.0.0.1:8500")
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: conf.ConsulAddRess,
		Scheme:  "http",
	})
	if err != nil {
		log.Println("NewConsulRegisterWithConfig failed")
		return err
	}
	if err = r.Register(&registry.Info{
		ServiceName: conf.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		log.Println("Register failed")
		return err
	}
	return nil
}

func closeConsul(addr net.Addr, conf *common.Conf) {
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: conf.ConsulAddRess,
		Scheme:  "http",
	})
	if err != nil {
		log.Println("NewConsulRegisterWithConfig failed")
		return
	}
	if err = r.Deregister(&registry.Info{
		ServiceName: conf.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		log.Println("Deregister failed")
	}
}
