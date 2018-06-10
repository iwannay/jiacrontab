package main

import (
	"jiacrontab/client/store"
	"jiacrontab/libs/rpc"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"
)

func newScheduler(config *config, crontab *crontab, store *store.Store) *scheduler {
	return &scheduler{
		config:  config,
		crontab: crontab,
		store:   store,
	}
}

type scheduler struct {
	config  *config
	crontab *crontab
	store   *store.Store
}

var globalConfig *config
var globalCrontab *crontab
var globalStore *store.Store
var globalDepend *depend
var startTime = time.Now()

func listenSignal(fn func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		sign := <-c
		log.Println("get signal:", sign)
		if fn != nil {
			fn()
		}
		log.Fatal("trying to exit gracefully...")

	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	globalConfig = newConfig()
	if globalConfig.debug == true {
		initPprof(globalConfig.pprofAddr)
	}

	globalStore = store.NewStore(globalConfig.dataFile)
	globalStore.Load()
	globalStore.Sync()

	globalCrontab = newCrontab(10)
	globalCrontab.run()

	globalDepend = newDepend()
	globalDepend.run()

	go listenSignal(func() {
		globalCrontab.lock.Lock()
		for k, v := range globalCrontab.handleMap {
			for _, item := range v.taskPool {
				item.cancel()
			}
			log.Printf("kill %s", k)
		}
		globalCrontab.lock.Unlock()
		globalStore.Update(func(s *store.Store) {
			for _, v := range s.TaskList {
				v.NumberProcess = 0
			}
		})

	})

	go RpcHeartBeat()
	rpc.ListenAndServe(globalConfig.rpcListenAddr, &Task{}, &Admin{})
}
