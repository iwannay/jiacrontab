package rpc

import (
	"context"
	"net/rpc"
	"sync"

	"github.com/iwannay/log"
)

var (
	defaultClients  *clients
	PingService     = "Srv.Ping"
	RegisterService = "Srv.Register"
)

type clients struct {
	lock    sync.RWMutex
	clients map[string]*Client
}

func (c *clients) get(addr string) *Client {
	var (
		cli *Client
		ok  bool
		op  ClientOptions
	)

	c.lock.Lock()
	defer c.lock.Unlock()
	if cli, ok = c.clients[addr]; ok {
		return cli
	}
	op.Network = "tcp4"
	op.Addr = addr
	cli = Dial(op)
	c.clients[addr] = cli
	go cli.Ping(PingService)

	return cli
}

func (c *clients) del(addr string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if cli, ok := c.clients[addr]; ok {
		cli.Close()
	}
	delete(c.clients, addr)
}

func Call(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := defaultClients.get(addr).Call(serviceMethod, context.TODO(), args, reply)
	if err == rpc.ErrShutdown {
		log.Debug("rpc remove", addr)
		Del(addr)
	}
	return err
}

func CallCtx(addr string, serviceMethod string, ctx context.Context, args interface{}, reply interface{}) error {
	err := defaultClients.get(addr).Call(serviceMethod, ctx, args, reply)
	if err == rpc.ErrShutdown {
		log.Debug("rpc remove", addr)
		Del(addr)
	}
	return err
}

func Del(addr string) {
	if defaultClients != nil {
		defaultClients.del(addr)
	}
}

func DelNode(addr string) {
	if defaultClients != nil {
		defaultClients.del(addr)
	}
}

func init() {
	defaultClients = &clients{
		clients: make(map[string]*Client),
	}
}
