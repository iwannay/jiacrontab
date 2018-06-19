package handle

import "github.com/kataras/iris"

func RuntimeInfo(ctx iris.Context) {
	ctx.View("runtime/info.html")
}
