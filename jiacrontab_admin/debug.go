package admin

import (
	"jiacrontab/pkg/base"
	"net/http/pprof"

	"github.com/kataras/iris"
)

func stat(ctx iris.Context) {
	data := base.Stat.Collect()
	ctx.JSON(data)
}

func pprofHandler(ctx iris.Context) {
	if h := pprof.Handler(ctx.Params().Get("key")); h != nil {
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
	}
}

func indexDebug(ctx iris.Context) {
	pprof.Index(ctx.ResponseWriter(), ctx.Request())
}
