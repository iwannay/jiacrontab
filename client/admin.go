package main

import (
	"jiacrontab/pkg/util"
)

type Admin struct{}

func (a *Admin) SystemInfo(args string, reply *map[string]interface{}) error {
	*reply = util.SystemInfo(startTime)
	return nil
}
