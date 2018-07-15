package handle

import (
	"jiacrontab/libs/base"
	"net/http/pprof"
	"runtime/debug"

	"github.com/kataras/iris"
)

func Stat(ctx iris.Context) {
	data := base.Stat.Collect()
	ctx.JSON(data)
}

func PprofHandler(ctx iris.Context) {

	if h := pprof.Handler(ctx.Params().Get("key")); h != nil {

		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
	}
}

func IndexDebug(ctx iris.Context) {
	pprof.Index(ctx.ResponseWriter(), ctx.Request())
}

func FreeMem(ctx iris.Context) {
	debug.FreeOSMemory()
}
