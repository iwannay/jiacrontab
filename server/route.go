package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/server/rpc"
	"jiacrontab/server/store"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func listTask(rw http.ResponseWriter, r *http.Request, m *modelView) {

	var addr string
	var systemInfo map[string]interface{}
	var locals proto.Mdata
	var clientList map[string]proto.ClientConf

	sortedKeys := make([]string, 0)
	sortedKeys2 := make([]string, 0)
	m.s.GetCopy("RPCClientList", &clientList)

	if clientList != nil && len(clientList) > 0 {
		for k := range clientList {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		firstK := sortedKeys[0]
		addr = replaceEmpty(r.FormValue("addr"), firstK)
	} else {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "nothing to show",
		}, nil)
		return
	}

	locals = make(proto.Mdata)

	if err := m.rpcCall(addr, "Task.All", "", &locals); err != nil {
		http.Redirect(rw, r, "/", http.StatusFound)
		return
	}

	for k := range locals {
		sortedKeys2 = append(sortedKeys2, k)
	}
	sort.Strings(sortedKeys2)

	m.rpcCall(addr, "Task.SystemInfo", "", &systemInfo)
	m.renderHtml2([]string{"listTask"}, map[string]interface{}{
		"title":         "灵魂百度",
		"list":          locals,
		"addrs":         sortedKeys,
		"listKey":       sortedKeys2,
		"rpcClientsMap": clientList,
		"client":        clientList[addr],
		"addr":          addr,
		"systemInfo":    systemInfo,
		"appName":       globalConfig.appName,
	}, template.FuncMap{
		"date":     date,
		"formatMs": int2floatstr,
	})

}

func index(rw http.ResponseWriter, r *http.Request, m *modelView) {
	var clientList map[string]proto.ClientConf
	if r.URL.Path != "/" {
		rw.WriteHeader(http.StatusNotFound)
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "404 page not found",
		}, nil)
		return
	}

	sInfo := libs.SystemInfo(startTime)
	sortedKeys := make([]string, 0)
	m.s.Get("RPCClientList", &clientList)

	for k := range clientList {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	m.renderHtml2([]string{"index"}, map[string]interface{}{
		"rpcClientsKey": sortedKeys,
		"rpcClientsMap": clientList,
		"systemInfo":    sInfo,
	}, template.FuncMap{
		"date": date,
	})

}

func updateTask(rw http.ResponseWriter, r *http.Request, m *modelView) {
	var reply bool

	sortedKeys := make([]string, 0)
	addr := strings.TrimSpace(r.FormValue("addr"))
	id := strings.TrimSpace(r.FormValue("taskId"))
	if addr == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "params error",
		}, nil)
		return
	}

	if r.Method == http.MethodPost {
		log.Println(r.Form)
		n := strings.TrimSpace(r.FormValue("taskName"))
		command := strings.TrimSpace(r.FormValue("command"))
		timeoutStr := strings.TrimSpace(r.FormValue("timeout"))
		timeout, err := strconv.Atoi(timeoutStr)
		optimeout := strings.TrimSpace(r.FormValue("optimeout"))
		if _, ok := map[string]bool{"email": true, "kill": true, "email_and_kill": true, "ignore": true}[optimeout]; !ok {
			optimeout = "ignore"
		}

		if err != nil {
			timeout = 0
		}

		a := r.FormValue("args")
		month := replaceEmpty(strings.TrimSpace(r.FormValue("month")), "*")
		weekday := replaceEmpty(strings.TrimSpace(r.FormValue("weekday")), "*")
		day := replaceEmpty(strings.TrimSpace(r.FormValue("day")), "*")
		hour := replaceEmpty(strings.TrimSpace(r.FormValue("hour")), "*")
		minute := replaceEmpty(strings.TrimSpace(r.FormValue("minute")), "*")

		if err := m.rpcCall(addr, "Task.Update", proto.TaskArgs{
			Id:        id,
			Name:      n,
			Command:   command,
			Args:      a,
			Timeout:   int64(timeout),
			OpTimeout: optimeout,
			Create:    time.Now().Unix(),
			C: struct {
				Weekday string
				Month   string
				Day     string
				Hour    string
				Minute  string
			}{

				Month:   month,
				Day:     day,
				Hour:    hour,
				Minute:  minute,
				Weekday: weekday,
			},
		}, &reply); err != nil {
			m.renderHtml2([]string{"public/error"}, map[string]interface{}{
				"error": err.Error(),
			}, nil)
			return
		}
		if reply {
			http.Redirect(rw, r, "/list?addr="+addr, http.StatusFound)
			return
		}

	} else {
		var t proto.TaskArgs
		var clientList map[string]*rpc.MrpcClient
		if id != "" {
			m.rpcCall(addr, "Task.Get", id, &t)
			if reply {
				http.Redirect(rw, r, "/list?addr="+addr, http.StatusFound)
				return
			}
		}
		m.s.GetCopy("RPCClientList", &clientList)
		if len(clientList) > 0 {
			for k := range clientList {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			firstK := sortedKeys[0]
			addr = replaceEmpty(r.FormValue("addr"), firstK)
		} else {
			m.renderHtml2([]string{"public/error"}, map[string]interface{}{
				"error": "nothing to show",
			}, nil)
			return
		}

		m.renderHtml2([]string{"updateTask"}, map[string]interface{}{
			"addr":          addr,
			"addrs":         sortedKeys,
			"rpcClientsMap": clientList,
			"task":          t,
			"allowCommands": globalConfig.allowCommands,
		}, nil)
	}

}

