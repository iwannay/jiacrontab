package admin

import (
	"jiacrontab/pkg/proto"
	"net/url"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"

	jwt "github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

func newApp() *iris.Application {

	app := iris.New()
	app.UseGlobal(newRecover())
	app.Logger().SetLevel("debug")
	app.Use(logger.New())

	app.RegisterView(iris.HTML("./public", ".html"))

	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	assetHandler := app.StaticHandler("./public", true, true)

	app.SPA(assetHandler).AddIndexName("index.html")

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

	adm := app.Party("/adm")
	{
		adm.Use(jwtHandler.Serve)

		adm.Post("/node/list", GetNodeList)
		adm.Post("/node/delete", DeleteNode)
		adm.Post("/node/group_node", GroupNode)

		adm.Post("/crontab/job/list", getJobList)
		adm.Post("/crontab/job/get", getJob)
		adm.Post("/crontab/job/log", getRecentLog)
		adm.Post("/crontab/job/edit", editJob)
		adm.Post("/crontab/job/action", actionTask)
		adm.Post("/crontab/job/exec", execTask)

		adm.Post("/config/get", getConfig)
		adm.Post("/runtime/info", runtimeInfo)

		adm.Post("/daemon/job/list", getDaemonJobList)
		adm.Post("/daemon/job/action", actionDaemonTask)
		adm.Post("/daemon/job/edit", editDaemonJob)
		adm.Post("/daemon/job/get", getDaemonJob)
		adm.Post("/daemon/job/log", getRecentDaemonLog)

		adm.Post("/group/list", GetGroupList)
		adm.Post("/group/edit", EditGroup)
		adm.Post("/group/set", SetGroup)

		adm.Post("/user/activity_list", getRelationEvent)
		// adm.Post("/user/job_history",getJobHistory)
		adm.Post("/user/auditJob", auditJob)
	}

	app.Post("/user/login", login)
	app.Post("/user/signUp", signUp)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", stat)
		debug.Get("/pprof/", indexDebug)
		debug.Get("/pprof/{key:string}", pprofHandler)
	}

	return app
}
