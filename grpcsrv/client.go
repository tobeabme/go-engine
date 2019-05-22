package grpcsrv

import (
	"fmt"
	"time"

	"github.com/tobeabme/go-engine/srvn"
	"google.golang.org/grpc"
)

type Client struct {
	Dialopts []grpc.DialOption
	ns       srvn.Service
}

func NewClient(provider srvn.Provider, endpoints ...string) (*Client, error) {
	cfg := srvn.Config{Provider: provider, Endpoints: endpoints}
	ns, err := srvn.NewDiscover(cfg)
	if err != nil {
		return nil, err
	}

	c := &Client{ns: ns}

	return c, err
}

// Dial grpc server
func (c *Client) Dial(serviceName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {

	r, err := c.ns.Resolver(serviceName)
	if err != nil {
		return nil, err
	}

	c.Dialopts = append(c.Dialopts, grpc.WithInsecure())
	c.Dialopts = append(c.Dialopts, grpc.WithBlock())
	c.Dialopts = append(c.Dialopts, grpc.WithTimeout(10*time.Second))
	c.Dialopts = append(c.Dialopts, grpc.WithUnaryInterceptor(UnaryLoggingClientInterceptor))

	c.Dialopts = append(c.Dialopts, grpc.WithBalancer(grpc.RoundRobin(r)))
	c.Dialopts = append(c.Dialopts, opts...)

	conn, err := grpc.Dial(serviceName, c.Dialopts...)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial %s: %v", serviceName, err)
	}
	return conn, nil
}
