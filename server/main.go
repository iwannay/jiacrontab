package main

import (
	"encoding/json"
	"jiacrontab/libs/proto"
	"jiacrontab/server/rpc"
	"jiacrontab/server/store"
	_ "net/http/pprof"
	"runtime"
	"time"
)

var globalConfig *config
var globalStore *store.Store
var globalJwt *mjwt

// var rpcClient *mrpcClient
// var rpcClients = make(map[string]*mrpcClient)
var globalReqFilter *reqFilter
var startTime = time.Now()
var registerRpcChan = make(chan string)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	globalConfig = newConfig()
	globalJwt = newJwt(globalConfig.tokenExpires, globalConfig.tokenCookieName, globalConfig.JWTSigningKey, globalConfig.tokenCookieMaxAge)

	globalStore = store.NewStore(globalConfig.dataFile)
	globalStore.Load()
	globalStore.Wrap(func(s *store.Store) {

		for k, v := range s.Data {
			switch k {
			case "RPCClientList":
				var rpcClientList map[string]proto.ClientConf
				b, err := json.Marshal(v)
				if err != nil {
					panic(err)
				}
				if err := json.Unmarshal(b, &rpcClientList); err != nil {
					panic(err)
				}
				s.Data["RPCClientList"] = rpcClientList
			}
		}
	})

	globalReqFilter = newReqFilter()

	go rpc.InitSrvRpc(globalConfig.defaultRPCPath, globalConfig.defaultRPCDebugPath, globalConfig.rpcAddr, &Logic{})

	initServer()
}
