package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type reqInfo struct {
	createTime int64
	updateTime int64
	counter    int64
	routeInfo  map[string]*routeInfo
}

type routeInfo struct {
	counter    int64
	createTime int64
	updateTime int64
}

type reqFilter struct {
	requestCounterMap map[string]*reqInfo
	callLimit         int64
	routeCallLimit    map[string]int64
	filterDuration    int64
	updateDuration    int64
	tokenMap          map[string]bool
	l                 sync.RWMutex
}

func newReqFilter() *reqFilter {

	r := &reqFilter{
		requestCounterMap: make(map[string]*reqInfo),
		callLimit:         1000, // 相同客户端filterDuration内调用接口最大次数
		filterDuration:    3600,
		updateDuration:    60 * 60, // 垃圾回收扫描周期
		routeCallLimit: map[string]int64{
			"/login:POST": 5,
		},
	}
	go r.update()
	return r
}

func (self *reqFilter) filter(rw http.ResponseWriter, r *http.Request) error {
	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return errors.New("request non-browser initiated")
	}
	k := fmt.Sprintf("%s|%s", getHttpClientIp(r), r.Header.Get("User-Agent"))
	self.l.Lock()
	defer self.l.Unlock()
	if v, ok := self.requestCounterMap[k]; ok {
		v.updateTime = time.Now().Unix()
		if (time.Now().Unix() - v.createTime) > self.filterDuration {
			v.counter = 0
			v.createTime = time.Now().Unix()
			log.Printf("%s the counter is reset to 0", k)
		}
		if v.counter > self.callLimit {
			errMsg := fmt.Sprintf("%s  call api %d reaches the maximum limit %d", k, v.counter, self.callLimit)
			log.Printf("%s:%s", k, errMsg)
			return errors.New(errMsg)
		}
		if e, ok := v.routeInfo[r.URL.Path+":"+r.Method]; ok {
			if e2, ok := self.routeCallLimit[r.URL.Path+":"+r.Method]; ok {
				e.updateTime = time.Now().Unix()
				if time.Now().Unix()-e.createTime > self.filterDuration {
					e.counter = 0
					e.createTime = time.Now().Unix()
					log.Printf("%s %s:%s the counter is reset to 0", k, r.URL.Path, r.Method)
				}
				if e.counter > e2 {
					errMsg := fmt.Sprintf("%s:%s call %d reaches the maximum limit %d", r.URL.Path, r.Method, e.counter, e2)
					log.Printf("%s:%s", k, errMsg)
					return errors.New(errMsg)
				}
			}
			e.counter++
		} else {
			v.routeInfo[r.URL.Path+":"+r.Method] = &routeInfo{
				createTime: time.Now().Unix(),
				updateTime: time.Now().Unix(),
				counter:    1,
			}
		}
		v.counter++
	} else {
		self.requestCounterMap[k] = &reqInfo{
			createTime: time.Now().Unix(),
			updateTime: time.Now().Unix(),
			counter:    1,
			routeInfo: map[string]*routeInfo{r.URL.Path + ":" + r.Method: &routeInfo{
				createTime: time.Now().Unix(),
				updateTime: time.Now().Unix(),
				counter:    1,
			}},
		}
	}

	return nil

}

func (self *reqFilter) idle() {
	now := time.Now().Unix()
	var dura int64 = 3600 * 24 // 一天未操作，清理客户端相关route信息
	self.l.Lock()
	defer self.l.Unlock()
	for k, v := range self.requestCounterMap {
		if now-v.updateTime > dura {
			delete(self.requestCounterMap, k)
			log.Printf("remove %s", k)
			return
		}
		for k2, v2 := range v.routeInfo {
			if now-v2.updateTime > dura {
				delete(v.routeInfo, k2)
				log.Printf("remove %s %s", k, k2)
				return
			}
		}
	}

}

func (self *reqFilter) countClient() int {
	self.l.RLock()
	count := len(self.requestCounterMap)
	self.l.RUnlock()
	return count
}

func (self *reqFilter) update() {
	t := time.Tick(time.Duration(self.updateDuration * int64(time.Second)))
	for {
		select {
		case <-t:
			start := time.Now().UnixNano()
			self.idle()
			end := time.Now().UnixNano()
			log.Printf("scaning request map cost %dns", end-start)
		}
	}
}
