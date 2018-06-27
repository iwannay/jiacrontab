package handle

import (
	"jiacrontab/libs/rpc"

	"github.com/kataras/iris"
)

func RuntimeInfo(ctx iris.Context) {
	var systemInfo map[string]interface{}
	addr := ctx.FormValue("addr")
	if err := rpc.Call(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		ctx.View("public/error.html", map[string]interface{}{
			"error": err,
		})
		return
	}

	ctx.ViewData("systemInfo", systemInfo)
	ctx.View("runtime/info.html")
}
