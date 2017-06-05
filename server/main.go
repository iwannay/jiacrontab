package main

import (
	"jiacrontab/server/rpc"
	"jiacrontab/server/store"
	_ "net/http/pprof"
	"runtime"
	"time"
)

var globalConfig *config
var globalStore *store.Store
var globalJwt *mjwt
var globalReqFilter *reqFilter
var startTime = time.Now()

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	globalConfig = newConfig()
	globalJwt = newJwt(globalConfig.tokenExpires, globalConfig.tokenCookieName, globalConfig.JWTSigningKey, globalConfig.tokenCookieMaxAge)

	globalStore = store.NewStore(globalConfig.dataFile)
	globalStore.Load()

	globalReqFilter = newReqFilter()

	go rpc.InitSrvRpc(globalConfig.defaultRPCPath, globalConfig.defaultRPCDebugPath, globalConfig.rpcAddr, &Logic{})

	initServer()
}
