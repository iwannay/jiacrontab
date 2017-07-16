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

	globalStore.Wrap(func(s *store.Store) {
		s.RpcClientList[args.Addr] = args
	}).Sync()

	log.Println("register client", args)
	return nil
}

func (l *Logic) Depends(args []proto.MScript, reply *bool) error {
	defer func() {
		log.Println("call Logic.Depends end")
	}()
	*reply = true
	for _, v := range args {
		if err := rpcCall(v.Dest, "Task.ExecDepend", v, &reply); err != nil {
			*reply = false
			return err
		}
	}

	return nil
}

func (l *Logic) DependDone(args proto.MScript, reply *bool) error {
	defer func() {
		log.Println("call Logic.DependDone end")
	}()
	*reply = true
	if err := rpcCall(args.Dest, "Task.ResolvedSDepends", args, &reply); err != nil {
		*reply = false
		return err
	}

	return nil
}
