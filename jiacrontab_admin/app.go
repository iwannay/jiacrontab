package admin

import (
	"errors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"net/url"

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
				ctx.respAuthFailed(errors.New("认证失败"))
				return
			}
			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,Content-Type")
		ctx.Next()
	}

	adm := app.Party("/adm",crs).AllowMethods(iris.MethodOptions)
	{
		adm.Use(jwtHandler.Serve)
		adm.Post("/crontab/job/list", GetJobList)
		adm.Post("/crontab/job/get", GetJob)
		adm.Post("/crontab/job/log", GetRecentLog)
		adm.Post("/crontab/job/edit", EditJob)
		adm.Post("/crontab/job/action", ActionTask)
		adm.Post("/crontab/job/exec", ExecTask)

		adm.Post("/config/get", GetConfig)
		adm.Post("/system/info", SystemInfo)

		adm.Post("/daemon/job/list", GetDaemonJobList)
		adm.Post("/daemon/job/action", ActionDaemonTask)
		adm.Post("/daemon/job/edit", EditDaemonJob)
		adm.Post("/daemon/job/get", GetDaemonJob)
		adm.Post("/daemon/job/log", GetRecentDaemonLog)

		adm.Post("/group/list", GetGroupList)
		adm.Post("/group/edit", EditGroup)

		adm.Post("/node/list", GetNodeList)
		adm.Post("/node/delete", DeleteNode)
		adm.Post("/node/group_node", GroupNode)

		adm.Post("/user/activity_list", GetActivityList)
		adm.Post("/user/job_history", GetJobHistory)
		adm.Post("/user/audit_job", AuditJob)
		adm.Post("/user/stat", UserStat)
		adm.Post("/user/signup", Signup)
		adm.Post("/user/group_user", GroupUser)
		adm.Post("/user/list", GetUserList)
	}

	app.Post("/user/login", Login)
	app.Post("/user/init_admin_user", IninAdminUser)

	debug := app.Party("/debug")
	{
		debug.Get("/stat", stat)
		debug.Get("/pprof/", indexDebug)
		debug.Get("/pprof/{key:string}", pprofHandler)
	}

	return app
}
