package main

import (
	"jiacrontab/client/store"
	"jiacrontab/libs/finder"
	"jiacrontab/libs/rpc"
	"log"
	"os"
	"os/signal"
	"time"

	"jiacrontab/model"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
var globalDaemon *daemon
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	model.CreateDB("sqlite3", "data/jiacrontab_client.db")
	model.DB().CreateTable(&model.DaemonTask{}, &model.CrontabTask{})
	model.DB().AutoMigrate(&model.DaemonTask{}, &model.CrontabTask{})

	globalConfig = newConfig()
	if globalConfig.debug == true {
		initPprof(globalConfig.pprofAddr)
	}

	globalStore = store.NewStore(globalConfig.dataFile)
	globalStore.Load()
	globalStore.Sync()
	globalStore.Export2DB()

	globalCrontab = newCrontab(10)
	globalCrontab.run()

	globalDepend = newDepend()
	globalDepend.run()

	globalDaemon = newDaemon(100)
	globalDaemon.run()

	go listenSignal(func() {
		globalCrontab.lock.Lock()
		for k, v := range globalCrontab.handleMap {
			for _, item := range v.taskPool {
				item.cancel()
			}
			log.Printf("kill %d", k)
		}
		globalCrontab.lock.Unlock()
		model.DB().Model(&model.CrontabTask{}).Update(map[string]interface{}{
			"number_process": 0,
			"timer_counter":  0,
		})
		globalDaemon.lock.Lock()
		for _, v := range globalDaemon.taskMap {
			v.cancel()
		}
		globalDaemon.lock.Unlock()
		globalDaemon.waitDone()

	})

	go RpcHeartBeat()

	if globalConfig.cleanTaskLog {
		go finder.SearchAndDeleteFileOnDisk(globalConfig.logPath, 24*time.Hour*30, 1<<30)
	}

	rpc.ListenAndServe(globalConfig.rpcListenAddr, &Logic{}, &DaemonTask{}, &Admin{}, &CrontabTask{})
}
