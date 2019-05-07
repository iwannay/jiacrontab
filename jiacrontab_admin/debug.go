package admin

import (
	"jiacrontab/pkg/base"
	"net/http/pprof"
)

func stat(ctx *myctx) {
	data := base.Stat.Collect()
	ctx.JSON(data)
}

func pprofHandler(ctx *myctx) {
	if h := pprof.Handler(ctx.Params().Get("key")); h != nil {
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
	}
}

func indexDebug(ctx *myctx) {
	pprof.Index(ctx.ResponseWriter(), ctx.Request())
}
