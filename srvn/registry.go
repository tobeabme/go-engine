package srvn

import (
	"fmt"
)

func NewRegister(cfg Config) (Service, error) {
	switch cfg.Provider {
	// case Etcd:
	// 	return NewEtcd(cfg)
	case Consul:
		return NewConsul(cfg)
	default:
		return nil, fmt.Errorf("Unsupported registry provider: %v", cfg.Provider)
	}
}

func (p Provider) String() string {
	switch p {
	case Etcd:
		return "Etcd"
	case Consul:
		return "Consul"
	default:
		return "UNKNOWN"
	}
}
