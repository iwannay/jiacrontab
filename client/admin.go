package main

import "jiacrontab/libs"

type Admin struct{}

func (a *Admin) SystemInfo(args string, reply *map[string]interface{}) error {
	*reply = libs.SystemInfo(startTime)
	return nil
}
