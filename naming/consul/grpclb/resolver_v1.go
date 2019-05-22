package grpclb

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/naming"
)

type gRPCResolver struct {
	c           *api.Client
	service     string
	tag         string
	passingOnly bool
}

// NewResolver initializes and returns a new Resolver.
// It resolves addresses for gRPC connections to the given service and tag.
// If the tag is irrelevant, use an empty string.
func NewResolver(client *api.Client, service, tag string) naming.Resolver {
	r := &gRPCResolver{
		c:           client,
		service:     service,
		tag:         tag,
		passingOnly: true,
	}

	return r
}

// Resolve creates a watcher for target. The watcher interface is implemented
// by Resolver as well, see Next and Close.
func (r *gRPCResolver) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &gRPCWatcher{
		c:      r.c,
		target: target,
		ctx:    ctx,
		cancel: cancel,
		addrs:  map[string]struct{}{},
		r:      r,
	}
	return w, nil
}

type gRPCWatcher struct {
	c         *api.Client
	target    string
	ctx       context.Context
	cancel    context.CancelFunc
	addrs     map[string]struct{}
	lastIndex uint64
	r         *gRPCResolver
}

// Next gets the next set of updates from the etcd resolver.
// Calls to Next should be serialized; concurrent calls are not safe since
// there is no way to reconcile the update ordering.
func (gw *gRPCWatcher) Next() ([]*naming.Update, error) {
	for {
		select {
		case <-gw.ctx.Done():
			err := errors.New("watcher has been closed.")
			logrus.Info("grpclb: received an exit signal,", err)
			return nil, nil
		default:
		}
		services, meta, err := gw.c.Health().Service(gw.target, gw.r.tag, gw.r.passingOnly, &api.QueryOptions{
			WaitIndex: gw.lastIndex,
		})
		if err != nil {
			logrus.Warn("error retrieving instances from Consul: %v", err)
			return nil, err
		}
		gw.lastIndex = meta.LastIndex

		addrs := map[string]struct{}{}
		for _, service := range services {
			saddr := service.Service.Address
			if len(saddr) == 0 {
				saddr = service.Node.Address
			}
			addrs[net.JoinHostPort(saddr, strconv.Itoa(service.Service.Port))] = struct{}{}
		}

		var updates []*naming.Update
		for addr := range gw.addrs {
			if _, ok := addrs[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
			}
		}

		for addr := range addrs {
			if _, ok := gw.addrs[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
			}
		}

		if len(updates) != 0 {
			gw.addrs = addrs
			return updates, nil
		}
	}

}

func (gw *gRPCWatcher) Close() {
	logrus.Info("grpclb: a close action was triggerd by grpc.")
	gw.cancel()
}
