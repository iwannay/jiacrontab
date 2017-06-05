package main

import (
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/server/store"
	"log"
)

type Logic struct{}

func (l *Logic) Register(args proto.ClientConf, reply *proto.MailArgs) error {
	defer libs.MRecover()

	*reply = proto.MailArgs{
		Host: globalConfig.mailHost,
		User: globalConfig.mailUser,
		Pass: globalConfig.mailPass,
		Port: globalConfig.mailPort,
	}

	// globalStore.Get(fmt.Sprintf("RPCClientList.%s", args.Addr), &clientConf)

	// if clientConf, ok = globalStore.SearchRPCClientList(args.Addr); ok {

	// }

	// if clientConf.State == 1 {
	// 	return nil
	// }

	globalStore.Wrap(func(s *store.Store) {
		s.RpcClientList[args.Addr] = proto.ClientConf{
			Addr:  args.Addr,
			State: 1,
		}
	}).Sync()

	log.Println("register client", args)

	return nil
}
