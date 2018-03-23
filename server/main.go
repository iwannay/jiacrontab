package main

import (
	"html/template"
	"jiacrontab/libs"
	"jiacrontab/server/conf"
	"jiacrontab/server/middleware"
	"jiacrontab/server/model"
	"jiacrontab/server/routes"
	"jiacrontab/server/rpc"
	_ "net/http/pprof"
	"runtime"

	jiawebConfig "github.com/iwannay/jiaweb/config"

	"github.com/iwannay/jiaweb"
)

var config *conf.Config

// var globalStore *store.Store

// var globalJwt *mjwt
// var globalReqFilter *reqFilter
// var startTime = time.Now()

func notFoundHandle(ctx jiaweb.Context) {
	ctx.RenderHtml([]string{"public/404"}, nil)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// globalJwt = newJwt(globalConfig.tokenExpires, globalConfig.tokenCookieName, globalConfig.JWTSigningKey, globalConfig.tokenCookieMaxAge)
	config = conf.ConfigArgs
	model.InitStore(config.DataFile)

	app := jiaweb.Classic(func(app *jiaweb.JiaWeb) {
		app.SetLogPath("logsfile")
		app.SetEnableLog(true)
		app.SetEnableDetailRequestData(true)
		app.SetPProfConfig(true, 10004)
		app.SetNotFoundHandle(notFoundHandle)
	})

	router(app)
	app.HttpServer.SetEnableIgnoreFavicon(true)
	app.HttpServer.SetEnableJwt(&jiawebConfig.JwtNode{
		Expire:       86400,
		Name:         "session-id",
		EnableJwt:    true,
		CookieMaxAge: 86400,
		SignKey:      "ASDASDFASDFIFIFJA234243",
	})

	app.HttpServer.RegisterModule(&jiaweb.HttpModule{
		Name: "initRender",
		OnBeginRequest: func(ctx jiaweb.Context) {
			origin := ctx.Response().QueryHeader("Origin")
			if origin == "" {
				origin = "http://dev.iwannay.cn"
			}

			ctx.HttpServer().Render.AddLocals(jiaweb.KValue{
				Key:   "requestPath",
				Value: ctx.Request().Path(),
			}, jiaweb.KValue{
				Key:   "staticDir",
				Value: "static",
			}, jiaweb.KValue{
				Key:   "title",
				Value: "jiacrontab",
			}, jiaweb.KValue{
				Key:   "site",
				Value: "jiacrontab.iwannay.cn",
			}, jiaweb.KValue{
				Key:   "appName",
				Value: "jiacrontab",
			}, jiaweb.KValue{
				Key:   "appVersion",
				Value: "v1.3.5",
			})
			ctx.HttpServer().Render.AppendTpl("public/head", "public/foot")
			ctx.Store().Store("appConfig", config)
		},
	})
	// app.Use(&middleware.CrossAllowMiddleware{})
	app.Use(&middleware.AuthMiddleware{})
	app.HttpServer.Render.AppendFunc(template.FuncMap{
		"date":     libs.Date,
		"formatMs": libs.Int2floatstr,
	})

	app.StartServer(20000)

	// globalReqFilter = newReqFilter()

	go rpc.InitSrvRpc(config.DefaultRPCPath, config.DefaultRPCDebugPath, config.RpcAddr, &routes.Logic{})

	// initServer()
}
