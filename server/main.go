package main

import (
	"jiacrontab/libs/rpc"
	"jiacrontab/server/conf"
	"jiacrontab/server/handle"
	"jiacrontab/server/model"
	_ "net/http/pprof"

	"github.com/kataras/iris/middleware/recover"

	"jiacrontab/libs"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
)

const (
	DefaultTitle  = "jiacrontab"
	DefaultLayout = "layouts/layout.html"
)

var config *conf.Config

func main() {

	model.InitStore(conf.ConfigArgs.DataFile)

	app := iris.New()
	app.Logger().SetLevel("debug")

	app.Use(recover.New())
	app.Use(logger.New())

	html := iris.HTML(conf.ConfigArgs.TplDir, conf.ConfigArgs.TplExt)
	html.AddFunc("date", libs.Date)
	html.AddFunc("formatMs", libs.Int2floatstr)
	html.Layout("layouts/layout.html")
	html.Reload(true)
	app.RegisterView(html)
	router(app)
	go rpc.ListenAndServe(conf.ConfigArgs.RpcAddr, &handle.Logic{})
	app.Run(iris.Addr(":20000"))
}
