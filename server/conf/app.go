package conf

import (
	"strings"
	"time"
)

const (
	AppName = "jiacrontab"
	Version = "v1.4.7"
)

var (
	AppService *App
	starTime   = time.Now()
)

type App struct {
	Debug           bool      `json:"debug"`
	HttpListenAddr  string    `json:"http_listen_addr"`
	StaticDir       string    `json:"static_dir"`
	DataFile        string    `json:"data_file"`
	RpcListenAddr   string    `json:"rpc_listen_addr"`
	ServerStartTime time.Time `json:"-"`
	TplDir          string    `json:"tpl_dir"`
	TplExt          string    `json:"tpl_ext"`

	User          string   `json:"user"`
	Passwd        string   `json:"-"`
	AllowCommands []string `json:"allow_commands"`
}

func LoadAppService() {
	app := cf.Section("app")
	AppService = &App{
		Debug:           app.Key("debug").MustBool(false),
		HttpListenAddr:  app.Key("http_listen_addr").MustString("0.0.0.0:20000"),
		StaticDir:       app.Key("static_dir").MustString("/static"),
		TplExt:          ".html",
		TplDir:          "template",
		DataFile:        app.Key("data_file").MustString("data.json"),
		RpcListenAddr:   app.Key("rpc_listen_addr").MustString(":20003"),
		User:            app.Key("app_user").MustString("admin"),
		Passwd:          app.Key("app_passwd").MustString("123456"),
		ServerStartTime: starTime,
		AllowCommands:   strings.Split(app.Key("allow_commands").MustString("php,python,node,curl,wget,lua"), ","),
	}
}
