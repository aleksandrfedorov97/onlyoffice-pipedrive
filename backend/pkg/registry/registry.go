package registry

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/go-micro/plugins/v4/registry/kubernetes"
	"github.com/go-micro/plugins/v4/registry/mdns"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/registry/cache"
)

// NewRegistry looks up envs and configures respective registries based on those variables. Defaults to memory
func NewRegistry(config *config.RegistryConfig) registry.Registry {
	var r registry.Registry
	switch config.Registry.RegistryType {
	case 1:
		r = kubernetes.NewRegistry(
			registry.Addrs(config.Registry.Addresses...),
		)
	case 2:
		r = consul.NewRegistry(
			registry.Addrs(config.Registry.Addresses...),
		)
	case 3:
		r = etcd.NewRegistry(
			registry.Addrs(config.Registry.Addresses...),
		)
	case 4:
		r = mdns.NewRegistry(
			registry.Addrs(config.Registry.Addresses...),
		)
	default:
		r = mdns.NewRegistry(
			registry.Addrs(config.Registry.Addresses...),
		)
	}

	return cache.New(r, cache.WithTTL(config.Registry.CacheTTL))
}
