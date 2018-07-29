package main

import (
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
	"jiacrontab/model"
	"log"
	"time"
)

const heartbeatPeriod = 1 * time.Minute

func RpcHeartBeat() {
	var mail proto.MailArgs

	err := rpcCall("Logic.Register", model.Client{
		Addr:           globalConfig.addr,
		DaemonTaskNum:  globalDaemon.count(),
		CrontabTaskNum: globalCrontab.count(),
		State:          1,
		Mail:           globalConfig.mailTo,
	}, &mail)

	if err != nil {
		log.Println("Logic.Register error:", err, "server addr:", globalConfig.rpcSrvAddr)
	}

	time.AfterFunc(heartbeatPeriod, RpcHeartBeat)
}

func rpcCall(serviceMethod string, args, reply interface{}) error {
	return rpc.Call(globalConfig.rpcSrvAddr, serviceMethod, args, reply)
}
