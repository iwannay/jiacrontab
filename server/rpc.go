package main

import (
	"errors"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type mrpcClient struct {
	proto.ClientConf
	client *rpc.Client
}

func (c *mrpcClient) call(serviceMethod string, args, reply interface{}) bool {
	if c.client == nil || c.State == 0 {
		err := errors.New("rpcClient failed initialize")
		log.Println(err)
		return false

	}
	if err := c.client.Call(serviceMethod, args, reply); err != nil {
		c.State = 0
		c.client.Close()
		globalStore.Update(nil)
		log.Printf("Rpc Call %s %s", serviceMethod, err)
		return false
	}
	return true
}

func newRpcClient(addr string) (*mrpcClient, error) {
	defer globalStore.Update(nil)
	globalStore.lock.Lock()
	if v, ok := globalStore.RpcClientList[addr]; ok {
		globalStore.lock.Unlock()
		if v.State == 1 && v.client != nil {
			return v, nil
		} else {
			c, err := libs.DialHTTP("tcp", addr, globalConfig.defaultRPCPath)
			v.client = c
			v.Addr = addr
			if err != nil {
				log.Println(err)
				v.State = 0
			} else {
				v.State = 1
			}
			return v, err
		}

	} else {
		globalStore.lock.Unlock()
		c, err := libs.DialHTTP("tcp", addr, globalConfig.defaultRPCPath)
		tmp := &mrpcClient{
			client: c,
		}

		tmp.Addr = addr
		if err != nil {
			log.Println("dialing:", err)
			tmp.State = 0
		} else {
			tmp.State = 1
		}
		globalStore.lock.Lock()
		globalStore.RpcClientList[addr] = tmp
		globalStore.lock.Unlock()

		return tmp, err
	}
}

func initSrvRpc(srcvr ...interface{}) {
	var err error
	server := rpc.NewServer()
	for _, v := range srcvr {
		if err = server.Register(v); err != nil {
			panic(err)
		}
	}
	server.HandleHTTP(globalConfig.defaultRPCPath, globalConfig.defaultRPCDebugPath)
	l, err := net.Listen("tcp", globalConfig.rpcAddr)

	if err != nil {
		panic(err)
	}
	log.Printf("rpc listen %s", globalConfig.rpcAddr)

	if err := http.Serve(l, nil); err != nil {
		panic(err)
	}
}
