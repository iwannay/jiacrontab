package main

import "jiacrontab/pkg/proto"

type Logic struct{}

func (l *Logic) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}
