package conf

import (
	"jiacrontab/pkg/util"

	"gopkg.in/ini.v1"
)

const configFile = "server.ini"

var (
	cf *ini.File
)

func Reload() {
	cf = loadConfig()
	LoadMailService()
	LoadJwtService()
	LoadAppService()
}

func Category() map[string]interface{} {
	mail := make(map[string]interface{})
	app := make(map[string]interface{})
	cat := make(map[string]interface{})
	jwt := make(map[string]interface{})
	util.Struct2Map(MailService, &mail)
	util.Struct2Map(AppService, &app)
	util.Struct2Map(JwtService, &jwt)

	cat["app"] = app
	cat["jwt"] = jwt
	cat["mail"] = mail
	return cat
}

func loadConfig() *ini.File {
	f, err := ini.Load(configFile)
	if err != nil {
		panic(err)
	}
	return f
}

func init() {
	Reload()
}
