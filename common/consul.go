package common

import (
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hashicorp/consul/api"
	consul "github.com/kitex-contrib/registry-consul"
)

func initConsul(addr net.Addr, conf *Conf) error {
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: conf.ConsulAddRess,
		Scheme:  "http",
	})
	if err != nil {
		return err
	}
	if err = r.Register(&registry.Info{
		ServiceName: conf.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		return err
	}
	return nil
}

func closeConsul(addr net.Addr, conf *Conf) error {
	r, err := consul.NewConsulRegisterWithConfig(&api.Config{
		Address: conf.ConsulAddRess,
		Scheme:  "http",
	})
	if err != nil {
		return err
	}
	if err = r.Deregister(&registry.Info{
		ServiceName: conf.ServiceName,
		Addr:        addr,
		StartTime:   time.Now(),
		Weight:      1,
	}); err != nil {
		return err
	}
	return nil
}
