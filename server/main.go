package main

import (
	"log"
	_ "net/http/pprof"
	"runtime"
	"time"
)

var globalConfig *config
var globalStore *Store
var globalJwt *mjwt

// var rpcClient *mrpcClient
// var rpcClients = make(map[string]*mrpcClient)
var globalReqFilter *reqFilter
var startTime = time.Now()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	globalConfig = newConfig()
	globalJwt = newJwt(globalConfig.tokenExpires, globalConfig.tokenCookieName, globalConfig.JWTSigningKey, globalConfig.tokenCookieMaxAge)

	globalStore = newStore(globalConfig.dataFile)
	if err := globalStore.Load(); err != nil {
		log.Printf("failed recover %s", err)
	}
	globalStore.Update(nil)

	globalReqFilter = newReqFilter()

	go initSrvRpc(&Logic{})
	initServer()
}
