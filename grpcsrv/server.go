package grpcsrv

import (
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/tobeabme/go-engine/srvn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	serviceName string
	addr        string
	grpcServer  *grpc.Server
}

type ServiceRegister func(s *grpc.Server)

func NewServer(serviceName, addr string) *Server {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.RPCCompressor(grpc.NewGZIPCompressor()))
	opts = append(opts, grpc.RPCDecompressor(grpc.NewGZIPDecompressor()))
	opts = append(opts, grpc_middleware.WithUnaryServerChain(
		UnaryRecoveryServerInterceptor,
		UnaryLoggingServerInterceptor,
	))

	// opts = append(opts, grpc.UnaryInterceptor(unaryServerInterceptor))
	srv := grpc.NewServer(opts...)
	return &Server{
		serviceName: serviceName,
		addr:        addr,
		grpcServer:  srv,
	}
}

// RegisterServiceToGrpc registers a service and its implementation to the gRPC
func (s *Server) RegisterServiceToGrpc(register ServiceRegister) {
	register(s.grpcServer)
}

// Register registers a service to consul or etcd
// Usage:
//		s.Register(srvn.Consul, "127.0.0.1:8500")
func (s *Server) Register(provider srvn.Provider, endpoints ...string) error {
	cfg := srvn.Config{Provider: provider, Endpoints: endpoints}

	srv, err := srvn.NewRegister(cfg)
	if err != nil {
		return err
	}

	err = srv.Register(s.serviceName, s.addr)
	if err != nil {
		log.Printf("%s is failed to register the %s to %s.", s.serviceName, s.addr, provider.String())
		return err
	}

	// log.Printf("%s is successful to register the %s to %s.", s.serviceName, s.addr, provider.String())

	return err
}

func (s *Server) StartUp() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Printf(fmt.Sprintf("RPC Server has been started up. %s", s.addr))

	reflection.Register(s.grpcServer)

	return s.grpcServer.Serve(lis)
}
