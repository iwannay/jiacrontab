package main

import (
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
)

type Logic struct{}

func (l *Logic) Register(args proto.ClientConf, reply *proto.MailArgs) error {
	defer libs.MRecover()
	log.Println("register client", args)
	if _, err := newRpcClient(args.Addr); err == nil {

		*reply = proto.MailArgs{
			Host: globalConfig.mailHost,
			User: globalConfig.mailUser,
			Pass: globalConfig.mailPass,
			Port: globalConfig.mailPort,
		}

	}
	return nil
}
