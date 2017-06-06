package main

import (
	"strings"

	"io/ioutil"

	"fmt"

	"gopkg.in/ini.v1"
)

const configFile = "server.ini"
const versionFile = "template/.VERSION"

type config struct {
	debug             bool
	addr              string
	staticDir         string
	defaultFaviconDir string
	tplExt            string
	tplDir            string
	dataFile          string
	rpcAddr           string

	JWTSigningKey     []byte
	tokenCookieName   string
	tokenExpires      int64
	tokenCookieMaxAge int64

	appName             string
	user                string
	passwd              string
	allowCommands       []string
	defaultRPCPath      string
	defaultRPCDebugPath string
	mailUser            string
	mailHost            string
	mailPass            string
	mailPort            string
	version             string
}

func newConfig() *config {
	c := &config{}
	c.reload()
	return c

}

func (c *config) reload() {
	cf := loadConfig()
	base := cf.Section("base")
	srvc := cf.Section("server")
	jwt := cf.Section("jwt")
	rpc := cf.Section("rpc")
	script := cf.Section("script")
	mail := cf.Section("mail")

	c.debug = base.Key("debug").MustBool(false)
	c.addr = srvc.Key("listen").MustString("0.0.0.0:20000")
	c.staticDir = srvc.Key("static_dir").MustString("/static")
	c.tplExt = ".html"
	c.tplDir = "template"
	c.dataFile = srvc.Key("data_file").MustString("data.json")
	c.defaultFaviconDir = srvc.Key("favicon").MustString("favicon.ico")
	c.JWTSigningKey = []byte(jwt.Key("signing_key").MustString("eyJhbGciOiJIUzI1"))
	c.rpcAddr = rpc.Key("listen").MustString(":20003")
	c.tokenCookieName = jwt.Key("token_cookie_name").MustString("access_token")
	c.tokenExpires = jwt.Key("expires").MustInt64(3000)
	c.tokenCookieMaxAge = jwt.Key("token_cookie_maxage").MustInt64(3000)
	c.user = base.Key("app_user").MustString("john")
	c.passwd = base.Key("app_passwd").MustString("john")
	c.appName = base.Key("app_name").MustString("jiacrontab")
	c.allowCommands = strings.Split(script.Key("allow_commands").MustString("php,/usr/local/bin/php,python,node,curl,wget,lua"), ",")
	c.defaultRPCDebugPath = "/debug/rpc"
	c.defaultRPCPath = "/__myrpc__"
	b, _ := ioutil.ReadFile(versionFile)
	c.mailHost = mail.Key("host").MustString("")
	c.mailUser = mail.Key("user").MustString("")
	c.mailPass = mail.Key("pass").MustString("")
	c.mailPort = mail.Key("port").MustString("25")
	c.version = string(b)
}

func (c *config) category() map[string]map[string]string {
	cat := make(map[string]map[string]string)

	cat["base"] = map[string]string{
		"version": c.version,
		"appName": c.appName,
	}
	cat["jwt"] = map[string]string{
		"JWTSigningKey":   string(c.JWTSigningKey),
		"tokenCookieName": c.tokenCookieName,
		"tokenExpires":    fmt.Sprintf("%d", c.tokenExpires),
	}

	cat["mail"] = map[string]string{
		"mailUser": c.mailUser,
		"mailHost": c.mailUser + ":" + c.mailPort,
	}

	cat["server"] = map[string]string{
		"listen":    c.addr,
		"staticDir": c.staticDir,
		"tplExt":    c.tplExt,
		"tplDir":    c.tplDir,
	}

	cat["rpc"] = map[string]string{
		"listen":              c.rpcAddr,
		"defaultRPCPath":      c.defaultRPCPath,
		"defaultRPCDebugPath": c.defaultRPCPath,
	}
	return cat
}

func loadConfig() *ini.File {
	f, err := ini.Load(configFile)
	if err != nil {
		panic(err)
	}
	return f
}
