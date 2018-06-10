package main

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kataras/iris"

	"jiacrontab/server/routes"

	"github.com/iwannay/jiaweb"
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

func registerController(i interface{}) jiaweb.HttpHandle {
	return func(ctx jiaweb.Context) error {
		action := strFirstUpper(ctx.QueryRouteParam("key"))
		t := reflect.TypeOf(i)
		if t.Kind() == reflect.Func {
			v := reflect.ValueOf(i)
			objSlice := v.Call([]reflect.Value{})
			if len(objSlice) >= 1 {
				v = objSlice[0]
				m := v.MethodByName(action)
				if m.IsValid() {
					errSli := m.Call([]reflect.Value{reflect.ValueOf(ctx)})
					if errSli[0].IsNil() {
						return nil
					}
					err := errSli[0].Interface().(error)
					return err
				}
			}

		}

		ctx.NotFound()

		return nil
	}
}

func router(app *iris.Application) {
	app.StaticWeb("/static", filepath.Join(file.GetCurrentDirectory(), "static"))
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
}
