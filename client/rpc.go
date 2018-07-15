package main

import (
	"jiacrontab/client/store"
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
	"jiacrontab/model"
	"log"
	"time"
)

const heartbeatPeriod = 1 * time.Minute

func RpcHeartBeat() {
	var mail proto.MailArgs

	log.Println("heart beat", globalConfig.rpcSrvAddr, "start")
	err := rpc.Call(globalConfig.rpcSrvAddr, "Logic.Register", model.Client{
		Addr:           globalConfig.addr,
		DaemonTaskNum:  globalDaemon.count(),
		CrontabTaskNum: globalCrontab.count(),
		State:          1,
		Mail:           globalConfig.mailTo,
	}, &mail)

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
