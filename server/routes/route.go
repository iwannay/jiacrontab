package routes

import (
	"crypto/md5"
	"fmt"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/server/conf"
	"jiacrontab/server/model"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iwannay/jiaweb"
)

func ListTask(ctx jiaweb.Context) error {

	var addr string
	var systemInfo map[string]interface{}
	var locals proto.Mdata
	var clientList map[string]proto.ClientConf
	var taskIdSli []string
	var r = ctx.Request()
	var m = model.NewModel()

	sortedTaskList := make([]*proto.TaskArgs, 0)
	sortedClientList := make([]proto.ClientConf, 0)

	clientList, _ = m.GetRPCClientList()

	if clientList != nil && len(clientList) > 0 {
		for _, v := range clientList {
			sortedClientList = append(sortedClientList, v)
		}
		sort.SliceStable(sortedClientList, func(i, j int) bool {
			return sortedClientList[i].Addr > sortedClientList[j].Addr
		})

		firstK := sortedClientList[0].Addr
		addr = libs.ReplaceEmpty(r.FormValue("addr"), firstK)
	} else {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "nothing to show",
		})
		return nil
	}

	locals = make(proto.Mdata)

	if err := m.RpcCall(addr, "Task.All", "", &locals); err != nil {
		ctx.Redirect("/", http.StatusFound)
		return err
	}

	if err := m.RpcCall(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		ctx.Redirect("/", http.StatusFound)
		return err
	}

	for _, v := range locals {
		taskIdSli = append(taskIdSli, v.Id)
		sortedTaskList = append(sortedTaskList, v)
	}
	sort.SliceStable(sortedTaskList, func(i, j int) bool {
		return sortedTaskList[i].Create > sortedTaskList[j].Create
	})

	tpl := []string{"listTask"}
	if cki, err := r.Cookie("model"); err == nil {
		if cki.Value == "batch" {
			tpl = []string{"batchListTask"}
		}
	}

	ctx.RenderHtml(tpl, map[string]interface{}{
		"title":      "灵魂百度",
		"list":       sortedTaskList,
		"addrs":      sortedClientList,
		"client":     clientList[addr],
		"systemInfo": systemInfo,
		"taskIds":    strings.Join(taskIdSli, ","),
		"url":        r.Url(),
	})

	return nil

}

func Index(ctx jiaweb.Context) error {
	var clientList map[string]proto.ClientConf
	var m = model.NewModel()

	sInfo := libs.SystemInfo(ctx.StartTime())
	clientList, _ = m.GetRPCClientList()
	sortedClientList := make([]proto.ClientConf, 0)

	for _, v := range clientList {
		sortedClientList = append(sortedClientList, v)
	}

	sort.Slice(sortedClientList, func(i, j int) bool {
		return (sortedClientList[i].Addr > sortedClientList[j].Addr) && (sortedClientList[i].State > sortedClientList[j].State)
	})
	ctx.RenderHtml([]string{"index"}, map[string]interface{}{
		"clientList": sortedClientList,
		"systemInfo": sInfo,
	})
	return nil

}

