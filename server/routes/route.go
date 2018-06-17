package routes

import (
	"crypto/md5"
	"encoding/json"

	"fmt"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
	"jiacrontab/server/conf"
	"jiacrontab/server/model"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"net/url"

	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
)

// var app *jiaweb.JiaWeb

const (
	minuteTimeLayout = "200601021504"
	dateTimeLayout   = "2006-01-02 15:04:05"
)

func ListTask(ctx iris.Context) {

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
		if ctx.IsAjax() {
			ctx.JSON(map[string]interface{}{
				"code": -1,
			})
			return
		}

		ctx.ViewData("error", "nothing to show")
		ctx.View("public/error.html")

		return
	}

	locals = make(proto.Mdata)

	if err := rpc.Call(addr, "Task.All", "", &locals); err != nil {

		if ctx.IsAjax() {
			ctx.JSON(map[string]interface{}{
				"code": -1,
			})
			return
		}
		ctx.Redirect("/", http.StatusFound)
		return
	}

	if err := rpc.Call(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		if ctx.IsAjax() {
			ctx.JSON(map[string]interface{}{
				"code": -1,
			})
			return
		}
		ctx.Redirect("/", http.StatusFound)
		return
	}

	for _, v := range locals {
		taskIdSli = append(taskIdSli, v.Id)
		sortedTaskList = append(sortedTaskList, v)
	}
	sort.SliceStable(sortedTaskList, func(i, j int) bool {
		return sortedTaskList[i].Create > sortedTaskList[j].Create
	})

	// tpl := "listTask.html"
	// if cki, err := r.Cookie("model"); err == nil {
	// 	if cki.Value == "batch" {
	// 		tpl = "batchListTask"
	// 	}
	// }
	if ctx.IsAjax() {
		ctx.JSON(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"taskList":   sortedTaskList,
				"clientList": sortedClientList,
				"systemInfo": systemInfo,
				"taskIdList": strings.Join(taskIdSli, ","),
				"url":        r.RequestURI,
			},
		})
		return
	}

	ctx.ViewData("list", sortedTaskList)
	ctx.ViewData("addrs", sortedClientList)
	ctx.ViewData("client", clientList[addr])
	ctx.ViewData("systemInfo", systemInfo)
	ctx.ViewData("taskIds", strings.Join(taskIdSli, ","))
	ctx.ViewData("url", r.RequestURI)
	ctx.View("list.html")

}

// Index 服务器列表页面
func Index(ctx iris.Context) {
	var clientList map[string]proto.ClientConf
	var m = model.NewModel()

	sInfo := libs.SystemInfo(conf.ConfigArgs.ServerStartTime)

	sInfoL, sInfoR := make(map[string]interface{}), make(map[string]interface{})

	i := 0
	for k, v := range sInfo {
		if i <= 5 {
			sInfoL[k] = v
		} else {
			sInfoR[k] = v
		}
		i++
	}

	fmt.Printf("%+v", sInfo)

	clientList, _ = m.GetRPCClientList()
	sortedClientList := make([]proto.ClientConf, 0)

	for _, v := range clientList {
		sortedClientList = append(sortedClientList, v)
	}

	sort.Slice(sortedClientList, func(i, j int) bool {
		return (sortedClientList[i].Addr > sortedClientList[j].Addr) && (sortedClientList[i].State > sortedClientList[j].State)
	})
	ctx.Application().Logger().Debug(ctx.GetViewData())

	ctx.ViewData("clientList", sortedClientList)
	ctx.ViewData("systemInfoLeft", sInfoL)
	ctx.ViewData("systemInfoRight", sInfoR)

	fmt.Printf("%+v", sInfo)
	bts, _ := json.Marshal(sInfo)

	ctx.ViewData("systemInfoList", string(bts))

	ctx.View("index.html")

}

