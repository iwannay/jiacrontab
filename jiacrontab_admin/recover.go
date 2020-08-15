package admin

import (
	"fmt"
	"jiacrontab/pkg/base"
	"net/http"
	"runtime"
	"strconv"

	"github.com/kataras/iris/v12"
)

func getRequestLogs(ctx *myctx) string {
	var status, ip, method, path string
	status = strconv.Itoa(ctx.GetStatusCode())
	path = ctx.Path()
	method = ctx.Method()
	ip = ctx.RemoteAddr()
	// the date should be logged by iris' Logger, so we skip them
	return fmt.Sprintf("%v %s %s %s", status, path, method, ip)
}

func newRecover(adm *Admin) iris.Handler {
	return func(c iris.Context) {
		ctx := wrapCtx(c, adm)
		base.Stat.AddConcurrentCount()
		defer func() {
			if err := recover(); err != nil {

				base.Stat.AddErrorCount(ctx.RequestPath(true), fmt.Errorf("%v", err), 1)

				if ctx.IsStopped() {
					return
				}

				var stacktrace string
				for i := 1; ; i++ {
					_, f, l, got := runtime.Caller(i)
					if !got {
						break

					}

					stacktrace += fmt.Sprintf("%s:%d\n", f, l)
				}

				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
				logMessage += fmt.Sprintf("At Request: %s\n", getRequestLogs(ctx))
				logMessage += fmt.Sprintf("Trace: %s\n", err)
				logMessage += fmt.Sprintf("\n%s", stacktrace)
				ctx.Application().Logger().Warn(logMessage)

				ctx.StatusCode(500)
				ctx.respError(http.StatusInternalServerError, fmt.Sprint(err), nil)
				ctx.StopExecution()
			}
		}()
		base.Stat.AddRequestCount(ctx.RequestPath(true), ctx.GetStatusCode(), 1)
		c.Next()
	}
}
