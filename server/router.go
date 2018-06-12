package main

import (
	"path/filepath"

	"strings"

	"github.com/kataras/iris"

	"jiacrontab/server/routes"

	"fmt"
	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	"github.com/iwannay/jiaweb/utils/file"
	"runtime"
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
	fmt.Println("hello boy")
	user := ctx.Values().Get("jwt").(*jwt.Token)
	fmt.Sprintln(user.SignedString(map[string]interface{}{
		"hello": "boyd",
	}))

	ctx.Writef("This is an authenticated request\n")
	ctx.Writef("Claim content:\n")

	ctx.Writef("%s", user.Signature)
}

func router(app *iris.Application) {
	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return config.JWTSigningKey, nil
		},

		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
	})

	app.Use(jwtHandler.Serve)

	app.StaticWeb("/static", filepath.Join(file.GetCurrentDirectory(), "static"))
	app.Use(func(ctx iris.Context) {
		ctx.ViewData("action", ctx.Request().URL.Path)
		ctx.ViewData("title", "jiacrontab")
		ctx.ViewData("goVersion", runtime.Version())
		ctx.ViewData("appVersion", "v1.3.5")
		ctx.ViewData("requestPath", ctx.Request().URL.Path)
		ctx.ViewData("staticDir", "static")
		ctx.Next()
	})
	app.Get("/", routes.Index)
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

	admin := app.Party("/admin")
	{
		admin.Get("/profile", jwtHandler.Serve)
	}

}
