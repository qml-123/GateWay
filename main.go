package main

import (
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
	server := http.NewServer(conf, conf.GetListenPort())
	addr, _ := net.ResolveTCPAddr("tcp", conf.)
	if err = initConsul(addr); err != nil {
		panic(err)
	}

	defer closeConsul(addr)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func initConsul(addr net.Addr) error {
	//r, err := consul.NewConsulRegister("127.0.0.1:8500")
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: "114.116.15.130:8500",
		Scheme:  "http",
	})
	if err != nil {
		log.Println("NewConsulRegisterWithConfig failed")
		return err
	}
	if err = r.Register(&registry.Info{
		ServiceName: model.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		log.Println("Register failed")
		return err
	}
	return nil
}

func closeConsul(addr net.Addr) {
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: "114.116.15.130:8500",
		Scheme:  "http",
	})
	if err != nil {
		log.Println("NewConsulRegisterWithConfig failed")
		return
	}
	if err = r.Deregister(&registry.Info{
		ServiceName: model.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		log.Println("Deregister failed")
	}
}
