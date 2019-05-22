// Package srvn provides an consul-backed and etcd-backed gRPC resolver for discovering gRPC services.
// srvn means service.naming

package srvn

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
	// lb "github.com/olivere/grpc/lb/consul"
	"code.qschou.com/peduli/go_common/naming/consul/grpclb"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

type consul struct {
	client   *api.Client
	dialopts []grpc.DialOption
}

// NewConsulRegistry returns a registryClient interface for given consul address
func NewConsul(c Config) (Service, error) {
	c.Endpoints = ConsulEndpointCheck(c.Endpoints)
	addr := c.Endpoints[0]

	cfg := api.DefaultConfig()
	cfg.Address = addr
	cli, err := api.NewClient(cfg)
	if err != nil {
		logrus.Errorf("Can't connect to consul server at %s", addr)
		return nil, err
	}
	return consul{client: cli}, nil
}

func (c consul) Register(serviceName string, addr string) error {
	if serviceName == "" {
		return errors.New("srvn.Register: no service name provided.")
	}

	if addr == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return fmt.Errorf("unable to determine local addr: %v", err)
		}
		defer conn.Close()

		localAddr := conn.LocalAddr().(*net.UDPAddr)
		addr = localAddr.IP.String()
		logrus.Warnf("srvn.Register: no addr provided.")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("Failed to split service addr to host and port: %v", err)
	}

	port_int, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("Failed to convert service's port from string to int: %v", err)
	}

	asr := &api.AgentServiceRegistration{
		ID:      host + ":" + port,
		Name:    serviceName,
		Port:    port_int,
		Address: host,
		Check: &api.AgentServiceCheck{
			// HTTP: fmt.Sprintf("http://%s:%d%s", host, port_int, "/health"),
			// GRPC: fmt.Sprintf("%v:%v/%v", host, port_int, serviceName),
			TCP: fmt.Sprintf("%s:%d", host, port_int),
			// TTL:                            (5 * time.Second).String(),
			Timeout:                        "2s",
			Interval:                       "3s",
			DeregisterCriticalServiceAfter: "15s", //delete the service after 10s when check failed
		},
		// EnableTagOverride: false,
		// Tags:              tags,
	}
	err = c.client.Agent().ServiceRegister(asr)
	if err != nil {
		logrus.Errorf("Failed to register service at '%s'. error: %v", addr, err)
	} else {
		//logrus.Infof("Regsitered service '%s' at consul.", serviceName)
	}

	return err
}

func (c consul) DeRegister(serviceName string) error {
	err := c.client.Agent().ServiceDeregister(serviceName)

	if err != nil {
		logrus.Errorf("Failed to deregister service by id: '%s'. Error: %v", serviceName, err)
	} else {
		logrus.Infof("Deregistered service '%s' at consul.", serviceName)
	}

	return err
}

// Dial grpc server
func (c consul) Dial(serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {

	r := grpclb.NewResolver(c.client, serviceName, "")

	c.dialopts = append(c.dialopts, grpc.WithInsecure())
	c.dialopts = append(c.dialopts, grpc.WithBlock())
	c.dialopts = append(c.dialopts, grpc.WithTimeout(10*time.Second))

	c.dialopts = append(c.dialopts, grpc.WithBalancer(grpc.RoundRobin(r)))
	c.dialopts = append(c.dialopts, opts...)

	conn, err := grpc.Dial(serviceName, c.dialopts...)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial %s: %v", serviceName, err)
	}
	return conn, nil
}

func (c consul) Resolver(serviceName string) (naming.Resolver, error) {
	r := grpclb.NewResolver(c.client, serviceName, "")
	return r, nil
}
