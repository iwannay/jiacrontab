package main

import (
	"path/filepath"

	"strings"

	"github.com/kataras/iris"

	"jiacrontab/server/routes"

	"fmt"
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

func h(ctx iris.Context) {

	user := ctx.Values().Get("jwt").(*jwt.Token)
	ctx.WriteString(fmt.Sprintf("%+v", user))
}

func router(app *iris.Application) {
	app.StaticWeb("/static", filepath.Join(file.GetCurrentDirectory(), "static"))

	app.OnAnyErrorCode(catchError)
	app.OnErrorCode(iris.StatusNotFound, notFound)

	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return conf.ConfigArgs.JWTSigningKey, nil
		},
		Extractor: func(ctx iris.Context) (string, error) {
			token, err := url.QueryUnescape(ctx.GetCookie(conf.ConfigArgs.TokenCookieName))
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
		ctx.ViewData("appVersion", "v1.3.5")
		ctx.ViewData("requestPath", ctx.Request().URL.Path)
		ctx.ViewData("staticDir", "static")
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

	app.Get("/", routes.Index)

	app.Get("/crontab/task/list", routes.ListTask)
	app.Get("/crontab/task/log", routes.RecentLog)
	app.Any("/crontab/task/edit", routes.EditTask)
	app.Get("/crontab/task/stop", routes.StopTask)
	app.Get("/crontab/task/start", routes.StartTask)
	app.Any("/login", routes.Login)
	app.Get("/logout", routes.Logout)
	app.Get("/readme", routes.Readme)
	app.Get("/crontab/task/quickStart", routes.QuickStart)
	app.Get("/reloadConfig", routes.ReloadConfig)
	app.Get("/deleteClient", routes.DeleteClient)
	app.Get("/viewConfig", routes.ViewConfig)
	app.Get("/crontab/task/stopAll", routes.StopAllTask)

}
