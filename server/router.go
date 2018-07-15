package main

import (
	"path/filepath"

	"strings"

	"github.com/kataras/iris"

	"jiacrontab/server/handle"

	"runtime"

	"jiacrontab/server/conf"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/iwannay/jiaweb/utils/file"
)

func strFirstUpper(str string) string {
	if str == "" {
		return str
	}
	sli := strings.SplitN(str, "", 2)

	if len(sli) == 2 {
		return strings.ToUpper(sli[0]) + sli[1]
	}
	return strings.ToUpper(sli[0])
}

func notFound(ctx iris.Context) {
	ctx.View("public/404.html")
}

func catchError(ctx iris.Context) {
	ctx.ViewData("error", "服务不可用")
	ctx.View("public/error.html")
}

func router(app *iris.Application) {
	app.StaticWeb(conf.AppService.StaticDir, filepath.Join(file.GetCurrentDirectory(), conf.AppService.StaticDir))

	app.OnAnyErrorCode(catchError)
	app.OnErrorCode(iris.StatusNotFound, notFound)

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(conf.JwtService.JWTSigningKey), nil
		},
		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetCookie(conf.JwtService.TokenCookieName))
			return token, err
		},

		ErrorHandler: func(ctx iris.Context, data string) {
			app.Logger().Error("jwt 认证失败", data)

			if ctx.RequestPath(true) != "/login" {
				ctx.Redirect("/login", http.StatusFound)
				return
			}

			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.Use(func(ctx iris.Context) {

		path := ctx.Request().URL.Path
		ctx.ViewData("action", filepath.Base(path))
		ctx.ViewData("controller", strings.Replace(filepath.Dir(path), `\`, `/`, -1))
		ctx.ViewData("title", "jiacrontab")
		ctx.ViewData("goVersion", runtime.Version())
		ctx.ViewData("appVersion", conf.Version)
		ctx.ViewData("requestPath", ctx.Request().URL.Path)
		ctx.ViewData("staticDir", conf.AppService.StaticDir)

		ctx.ViewData("addr", ctx.FormValue("addr"))
		ctx.Next()
	})

	app.Use(jwtHandler.Serve, func(ctx iris.Context) {
		token, ok := ctx.Values().Get("jwt").(*jwt.Token)
		if ok {
			ctx.ViewData("user", token.Claims)
		}
		ctx.Next()
	})

	app.Get("/", handle.Index)

	app.Any("/crontab/task/list", handle.ListTask)
	app.Get("/crontab/task/log", handle.RecentLog)
	app.Any("/crontab/task/edit", handle.EditTask)
	app.Get("/crontab/task/stop", handle.StopTask)
	app.Get("/crontab/task/start", handle.StartTask)
	app.Any("/login", handle.Login)

	app.Get("/logout", handle.Logout)
	app.Get("/readme", handle.Readme)
	app.Get("/crontab/task/quickStart", handle.QuickStart)
	app.Get("/reloadConfig", handle.ReloadConfig)
	app.Get("/deleteClient", handle.DeleteClient)
	app.Any("/viewConfig", handle.ViewConfig)
	app.Get("/crontab/task/stopAll", handle.StopAllTask)

	app.Any("/daemon/task/list", handle.ListDaemonTask)
	app.Get("/daemon/task/action", handle.ActionDaemonTask)
	app.Any("/daemon/task/edit", handle.EditDaemonTask)
	app.Get("/daemon/task/log", handle.RecentDaemonLog)

	app.Get("/runtime/info", handle.RuntimeInfo)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", handle.Stat)
		debug.Get("/pprof/", handle.IndexDebug)
		debug.Get("/pprof/{key:string}", handle.PprofHandler)
		debug.Get("/freemem", handle.FreeMem)

	}

}
