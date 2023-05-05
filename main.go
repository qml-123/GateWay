package main

import (
	"log"

	es_log "github.com/qml-123/GateWay/kitex_gen/es_log/logservice"
)

func main() {
	svr := es_log.NewServer(new(LogServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
