package rpc

import (
	"errors"
	"jiacrontab/libs"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

var (
	defaultRPCPath      string
	defaultRPCDebugPath string
	listenAddr          string
)

type MrpcClient struct {
	client *rpc.Client
}

func (c *MrpcClient) Call(serviceMethod string, args, reply interface{}) error {
	var err error
	if c.client == nil {
		err = errors.New("rpcClient failed initialize")
		log.Println(err)
		return err

	}

	if err = c.client.Call(serviceMethod, args, reply); err != nil {
		log.Printf("Rpc Call %s %s", serviceMethod, err)
	}
	c.client.Close()
	c.client = nil
	return err
}

func NewRpcClient(addr string) (*MrpcClient, error) {

	c, err := libs.DialHTTP("tcp", addr, defaultRPCPath)

	if err != nil {
		return nil, err
	}

	m := &MrpcClient{
		client: c,
	}

	return m, nil
}

func InitSrvRpc(rpcPath string, rpcDebugPath string, addr string, srcvr ...interface{}) {

	var err error
	defaultRPCPath = rpcPath
	defaultRPCDebugPath = rpcDebugPath
	listenAddr = addr

	server := rpc.NewServer()
	for _, v := range srcvr {
		if err = server.Register(v); err != nil {
			panic(err)
		}
	}
	server.HandleHTTP(defaultRPCPath, defaultRPCDebugPath)
	l, err := net.Listen("tcp", listenAddr)

	if err != nil {
		panic(err)
	}
	log.Printf("rpc listen %s", listenAddr)

	if err := http.Serve(l, nil); err != nil {
		panic(err)
	}
}
