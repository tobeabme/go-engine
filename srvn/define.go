// Package srvn provides an consul-backed and etcd-backed gRPC resolver for discovering gRPC services.
// srvn means service.naming

package srvn

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

type Provider int

const (
	Etcd Provider = iota
	Consul
)

type Config struct {
	Provider  Provider
	Endpoints []string
}

type Service interface {
	Registry
	Discovery
}

// Registry interface for extend
type Registry interface {
	Register(serviceName string, addr string) error
	DeRegister(string) error
}

type Discovery interface {
	Dial(serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	Resolver(serviceName string) (naming.Resolver, error)
}
