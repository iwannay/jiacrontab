package main

import (
	"jiacrontab/client/store"
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
	"log"
	"time"
)

const heartbeatPeriod = 1 * time.Minute

func RpcHeartBeat() {
	var mail proto.MailArgs

	log.Println("heart beat", globalConfig.rpcSrvAddr, "start")
	err := rpc.Call(globalConfig.rpcSrvAddr, "Logic.Register", proto.ClientConf{
		Addr:  globalConfig.addr,
		State: 1,
		Mail:  globalConfig.mailTo,
	}, &mail)
	log.Println("heart beat", globalConfig.rpcSrvAddr, "end")

	if err != nil {
		log.Println(" heart beat error:", err, "server addr:", globalConfig.rpcSrvAddr)
	}

	globalStore.Update(func(s *store.Store) {
		s.Mail = mail
	}).Sync()

	time.AfterFunc(heartbeatPeriod, func() {
		RpcHeartBeat()
	})
}

func rpcCall(serviceMethod string, args, reply interface{}) error {
	return rpc.Call(globalConfig.rpcSrvAddr, serviceMethod, args, reply)
}