func UpdateTask(ctx jiaweb.Context) error {
	var reply bool
	var r = ctx.Request()
	var m = model.NewModel()

	sortedKeys := make([]string, 0)
	addr := strings.TrimSpace(r.FormValue("addr"))
	id := strings.TrimSpace(r.FormValue("taskId"))
	if addr == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "params error",
		})
		return nil
	}

	if r.Method == http.MethodPost {
		var unExitM, sync bool
		var pipeCommandList [][]string
		n := strings.TrimSpace(r.FormValue("taskName"))
		command := strings.TrimSpace(r.FormValue("command"))
		timeoutStr := strings.TrimSpace(r.FormValue("timeout"))
		mConcurrentStr := strings.TrimSpace(r.FormValue("maxConcurrent"))
		unpdExitM := r.FormValue("unexpectedExitMail")
		mSync := r.FormValue("sync")
		mailTo := strings.TrimSpace(r.FormValue("mailTo"))
		optimeout := strings.TrimSpace(r.FormValue("optimeout"))
		pipeCommands := r.PostForm["command[]"]
		pipeArgs := r.PostForm["args[]"]
		destSli := r.PostForm["depends[dest]"]
		cmdSli := r.PostForm["depends[command]"]
		argsSli := r.PostForm["depends[args]"]
		timeoutSli := r.PostForm["depends[timeout]"]
		depends := make([]proto.MScript, len(destSli))

		for k, v := range pipeCommands {
			pipeCommandList = append(pipeCommandList, []string{v, pipeArgs[k]})
		}

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
		month := libs.ReplaceEmpty(strings.TrimSpace(r.FormValue("month")), "*")
		weekday := libs.ReplaceEmpty(strings.TrimSpace(r.FormValue("weekday")), "*")
		day := libs.ReplaceEmpty(strings.TrimSpace(r.FormValue("day")), "*")
		hour := libs.ReplaceEmpty(strings.TrimSpace(r.FormValue("hour")), "*")
		minute := libs.ReplaceEmpty(strings.TrimSpace(r.FormValue("minute")), "*")

		if err := m.RpcCall(addr, "Task.Update", proto.TaskArgs{
			Id:                 id,
			Name:               n,
			Command:            command,
			Args:               a,
			PipeCommands:       pipeCommandList,
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
			ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
		if reply {
			ctx.Redirect("/list?addr="+addr, http.StatusFound)
			return nil
		}

	} else {
		var t proto.TaskArgs
		var clientList map[string]proto.ClientConf

		if id != "" {
			err := m.RpcCall(addr, "Task.Get", id, &t)
			if err != nil {
				ctx.Redirect("/list?addr="+addr, http.StatusFound)
				return err
			}
		} else {
			client, _ := m.SearchRPCClientList(addr)
			t.MailTo = client.Mail
		}
		if t.MaxConcurrent == 0 {
			t.MaxConcurrent = 1
		}

		clientList, _ = m.GetRPCClientList()

		if len(clientList) > 0 {
			for k := range clientList {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			firstK := sortedKeys[0]
			addr = libs.ReplaceEmpty(r.FormValue("addr"), firstK)
		} else {
			ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
				"error": "nothing to show",
			})
			return nil
		}

		ctx.RenderHtml([]string{"updateTask"}, map[string]interface{}{
			"addr":          addr,
			"addrs":         sortedKeys,
			"rpcClientsMap": clientList,
			"task":          t,
			"allowCommands": conf.ConfigArgs.AllowCommands,
		})
	}
	return nil

}

func StopTask(ctx jiaweb.Context) error {
	var r = ctx.Request()
	var m = model.NewModel()
	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	action := libs.ReplaceEmpty(r.FormValue("action"), "stop")
	var reply bool
	if taskId == "" || addr == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return nil
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
	if err := m.RpcCall(addr, method, taskId, &reply); err != nil {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": err,
		})
		return err
	}
	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return nil
	}

	ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
		"error": fmt.Sprintf("failed %s %s", method, taskId),
	})
	return nil

}

func StopAllTask(ctx jiaweb.Context) error {
	var r = ctx.Request()
	var m = model.NewModel()
	taskIds := strings.TrimSpace(r.FormValue("taskIds"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	method := "Task.StopAll"
	taskIdSli := strings.Split(taskIds, ",")
	var reply bool
	if len(taskIdSli) == 0 || addr == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return nil
	}

	if err := m.RpcCall(addr, method, taskIdSli, &reply); err != nil {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": err,
		})
		return err
	}
	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return nil
	}

	ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
		"error": fmt.Sprintf("failed %s %v", method, taskIdSli),
	})

	return nil

}

func StartTask(ctx jiaweb.Context) error {
	var r = ctx.Request()
	var m = model.NewModel()
	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply bool
	if taskId == "" || addr == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return nil
	}

	if err := m.RpcCall(addr, "Task.Start", taskId, &reply); err != nil {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return err
	}

	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return nil
	}

	ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
		"error": "failed start task" + taskId,
	})

	return nil
}

