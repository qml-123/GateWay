package id

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/qml-123/GateWay/pkg/redis"
)

const (
	GenerateIDLock = "generate_id_lock"
	GenerateIDKey  = "generate_id_key"
)

var (
	_node *snowflake.Node
)

func InitGen() error {
	var id int64
	for {
		success, err := redis.SetNX(GenerateIDLock, 1, 30*time.Second)
		if err == nil && success {
			id, err = redis.Incr(GenerateIDKey)
			if err == nil {
				redis.Del(GenerateIDLock)
				break
			}
			redis.Del(GenerateIDLock)
		}

		time.Sleep(2 * time.Second)
	}

	var err error
	_node, err = snowflake.NewNode(id)
	if err != nil {
		return err
	}
	return nil
}

func Generate() snowflake.ID {
	return _node.Generate()
}

func GenerateIDBase58() string {
	return _node.Generate().Base58()
}
