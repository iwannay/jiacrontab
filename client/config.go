package main

import ini "gopkg.in/ini.v1"
import "log"

const configFile = "client.ini"

type config struct {
	debug               bool
	debugScript         bool
	rpcListenAddr       string
	addr                string
	dataFile            string
	rpcSrvAddr          string
	logPath             string
	defaultRPCPath      string
	defaultRPCDebugPath string
	pprofAddr           string
	mailUser            string
	mailPass            string
	mailHost            string
	mailTo              string
}

func newConfig() *config {
	cf := loadConfig()

	rpc := cf.Section("rpc")
	logc := cf.Section("log")
	base := cf.Section("base")
	server := cf.Section("server")
	mail := cf.Section("mail")

	c := &config{
		debug:         base.Key("debug").MustBool(false),
		debugScript:   base.Key("debugScript").MustBool(false),
		rpcListenAddr: rpc.Key("listen").MustString(":20001"),
		pprofAddr:     server.Key("pprof_addr").MustString(":20002"),
		addr:          rpc.Key("local_addr").MustString("localhost:20002"),
		rpcSrvAddr:    rpc.Key("server_addr").MustString("localhost:20003"),
		dataFile:      base.Key("data_file").MustString("data.json"),
		logPath:       logc.Key("dir").MustString("./logs"),
		mailTo:        mail.Key("to").MustString(""),

		defaultRPCDebugPath: "/debug/rpc",
		defaultRPCPath:      "/__myrpc__",
	}
	log.Printf("config:%v", *c)
	return c
}

func loadConfig() *ini.File {
	f, err := ini.Load(configFile)
	if err != nil {
		panic(err)
	}
	return f
}
