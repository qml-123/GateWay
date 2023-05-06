package common

import (
	"encoding/json"
	"os"
)


type Methods struct {
	RpcFunction string `json:"rpc_function"`
	HttpMethod  string `json:"http_method"`
	HttpPath    string `json:"http_path"`
}

type Api struct {
	Name    string     `json:"name"`
	Methods []*Methods `json:"methods"`
}

type Conf struct {
	ServiceName   string `json:"service_name"`
	ListenPort    int    `json:"listen_port"`
	ListenIp      string `json:"listen_ip"`
	ConsulAddRess string `json:"consul_address"`
	Api           []*Api `json:"api"`
}

func GetJsonFromFile(filePath string) (*Conf, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取 JSON 数据
	decoder := json.NewDecoder(file)
	var conf *Conf
	err = decoder.Decode(&conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