func Login(ctx jiaweb.Context) error {
	fmt.Println("hahah")
	var r = ctx.Request()
	if r.Method == http.MethodPost {

		u := r.FormValue("username")
		pwd := r.FormValue("passwd")
		remb := r.FormValue("remember")

		if u == conf.ConfigArgs.User && pwd == conf.ConfigArgs.Passwd {
			// md5p := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
			clientFeature := ctx.RemoteIP() + "-" + ctx.Request().Header.Get("User-Agent")
			fmt.Println("client feature", clientFeature)
			clientSign := fmt.Sprintf("%x", md5.Sum([]byte(clientFeature)))
			fmt.Println("client md5", clientSign)
			if remb == "yes" {
				ctx.GenerateToken(map[string]interface{}{
					"user":       u,
					"clientSign": clientSign,
				})
				// globalJwt.accessToken(rw, r, u, md5p)
			} else {
				// globalJwt.accessTempToken(rw, r, u, md5p)
				ctx.GenerateSeesionToken(map[string]interface{}{
					"user":       u,
					"clientSign": clientSign,
				})
			}

			ctx.Redirect("/", http.StatusFound)
			return nil
		}

		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "auth failed",
		})

	} else {
		// var user map[string]interface{}

		// if ctx.VerifyToken(&user) {
		// 	ctx.Redirect("/", http.StatusFound)
		// 	return nil
		// }
		ctx.RenderHtml([]string{"login"}, nil)

	}
	return nil
}

func QuickStart(ctx jiaweb.Context) error {
	var r = ctx.Request()
	var m = model.NewModel()

	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply []byte
	if taskId == "" || addr == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return nil
	}

	if err := m.RpcCall(addr, "Task.QuickStart", taskId, &reply); err != nil {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": err,
		})
		return nil
	}
	logList := strings.Split(string(reply), "\n")
	ctx.RenderHtml([]string{"log"}, map[string]interface{}{
		"logList": logList,
		"addr":    addr,
	})
	return nil
}

func Logout(ctx jiaweb.Context) error {
	ctx.CleanToken()
	ctx.Redirect("/login", http.StatusFound)
	return nil
}

func RecentLog(ctx jiaweb.Context) error {
	var r = ctx.Request()
	var m = model.NewModel()
	id := r.FormValue("taskId")
	addr := r.FormValue("addr")
	var content []byte
	if id == "" {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": "param error",
		})
		return nil
	}
	if err := m.RpcCall(addr, "Task.Log", id, &content); err != nil {
		ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
			"error": err,
		})
		return nil
	}
	logList := strings.Split(string(content), "\n")

	ctx.RenderHtml([]string{"log"}, map[string]interface{}{
		"logList": logList,
		"addr":    addr,
	})
	return nil

}

func Readme(ctx jiaweb.Context) error {

	ctx.RenderHtml([]string{"readme"}, nil)
	return nil

}

func ReloadConfig(ctx jiaweb.Context) error {
	conf.ConfigArgs.Reload()
	ctx.Redirect("/", http.StatusFound)
	return nil
}

func DeleteClient(ctx jiaweb.Context) error {
	m := model.NewModel()
	r := ctx.Request()
	addr := r.FormValue("addr")
	m.InnerStore().Wrap(func(s *model.Store) {

		if v, ok := s.RpcClientList[addr]; ok {
			if v.State == 1 {
				return
			}
		}
		delete(s.RpcClientList, addr)

	}).Sync()
	ctx.Redirect("/", http.StatusFound)
	return nil
}

func ViewConfig(ctx jiaweb.Context) error {

	c := conf.ConfigArgs.Category()
	r := ctx.Request()

	if r.Method == http.MethodPost {
		mailTo := strings.TrimSpace(r.FormValue("mailTo"))
		libs.SendMail("测试邮件", "测试邮件请勿回复", conf.ConfigArgs.MailHost, conf.ConfigArgs.MailUser, conf.ConfigArgs.MailPass, conf.ConfigArgs.MailPort, mailTo)
	}

	ctx.RenderHtml([]string{"viewConfig"}, map[string]interface{}{
		"configs": c,
	})
	return nil
}

func Model(ctx jiaweb.Context) error {
	rw := ctx.Response().ResponseWriter()
	r := ctx.Request()
	val := r.FormValue("type")
	url := r.FormValue("url")
	http.SetCookie(rw, &http.Cookie{
		Name:     "model",
		Path:     "/",
		Value:    val,
		HttpOnly: true,
	})

	ctx.Redirect(url, http.StatusFound)
	return nil

}
