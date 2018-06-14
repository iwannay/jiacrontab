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

//func registerController(i interface{}) jiaweb.HttpHandle {
//	return func(ctx jiaweb.Context) error {
//		action := strFirstUpper(ctx.QueryRouteParam("key"))
//		t := reflect.TypeOf(i)
//		if t.Kind() == reflect.Func {
//			v := reflect.ValueOf(i)
//			objSlice := v.Call([]reflect.Value{})
//			if len(objSlice) >= 1 {
//				v = objSlice[0]
//				m := v.MethodByName(action)
//				if m.IsValid() {
//					errSli := m.Call([]reflect.Value{reflect.ValueOf(ctx)})
//					if errSli[0].IsNil() {
//						return nil
//					}
//					err := errSli[0].Interface().(error)
//					return err
//				}
//			}
//
//		}
//
//		ctx.NotFound()
//
//		return nil
//	}
//}

func h(ctx iris.Context) {

	user := ctx.Values().Get("jwt").(*jwt.Token)
	ctx.WriteString(fmt.Sprintf("%+v", user))
}

func router(app *iris.Application) {
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
			ctx.Redirect("/login", http.StatusFound)
		},

		SigningMethod: jwt.SigningMethodHS256,
	})

	app.Use(func(ctx iris.Context) {

		ctx.ViewData("action", ctx.Request().URL.Path)
		ctx.ViewData("title", "jiacrontab")
		ctx.ViewData("goVersion", runtime.Version())
		ctx.ViewData("appVersion", "v1.3.5")
		ctx.ViewData("requestPath", ctx.Request().URL.Path)
		ctx.ViewData("staticDir", "static")
		ctx.Next()
	})

	app.StaticWeb("/static", filepath.Join(file.GetCurrentDirectory(), "static"))
	admin := app.Party("/admin", jwtHandler.Serve, func(ctx iris.Context) {
		token := ctx.Values().Get("jwt").(*jwt.Token)
		ctx.ViewData("user", token.Claims)
		ctx.Next()
	})
	{
		admin.Get("/", routes.Index)
	}

	// app.Get("/", routes.Index)
	// app.Get("/admin", routes.Index)
	app.Get("/list", routes.ListTask)
	app.Get("/log", routes.RecentLog)
	app.Any("/update", routes.UpdateTask)
	app.Get("/stop", routes.StopTask)
	app.Get("/start", routes.StartTask)
	app.Any("/login", routes.Login)
	app.Get("/logout", routes.Logout)
	app.Get("/readme", routes.Readme)
	app.Get("/quickStart", routes.QuickStart)
	app.Get("/reloadConfig", routes.ReloadConfig)
	app.Get("/deleteClient", routes.DeleteClient)
	app.Get("/viewConfig", routes.ViewConfig)
	app.Get("/stopAllTask", routes.StopAllTask)
	app.Get("/model", routes.Model)

}
