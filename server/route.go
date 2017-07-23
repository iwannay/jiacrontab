package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/server/store"
	"log"
	"net/http"
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
	var taskIdSli []string

	sortedTaskList := make([]*proto.TaskArgs, 0)
	sortedClientList := make([]proto.ClientConf, 0)
	clientList, _ = m.s.GetRPCClientList()

	if clientList != nil && len(clientList) > 0 {
		for _, v := range clientList {
			sortedClientList = append(sortedClientList, v)
		}
		sort.SliceStable(sortedClientList, func(i, j int) bool {
			return sortedClientList[i].Addr > sortedClientList[j].Addr
		})

		firstK := sortedClientList[0].Addr
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

	if err := m.rpcCall(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		http.Redirect(rw, r, "/", http.StatusFound)
		return
	}

	for _, v := range locals {
		taskIdSli = append(taskIdSli, v.Id)
		sortedTaskList = append(sortedTaskList, v)
	}
	sort.SliceStable(sortedTaskList, func(i, j int) bool {
		return sortedTaskList[i].Create > sortedTaskList[j].Create
	})

	m.renderHtml2([]string{"listTask"}, map[string]interface{}{
		"title":      "灵魂百度",
		"list":       sortedTaskList,
		"addrs":      sortedClientList,
		"client":     clientList[addr],
		"systemInfo": systemInfo,
		"taskIds":    strings.Join(taskIdSli, ","),
		"appName":    globalConfig.appName,
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
	clientList, _ = m.s.GetRPCClientList()
	sortedClientList := make([]proto.ClientConf, 0)

	for _, v := range clientList {
		sortedClientList = append(sortedClientList, v)
	}

	sort.Slice(sortedClientList, func(i, j int) bool {
		return (sortedClientList[i].Addr > sortedClientList[j].Addr) && (sortedClientList[i].State > sortedClientList[j].State)
	})
	m.renderHtml2([]string{"index"}, map[string]interface{}{
		"clientList": sortedClientList,
		"systemInfo": sInfo,
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
		var unExitM, sync bool
		n := strings.TrimSpace(r.FormValue("taskName"))
		command := strings.TrimSpace(r.FormValue("command"))
		timeoutStr := strings.TrimSpace(r.FormValue("timeout"))
		mConcurrentStr := strings.TrimSpace(r.FormValue("maxConcurrent"))
		unpdExitM := r.FormValue("unexpectedExitMail")
		mSync := r.FormValue("sync")
		mailTo := strings.TrimSpace(r.FormValue("mailTo"))
		optimeout := strings.TrimSpace(r.FormValue("optimeout"))

		destSli := r.PostForm["depends[dest]"]
		cmdSli := r.PostForm["depends[command]"]
		argsSli := r.PostForm["depends[args]"]
		timeoutSli := r.PostForm["depends[timeout]"]
		depends := make([]proto.MScript, len(destSli))

		for k, v := range destSli {
			depends[k].Dest = v
			depends[k].From = addr
			depends[k].Args = argsSli[k]
			tmpT, err := strconv.Atoi(timeoutSli[k])

			if err != nil {
				depends[k].Timeout = 0
			} else {
				depends[k].Timeout = int64(tmpT)
			}
			depends[k].Command = cmdSli[k]
		}

		if unpdExitM == "1" {
			unExitM = true
		} else {
			unExitM = false
		}
		if mSync == "1" {
			sync = true
		} else {
			sync = false
		}

		if _, ok := map[string]bool{"email": true, "kill": true, "email_and_kill": true, "ignore": true}[optimeout]; !ok {
			optimeout = "ignore"
		}
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			timeout = 0
		}

		maxConcurrent, err := strconv.Atoi(mConcurrentStr)
		if err != nil {
			maxConcurrent = 10
		}

		a := r.FormValue("args")
		month := replaceEmpty(strings.TrimSpace(r.FormValue("month")), "*")
		weekday := replaceEmpty(strings.TrimSpace(r.FormValue("weekday")), "*")
		day := replaceEmpty(strings.TrimSpace(r.FormValue("day")), "*")
		hour := replaceEmpty(strings.TrimSpace(r.FormValue("hour")), "*")
		minute := replaceEmpty(strings.TrimSpace(r.FormValue("minute")), "*")

		if err := m.rpcCall(addr, "Task.Update", proto.TaskArgs{
			Id:                 id,
			Name:               n,
			Command:            command,
			Args:               a,
			Timeout:            int64(timeout),
			OpTimeout:          optimeout,
			Create:             time.Now().Unix(),
			MailTo:             mailTo,
			MaxConcurrent:      maxConcurrent,
			Depends:            depends,
			UnexpectedExitMail: unExitM,
			Sync:               sync,
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
		var clientList map[string]proto.ClientConf

		if id != "" {
			err := m.rpcCall(addr, "Task.Get", id, &t)
			if err != nil {
				http.Redirect(rw, r, "/list?addr="+addr, http.StatusFound)
				return
			}
		} else {
			client, _ := m.s.SearchRPCClientList(addr)
			t.MailTo = client.Mail
		}
		if t.MaxConcurrent == 0 {
			t.MaxConcurrent = 1
		}

		clientList, _ = m.s.GetRPCClientList()

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

func stopAllTask(rw http.ResponseWriter, r *http.Request, m *modelView) {
	taskIds := strings.TrimSpace(r.FormValue("taskIds"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	method := "Task.StopAll"
	taskIdSli := strings.Split(taskIds, ",")
	var reply bool
	if len(taskIdSli) == 0 || addr == "" {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		}, nil)
		return
	}

	if err := m.rpcCall(addr, method, taskIdSli, &reply); err != nil {
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
		"error": fmt.Sprintf("failed %s %v", method, taskIdSli),
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

		if v, ok := s.RpcClientList[addr]; ok {
			if v.State == 1 {
				return
			}
		}
		delete(s.RpcClientList, addr)

	}).Sync()
	http.Redirect(rw, r, "/", http.StatusFound)
}

func viewConfig(rw http.ResponseWriter, r *http.Request, m *modelView) {

	c := globalConfig.category()

	if r.Method == http.MethodPost {
		mailTo := strings.TrimSpace(r.FormValue("mailTo"))
		libs.SendMail("测试邮件", "测试邮件请勿回复", globalConfig.mailHost, globalConfig.mailUser, globalConfig.mailPass, globalConfig.mailPort, mailTo)
	}

	m.renderHtml2([]string{"viewConfig"}, map[string]interface{}{
		"configs": c,
	}, nil)
	return
}