func stopTask(rw http.ResponseWriter, r *http.Request, m *modelView) {
	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	action := replaceEmpty(r.FormValue("action"), "stop")
	var reply bool
	if taskId == "" || addr == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}

	// if c, err := newRpcClient(addr); err != nil {
	// 	m.renderHtml2([]string{"public/error"}, map[string]interface{}{
	// 		"error": "failed stop task" + taskId,
	// 	}, nil)
	// 	return
	// } else {
	var method string
	if action == "stop" {
		method = "Task.Stop"
	} else if action == "delete" {
		method = "Task.Delete"
	} else {
		method = "Task.Kill"
	}
	if err := m.rpcCall(addr, method, taskId, &reply); err != nil {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": err,
		}, nil)
		return
	}
	if reply {
		http.Redirect(rw, r, "/list?addr="+addr, http.StatusFound)
		return
	}

	m.renderHtml2([]string{"public/error"}, map[string]interface{}{
		"error": fmt.Sprintf("failed %s %s", method, taskId),
	}, nil)

}

func startTask(rw http.ResponseWriter, r *http.Request, m *modelView) {
	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply bool
	if taskId == "" || addr == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}

	if err := m.rpcCall(addr, "Task.Start", taskId, &reply); err != nil {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}

	if reply {
		http.Redirect(rw, r, "/list?addr="+addr, http.StatusFound)
		return
	}

	m.renderHtml2([]string{"error"}, map[string]interface{}{
		"error": "failed start task" + taskId,
	}, nil)

}

func login(rw http.ResponseWriter, r *http.Request, m *modelView) {
	if r.Method == http.MethodPost {

		u := r.FormValue("username")
		pwd := r.FormValue("passwd")
		remb := r.FormValue("remember")

		if u == globalConfig.user && pwd == globalConfig.passwd {
			md5p := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
			if remb == "yes" {
				globalJwt.accessToken(rw, r, u, md5p)
			} else {
				globalJwt.accessTempToken(rw, r, u, md5p)
			}

			http.Redirect(rw, r, "/", http.StatusFound)
			return
		}

		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "auth failed",
		}, nil)

	} else {
		var user map[string]interface{}
		if globalJwt.auth(rw, r, &user) {
			http.Redirect(rw, r, "/", http.StatusFound)
			return
		}
		m.renderHtml2([]string{"login"}, nil, nil)

	}
}

func quickStart(rw http.ResponseWriter, r *http.Request, m *modelView) {
	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply []byte
	if taskId == "" || addr == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}

	if err := m.rpcCall(addr, "Task.QuickStart", taskId, &reply); err != nil {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": err,
		}, nil)
		return
	}
	logList := strings.Split(string(reply), "\n")
	m.renderHtml2([]string{"log"}, map[string]interface{}{
		"logList": logList,
		"addr":    addr,
	}, nil)

}

func logout(rw http.ResponseWriter, r *http.Request, m *modelView) {
	globalJwt.cleanCookie(rw)
	http.Redirect(rw, r, "/login", http.StatusFound)
}

func recentLog(rw http.ResponseWriter, r *http.Request, m *modelView) {
	id := r.FormValue("taskId")
	addr := r.FormValue("addr")
	var content []byte
	if id == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}
	if err := m.rpcCall(addr, "Task.Log", id, &content); err != nil {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": err,
		}, nil)
		return
	}
	logList := strings.Split(string(content), "\n")

	m.renderHtml2([]string{"log"}, map[string]interface{}{
		"logList": logList,
		"addr":    addr,
	}, nil)
	return

}

func readme(rw http.ResponseWriter, r *http.Request, m *modelView) {

	m.renderHtml2([]string{"readme"}, map[string]interface{}{}, nil)
	return

}

func reloadConfig(rw http.ResponseWriter, r *http.Request, m *modelView) {
	globalConfig.reload()
	http.Redirect(rw, r, "/", http.StatusFound)
	log.Println("reload config")
}

func deleteClient(rw http.ResponseWriter, r *http.Request, m *modelView) {

	addr := r.FormValue("addr")
	m.s.Wrap(func(s *store.Store) {
		if clientList, ok := s.Data["RPCClientList"].(map[string]proto.ClientConf); ok {
			if v, ok := clientList[addr]; ok {
				if v.State == 1 {
					return
				}
			}
			delete(clientList, addr)
		}
	}).Sync()
	http.Redirect(rw, r, "/", http.StatusFound)
}

func viewConfig(rw http.ResponseWriter, r *http.Request, m *modelView) {
	c := make(map[string]interface{})
	values := reflect.ValueOf(*globalConfig)
	types := reflect.TypeOf(*globalConfig)
	l := values.NumField()
	for i := 0; i < l; i++ {
		v := values.Field(i)
		t := types.Field(i)
		if t.Name == "passwd" {
			continue
		}
		switch v.Kind() {
		case reflect.String:
			c[t.Name] = v.String()
		case reflect.Int64:
			c[t.Name] = v.Int()
		case reflect.Bool:
			c[t.Name] = v.Bool()
		}

	}

	m.renderHtml2([]string{"viewConfig"}, map[string]interface{}{
		"configs": c,
	}, nil)
	return
}