func UpdateTask(ctx iris.Context) {
	var reply bool
	var r = ctx.Request()
	var m = model.NewModel()

	sortedKeys := make([]string, 0)
	addr := strings.TrimSpace(r.FormValue("addr"))
	id := strings.TrimSpace(r.FormValue("taskId"))
	if addr == "" {
		// ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
		// 	"error": "params error",
		// })
		ctx.ViewData("error", "params error")
		ctx.View("public/error.html")
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

		if err := rpc.Call(addr, "Task.Update", proto.TaskArgs{
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
			ctx.ViewData("error", err.Error())
			ctx.View("public/error.html")
		}
		if reply {
			ctx.Redirect("/list?addr="+addr, http.StatusFound)
			return
		}

	} else {
		var t proto.TaskArgs
		var clientList map[string]proto.ClientConf

		if id != "" {
			err := rpc.Call(addr, "Task.Get", id, &t)
			if err != nil {
				ctx.Redirect("/list?addr="+addr, http.StatusFound)
				return

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

			ctx.ViewData("error", "nothing to show")
			ctx.View("public/error.html")
			return
		}

		ctx.ViewData("addr", addr)
		ctx.ViewData("addrs", sortedKeys)
		ctx.ViewData("rpcClientsMap", clientList)
		ctx.ViewData("task", t)
		ctx.ViewData("allowCommands", conf.ConfigArgs.AllowCommands)
		ctx.View("public/error.html")
	}

}

func StopTask(ctx iris.Context) {
	var r = ctx.Request()

	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	action := libs.ReplaceEmpty(r.FormValue("action"), "stop")
	var reply bool
	if taskId == "" || addr == "" {
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}

	var method string
	if action == "stop" {
		method = "Task.Stop"
	} else if action == "delete" {
		method = "Task.Delete"
	} else {
		method = "Task.Kill"
	}
	if err := rpc.Call(addr, method, taskId, &reply); err != nil {
		ctx.ViewData("error", err)
		ctx.View("public/error.html")
		return
	}
	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return
	}

	ctx.ViewData("error", fmt.Sprintf("failed %s %s", method, taskId))
	ctx.View("public/error.html")
}

func StopAllTask(ctx iris.Context) {
	var r = ctx.Request()

	taskIds := strings.TrimSpace(r.FormValue("taskIds"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	method := "Task.StopAll"
	taskIdSli := strings.Split(taskIds, ",")
	var reply bool
	if len(taskIdSli) == 0 || addr == "" {
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}

	if err := rpc.Call(addr, method, taskIdSli, &reply); err != nil {
		ctx.ViewData("error", err)
		ctx.View("public/error.html")
		return
	}
	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return
	}

	ctx.ViewData("error", fmt.Sprintf("failed %s %v", method, taskIdSli))
	ctx.View("public/error.html")
}

func StartTask(ctx iris.Context) {
	var r = ctx.Request()

	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply bool
	if taskId == "" || addr == "" {
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}

	if err := rpc.Call(addr, "Task.Start", taskId, &reply); err != nil {
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}

	if reply {
		ctx.Redirect("/list?addr="+addr, http.StatusFound)
		return
	}

	ctx.ViewData("error", "failed start task"+taskId)
	ctx.View("public/error.html")
}

func Login(ctx iris.Context) {
	var r = ctx.Request()
	if r.Method == http.MethodPost {

		u := r.FormValue("username")
		pwd := r.FormValue("passwd")
		remb := r.FormValue("remember")

		if u == conf.ConfigArgs.User && pwd == conf.ConfigArgs.Passwd {

			clientFeature := ctx.RemoteAddr() + "-" + ctx.Request().Header.Get("User-Agent")

			clientSign := fmt.Sprintf("%x", md5.Sum([]byte(clientFeature)))
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user":       u,
				"clientSign": clientSign,
			}).SignedString(conf.ConfigArgs.JWTSigningKey)

			if err != nil {
				ctx.ViewData("error", fmt.Sprint("无法生成访问凭证:", err))
				ctx.View("public/error.html")
				return
			}
			if remb == "yes" {
				ctx.SetCookieKV(conf.ConfigArgs.TokenCookieName, url.QueryEscape(token), iris.CookiePath("/"),
					iris.CookieExpires(time.Duration(conf.ConfigArgs.TokenExpires)), iris.CookieHTTPOnly(true))
			} else {
				ctx.SetCookieKV(conf.ConfigArgs.TokenCookieName, url.QueryEscape(token))
			}

			ctx.Redirect("/", http.StatusFound)
			return
		}

		ctx.ViewData("error", "auth failed")
		ctx.View("public/error.html")
		return
	}
	ctx.View("login.html")

}

func QuickStart(ctx iris.Context) {
	var r = ctx.Request()

	taskId := strings.TrimSpace(r.FormValue("taskId"))
	addr := strings.TrimSpace(r.FormValue("addr"))
	var reply []byte
	if taskId == "" || addr == "" {
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}

	if err := rpc.Call(addr, "Task.QuickStart", taskId, &reply); err != nil {
		ctx.ViewData("error", err.Error())
		ctx.View("public/error.html")
		return

	}
	logList := strings.Split(string(reply), "\n")

	ctx.ViewData("logList", logList)
	ctx.ViewData("addr", addr)
	ctx.View("log.html")
}

func Logout(ctx iris.Context) {
	// TOTO 清理token
	// ctx.CleanToken()
	ctx.Redirect("/login", http.StatusFound)

}

func RecentLog(ctx iris.Context) {
	var r = ctx.Request()

	id := r.FormValue("taskId")
	addr := r.FormValue("addr")
	var content []byte
	if id == "" {
		// ctx.RenderHtml([]string{"public/error"}, map[string]interface{}{
		// 	"error": "param error",
		// })
		ctx.ViewData("error", "param error")
		ctx.View("public/error.html")
		return
	}
	if err := rpc.Call(addr, "Task.Log", id, &content); err != nil {

		ctx.ViewData("error", err)
		ctx.View("public/error.html")
		return

	}
	logList := strings.Split(string(content), "\n")

	ctx.ViewData("logList", logList)
	ctx.ViewData("addr", addr)
	ctx.View("public/error.html")

}

func Readme(ctx iris.Context) {
	ctx.View("readme.html")
}

func ReloadConfig(ctx iris.Context) {
	conf.ConfigArgs.Reload()
	ctx.Redirect("/", http.StatusFound)

}

func DeleteClient(ctx iris.Context) {
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

}

func ViewConfig(ctx iris.Context) {

	c := conf.ConfigArgs.Category()
	r := ctx.Request()

	if r.Method == http.MethodPost {
		mailTo := strings.TrimSpace(r.FormValue("mailTo"))
		libs.SendMail("测试邮件", "测试邮件请勿回复", conf.ConfigArgs.MailHost, conf.ConfigArgs.MailUser, conf.ConfigArgs.MailPass, conf.ConfigArgs.MailPort, mailTo)
	}

	ctx.ViewData("configs", c)
	ctx.View("viewConfig.html")
}

func Model(ctx iris.Context) {
	val := ctx.FormValue("type")
	url := ctx.FormValue("url")
	ctx.SetCookie(&http.Cookie{
		Name:     "model",
		Path:     "/",
		Value:    val,
		HttpOnly: true,
	})

	ctx.Redirect(url, http.StatusFound)

}
