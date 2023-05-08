package main

import (
	"fmt"
	"log"
	"net"

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
	if err = common.InitConsul(addr, conf); err != nil {
		panic(err)
	}

	defer common.CloseConsul(addr, conf)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
