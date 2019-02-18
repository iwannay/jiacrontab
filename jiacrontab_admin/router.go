package admin

import (
	"jiacrontab/pkg/proto"
	"net/url"
	"path/filepath"

	"github.com/kataras/iris"

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
			token, err := url.QueryUnescape(ctx.GetHeader(cfg.Jwt.Name))
			return token, err
		},
		Expiration: true,

		ErrorHandler: func(c iris.Context, data string) {
			ctx := wrapCtx(c)
			app.Logger().Error("jwt 认证失败:", data)

			if ctx.RequestPath(true) != "/user/login" && ctx.RequestPath(true) != "/user/signUp" {
				ctx.respError(proto.Code_FailedAuth, "认证失败", nil)
				return
			}
			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.UseGlobal(newRecover())

	adm := app.Party("/adm")
	{
		adm.Use(jwtHandler.Serve)

		adm.Post("/node/list", getNodeList)
		adm.Post("/node/delete", deleteNode)
		adm.Post("/crontab/job/list", getJobList)
		adm.Post("/crontab/job/get", getJob)
		adm.Post("/crontab/job/log", getRecentLog)
		adm.Post("/crontab/job/edit", editJob)
		adm.Post("/crontab/job/stop", actionTask)
		adm.Post("/crontab/job/start", startTask)
		adm.Post("/crontab/job/exec", execTask)

		adm.Post("/config/get", getConfig)
		adm.Post("/runtime/info", runtimeInfo)

		adm.Post("/daemon/job/list", getDaemonJobList)
		adm.Post("/daemon/job/action", actionDaemonTask)
		adm.Post("/daemon/job/edit", editDaemonJob)
		adm.Post("/daemon/job/get", getDaemonJob)
		adm.Post("/daemon/job/log", getRecentDaemonLog)

		adm.Post("/group/list", getGroupList)
		adm.Post("/group/edit", editGroup)
		adm.Post("/group/set", setGroup)

		adm.Post("/user/activity_list", getRelationEvent)
	}

	app.Post("/user/login", login)
	app.Post("/user/signUp", signUp)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", stat)
		debug.Get("/pprof/", indexDebug)
		debug.Get("/pprof/{key:string}", pprofHandler)
	}

}
