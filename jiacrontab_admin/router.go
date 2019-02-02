package admin

import (
	"path/filepath"

	"github.com/kataras/iris"

	"jiacrontab/pkg/proto"

	"jiacrontab/server/conf"
	"net/http"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/iwannay/jiaweb/utils/file"
)

func route(app *iris.Application) {
	app.StaticWeb(cfg.App.StaticDir, filepath.Join(file.GetCurrentDirectory(), cfg.App.StaticDir))

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Jwt.SigningKey), nil
		},

		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetCookie(conf.JwtService.TokenCookieName))
			return token, err
		},

		ErrorHandler: func(ctx iris.Context, data string) {
			app.Logger().Error("jwt 认证失败", data)
			if ctx.RequestPath(true) != "/login" {
				ctx.Redirect("/login", http.StatusFound)
				ctx.JSON(respError(proto.Code_FailedAuth, "auth failed", nil))
				return
			}

			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.Use(jwtHandler.Serve, func(ctx iris.Context) {
		token := ctx.Values().Get("jwt").(*jwt.Token)
		ctx.Values().Set("sess", token.Claims)
		ctx.Next()
	})

	app.Post("/client/list", getClientList)
	app.Post("/crontab/job/list", getJobList)
	app.Post("/crontab/job/log", getRecentLog)
	app.Post("/crontab/job/edit", editJob)
	app.Post("/crontab/job/stop", stopTask)
	app.Post("/crontab/job/start", startTask)
	app.post("/crontab/job/exec", execTask)

	app.Post("/user/login", login)
	app.Get("/crontab/job/exec", exec)

	// app.Get("/reloadConfig", handle.ReloadConfig)
	// app.Get("/deleteClient", handle.DeleteClient)
	// app.Any("/viewConfig", handle.ViewConfig)
	// app.Get("/crontab/task/stopAll", handle.StopAllTask)

	// app.Any("/daemon/task/list", handle.ListDaemonTask)
	// app.Get("/daemon/task/action", handle.ActionDaemonTask)
	// app.Any("/daemon/task/edit", handle.EditDaemonTask)
	// app.Get("/daemon/task/log", handle.RecentDaemonLog)

	// app.Get("/runtime/info", handle.RuntimeInfo)

	// debug := app.Party("/debug")
	// {
	// 	debug.Get("/stat", handle.Stat)
	// 	debug.Get("/pprof/", handle.IndexDebug)
	// 	debug.Get("/pprof/{key:string}", handle.PprofHandler)
	// 	debug.Get("/freemem", handle.FreeMem)

	// }

}
