package main

import (
	"reflect"
	"strings"

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

func router(app *jiaweb.JiaWeb) {
	var route = app.HttpServer.Route()
	route.ServerFile("/static/<key:.*>", file.GetCurrentDirectory()+"/static")
	route.GET("/", routes.Index)
	route.GET("/list", routes.ListTask)
	route.GET("/log", routes.RecentLog)
	route.GETPOST("/update", routes.UpdateTask)
	route.GET("/stop", routes.StopTask)
	route.GET("/start", routes.StartTask)
	route.GETPOST("/login", routes.Login)
	route.GET("/logout", routes.Logout)
	route.GET("/readme", routes.Readme)
	route.GET("/quickStart", routes.QuickStart)
	route.GET("/reloadConfig", routes.ReloadConfig)
	route.GET("/deleteClient", routes.DeleteClient)
	route.GET("/viewConfig", routes.ViewConfig)
	route.GET("/stopAllTask", routes.StopAllTask)
	route.GET("/model", routes.Model)
}
