package admin

import (
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"

	"jiacrontab/models"

	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/iris-contrib/middleware/cors"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/kataras/iris/context"
)

func newApp(adm *Admin) *iris.Application {

	app := iris.New()
	app.UseGlobal(newRecover(adm))
	app.Logger().SetLevel(adm.getOpts().App.LogLevel)
	app.Use(logger.New())
	app.StaticEmbeddedGzip("/", "./assets/", GzipAsset, GzipAssetNames)

	cfg := adm.getOpts()

	wrapHandler := func(h func(ctx *myctx)) context.Handler {
		return func(c iris.Context) {
			h(wrapCtx(c, adm))
		}
	}

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Jwt.SigningKey), nil
		},

		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetHeader(cfg.Jwt.Name))
			return token, err
		},
		Expiration: true,

		ErrorHandler: func(c iris.Context, err error) {
			ctx := wrapCtx(c, adm)
			if ctx.RequestPath(true) != "/user/login" {
				ctx.respAuthFailed(fmt.Errorf("Token verification failed(%s)", err))
				return
			}
			ctx.Next()
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	crs := cors.New(cors.Options{
		Debug:            true,
		AllowedHeaders:   []string{"Content-Type", "Token"},
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
	})

	app.Use(crs)
	app.AllowMethods(iris.MethodOptions)
	app.Get("/", func(ctx iris.Context) {
		if atomic.LoadInt32(&adm.initAdminUser) == 1 {
			ctx.SetCookieKV("ready", "true", func(c *http.Cookie) {
				c.HttpOnly = false
			})
		} else {
			ctx.SetCookieKV("ready", "false", func(c *http.Cookie) {
				c.HttpOnly = false
			})
		}
		ctx.Header("Cache-Control", "no-cache")
	})

	v1 := app.Party("/v1")
	{
		v1.Post("/user/login", wrapHandler(Login))
		v1.Post("/app/init", wrapHandler(InitApp))
	}

	v2 := app.Party("/v2")
	{
		v2.Use(jwtHandler.Serve)
		v2.Use(wrapHandler(func(ctx *myctx) {
			if err := ctx.parseClaimsFromToken(); err != nil {
				ctx.respJWTError(err)
				return
			}
			ctx.Next()
		}))
		v2.Post("/crontab/job/list", wrapHandler(GetJobList))
		v2.Post("/crontab/job/get", wrapHandler(GetJob))
		v2.Post("/crontab/job/log", wrapHandler(GetRecentLog))
		v2.Post("/crontab/job/edit", wrapHandler(EditJob))
		v2.Post("/crontab/job/action", wrapHandler(ActionTask))
		v2.Post("/crontab/job/exec", wrapHandler(ExecTask))

		v2.Post("/config/get", wrapHandler(GetConfig))
		v2.Post("/config/mail/send", wrapHandler(SendTestMail))
		v2.Post("/system/info", wrapHandler(SystemInfo))

		v2.Post("/daemon/job/list", wrapHandler(GetDaemonJobList))
		v2.Post("/daemon/job/action", wrapHandler(ActionDaemonTask))
		v2.Post("/daemon/job/edit", wrapHandler(EditDaemonJob))
		v2.Post("/daemon/job/get", wrapHandler(GetDaemonJob))
		v2.Post("/daemon/job/log", wrapHandler(GetRecentDaemonLog))

		v2.Post("/group/list", wrapHandler(GetGroupList))
		v2.Post("/group/edit", wrapHandler(EditGroup))

		v2.Post("/node/list", wrapHandler(GetNodeList))
		v2.Post("/node/delete", wrapHandler(DeleteNode))
		v2.Post("/node/group_node", wrapHandler(GroupNode))

		v2.Post("/user/activity_list", wrapHandler(GetActivityList))
		v2.Post("/user/job_history", wrapHandler(GetJobHistory))
		v2.Post("/user/audit_job", wrapHandler(AuditJob))
		v2.Post("/user/stat", wrapHandler(UserStat))
		v2.Post("/user/signup", wrapHandler(Signup))
		v2.Post("/user/edit", wrapHandler(EditUser))
		v2.Post("/user/delete", wrapHandler(DeleteUser))
		v2.Post("/user/group_user", wrapHandler(GroupUser))
		v2.Post("/user/list", wrapHandler(GetUserList))
	}

	debug := app.Party("/debug")
	{
		debug.Get("/stat", wrapHandler(stat))
		debug.Get("/pprof/", wrapHandler(indexDebug))
		debug.Get("/pprof/{key:string}", wrapHandler(pprofHandler))
	}

	return app
}

// InitApp 初始化应用
func InitApp(ctx *myctx) {
	var (
		err     error
		user    models.User
		reqBody InitAppReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if ret := models.DB().Take(&user, "group_id=?", 1); ret.Error == nil && ret.RowsAffected > 0 {
		ctx.respNotAllowed()
		return
	}

	user.Username = reqBody.Username
	user.Passwd = reqBody.Passwd
	user.Root = true
	user.GroupID = models.SuperGroup.ID
	user.Mail = reqBody.Mail

	if err = user.Create(); err != nil {
		ctx.respBasicError(err)
		return
	}
	atomic.StoreInt32(&ctx.adm.initAdminUser, 1)
	ctx.respSucc("", true)
}
