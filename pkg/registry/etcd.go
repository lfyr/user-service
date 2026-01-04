package registry

import (
	"go-micro.dev/v4/registry"
	"time"

	"github.com/go-micro/plugins/v4/registry/etcd"
	"user-service/config"
)

// InitEtcdRegistry 初始化Etcd注册表
func InitEtcdRegistry(cfg *config.EtcdConfig) registry.Registry {
	return etcd.NewRegistry(
		registry.Addrs(cfg.Adders...),
		registry.Timeout(5*time.Second),
	)
}
