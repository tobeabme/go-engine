package srvn

import (
	"fmt"
)

func NewDiscover(cfg Config) (Service, error) {
	switch cfg.Provider {
	case Consul:
		return NewConsul(cfg)
	default:
		return nil, fmt.Errorf("Unsupported registry provider: %v", cfg.Provider)
	}
}
