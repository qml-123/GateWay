package common

import (
	"encoding/json"
	"os"

	"github.com/qml-123/GateWay/model"
)

func GetJsonFromFile(filePath string) (*model.Conf, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取 JSON 数据
	decoder := json.NewDecoder(file)
	var conf *model.Conf
	err = decoder.Decode(&conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
