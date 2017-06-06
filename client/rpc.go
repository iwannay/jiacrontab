package main

import (
	"errors"
	"jiacrontab/client/store"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

var rpcClient *rpc.Client

const retryPeriod = 10 * time.Second
const heartbeatPeriod = 600 * time.Second

func initSrvRpc(srcvr ...interface{}) error {
	var err error
	server := rpc.NewServer()
	for _, v := range srcvr {

		if err = server.Register(v); err != nil {
			return err
		}
	}
	server.HandleHTTP(globalConfig.defaultRPCPath, globalConfig.defaultRPCDebugPath)

	l, err := net.Listen("tcp", globalConfig.rpcListenAddr)
	if err != nil {
		return err
	}
	log.Printf("rpc listen %s", globalConfig.rpcListenAddr)

	return http.Serve(l, nil)
}

func initClientRpc() {
	var err error
	var mail proto.MailArgs

	rpcClient, err = libs.DialHTTP("tcp", globalConfig.rpcSrvAddr, globalConfig.defaultRPCPath)
	if err != nil {
		log.Printf("Rpc:%v", err)
		time.AfterFunc(retryPeriod, initClientRpc)
	} else {
		if err := rpcCall("Logic.Register", proto.ClientConf{
			Addr:  globalConfig.addr,
			State: 1,
			Mail:  globalConfig.mailTo,
		}, &mail); err != nil {
			if rpcClient != nil {
				rpcClient.Close()
			}
			time.AfterFunc(retryPeriod, initClientRpc)
		}

		globalStore.Update(func(s *store.Store) {
			s.Mail = mail
		}).Sync()

	}

}

func pingRpcSrv() {
	var mail proto.MailArgs
	time.AfterFunc(heartbeatPeriod, func() {

		err := rpcCall("Logic.Register", proto.ClientConf{
			Addr:  globalConfig.addr,
			State: 1,
		}, &mail)

		globalStore.Update(func(s *store.Store) {
			s.Mail = mail
		}).Sync()

		if err != nil {
			if rpcClient != nil {
				rpcClient.Close()
			}
			rpcClient, _ = libs.DialHTTP("tcp", globalConfig.rpcSrvAddr, globalConfig.defaultRPCPath)
		}

		pingRpcSrv()
	})
}

func rpcCall(serviceMethod string, args, reply interface{}) error {
	if rpcClient == nil {
		return errors.New("RpcClient failed initialize")
	}
	err := rpcClient.Call(serviceMethod, args, reply)
	log.Println("RpcCall", serviceMethod, err)
	return err
}
