package conf

import (
	"strings"

	"io/ioutil"

	"fmt"

	"gopkg.in/ini.v1"
)

const configFile = "server.ini"
const versionFile = "template/.VERSION"

var ConfigArgs *Config

type Config struct {
	Debug             bool
	Addr              string
	StaticDir         string
	DefaultFaviconDir string
	TplExt            string
	TplDir            string
	DataFile          string
	RpcAddr           string

	JWTSigningKey     []byte
	TokenCookieName   string
	TokenExpires      int64
	TokenCookieMaxAge int64

	AppName             string
	User                string
	Passwd              string
	AllowCommands       []string
	DefaultRPCPath      string
	DefaultRPCDebugPath string
	MailUser            string
	MailHost            string
	MailPass            string
	MailPort            string
	Version             string
}

func NewConfig() *Config {
	c := &Config{}
	c.Reload()
	return c

}

func init() {
	ConfigArgs = NewConfig()
}

func InitConfig() {
	ConfigArgs = NewConfig()
}

func (c *Config) Reload() {
	cf := loadConfig()
	base := cf.Section("base")
	srvc := cf.Section("server")
	jwt := cf.Section("jwt")
	rpc := cf.Section("rpc")
	script := cf.Section("script")
	mail := cf.Section("mail")

	c.Debug = base.Key("debug").MustBool(false)
	c.Addr = srvc.Key("listen").MustString("0.0.0.0:20000")
	c.StaticDir = srvc.Key("static_dir").MustString("/static")
	c.TplExt = ".html"
	c.TplDir = "template"
	c.DataFile = srvc.Key("data_file").MustString("data.json")
	c.DefaultFaviconDir = srvc.Key("favicon").MustString("favicon.ico")
	c.JWTSigningKey = []byte(jwt.Key("signing_key").MustString("eyJhbGciOiJIUzI1"))
	c.RpcAddr = rpc.Key("listen").MustString(":20003")
	c.TokenCookieName = jwt.Key("token_cookie_name").MustString("access_token")
	c.TokenExpires = jwt.Key("expires").MustInt64(3000)
	c.TokenCookieMaxAge = jwt.Key("token_cookie_maxage").MustInt64(3000)
	c.User = base.Key("app_user").MustString("john")
	c.Passwd = base.Key("app_passwd").MustString("john")
	c.AppName = base.Key("app_name").MustString("jiacrontab")
	c.AllowCommands = strings.Split(script.Key("allow_commands").MustString("php,/usr/local/bin/php,python,node,curl,wget,lua"), ",")
	c.DefaultRPCDebugPath = "/debug/rpc"
	c.DefaultRPCPath = "/__myrpc__"
	b, _ := ioutil.ReadFile(versionFile)
	c.MailHost = mail.Key("host").MustString("")
	c.MailUser = mail.Key("user").MustString("")
	c.MailPass = mail.Key("pass").MustString("")
	c.MailPort = mail.Key("port").MustString("25")
	c.Version = string(b)
}

func (c *Config) Category() map[string]map[string]string {
	cat := make(map[string]map[string]string)

	cat["base"] = map[string]string{
		"version": c.Version,
		"appName": c.AppName,
	}
	cat["jwt"] = map[string]string{
		"JWTSigningKey":   string(c.JWTSigningKey),
		"tokenCookieName": c.TokenCookieName,
		"tokenExpires":    fmt.Sprintf("%d", c.TokenExpires),
	}

	cat["mail"] = map[string]string{
		"mailUser": c.MailUser,
		"mailHost": c.MailUser + ":" + c.MailPort,
	}

	cat["server"] = map[string]string{
		"listen":    c.Addr,
		"staticDir": c.StaticDir,
		"tplExt":    c.TplExt,
		"tplDir":    c.TplDir,
	}

	cat["rpc"] = map[string]string{
		"listen":              c.RpcAddr,
		"defaultRPCPath":      c.DefaultRPCPath,
		"defaultRPCDebugPath": c.DefaultRPCPath,
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
