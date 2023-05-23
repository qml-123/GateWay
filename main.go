package main

import (
	"context"
	"fmt"
	"net"

	"github.com/qml-123/GateWay/pkg/http"
	"github.com/qml-123/GateWay/pkg/id"
	"github.com/qml-123/GateWay/pkg/log"
	"github.com/qml-123/GateWay/pkg/redis"
	"github.com/qml-123/app_log/common"
	"github.com/qml-123/app_log/logger"
)

const (
	configPath = "config/services.json"
)

func main() {
	ctx := context.Background()
	conf, err := common.GetJsonFromFile(configPath)
	if err != nil {
		panic(err)
	}
	// log
	if err = log.InitLogger(conf.EsUrl); err != nil {
		panic(err)
	}

	// redis
	if err = redis.InitRedis(); err != nil {
		panic(err)
	}

	// init id_gen
	if err = id.InitGen(); err != nil {
		panic(err)
	}

	server := http.NewServer(conf, conf.ListenPort)
	addr, _ := net.ResolveTCPAddr("tcp", conf.ListenIp+":"+fmt.Sprintf("%d", conf.ListenPort))
	if err = common.InitConsul(addr, conf); err != nil {
		panic(err)
	}

	defer common.CloseConsul(addr, conf)
	logger.Info(ctx, "init success")
	if err := server.Run(); err != nil {
		logger.Warn(ctx, "Failed to run server: %v", err)
	}
}
