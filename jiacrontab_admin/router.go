package admin

import (
	"fmt"
	"jiacrontab/pkg/base"
	"path/filepath"

	"github.com/kataras/iris"

	"jiacrontab/pkg/proto"

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
			token, err := url.QueryUnescape(ctx.GetCookie(cfg.Jwt.Name))
			return token, err
		},

		ErrorHandler: func(c iris.Context, data string) {
			ctx := wrapCtx(c)
			app.Logger().Error("jwt 认证失败", data)
			if ctx.RequestPath(true) != "/user/login" {
				ctx.respError(proto.Code_FailedAuth, "认证失败", nil)
				return
			}
			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.UseGlobal(func(ctx iris.Context) {
		base.Stat.AddConcurrentCount()
		defer func() {
			if err := recover(); err != nil {
				base.Stat.AddErrorCount(ctx.RequestPath(true), fmt.Errorf("%v", err), 1)
			}
			base.Stat.AddRequestCount(ctx.RequestPath(true), ctx.GetStatusCode(), 1)
		}()
		ctx.Next()
	})

	app.Use(jwtHandler.Serve, func(ctx iris.Context) {
		token := ctx.Values().Get("jwt").(*jwt.Token)
		ctx.Values().Set("sess", token.Claims)
		ctx.Next()
	})

	app.Post("/node/list", getNodeList)
	app.Post("/node/delete", deleteNode)
	app.Post("/crontab/job/list", getJobList)
	app.Post("/crontab/job/get", getJob)
	app.Post("/crontab/job/log", getRecentLog)
	app.Post("/crontab/job/edit", editJob)
	app.Post("/crontab/job/stop", stopTask)
	app.Post("/crontab/job/start", startTask)
	app.Post("/crontab/job/exec", execTask)

	app.Post("/user/login", login)
	app.Post("/config/get", getConfig)
	app.Post("/runtime/info", runtimeInfo)

	app.Post("/daemon/job/list", getDaemonJobList)
	app.Post("/daemon/job/action", actionDaemonTask)
	app.Post("/daemon/job/edit", editDaemonJob)
	app.Post("/daemon/job/get", getDaemonJob)
	app.Post("/daemon/job/log", getRecentDaemonLog)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", stat)
		debug.Get("/pprof/", indexDebug)
		debug.Get("/pprof/{key:string}", pprofHandler)
	}

}
