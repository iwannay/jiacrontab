package handle

import (
	"jiacrontab/model"

	"github.com/kataras/iris"
)

func RuntimeInfo(ctx iris.Context) {
	var systemInfo map[string]interface{}
	var clientList []model.Client
	addr := ctx.FormValue("addr")
	if err := rpcCall(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		ctx.View("public/error.html", map[string]interface{}{
			"error": err,
		})
		return
	}
	model.DB().Model(&model.Client{}).Find(&clientList)

	ctx.ViewData("addr", addr)
	ctx.ViewData("clientList", clientList)
	ctx.ViewData("systemInfo", systemInfo)
	ctx.View("runtime/info.html")
}
