package jiaweb

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/iwannay/jiaweb/logger"
	"github.com/iwannay/jiaweb/utils"
)

type (
	MiddlewareFunc func() Middleware
	Middleware     interface {
		Handle(ctx Context) error
		SetNext(m Middleware)
		Next(ctx Context) error
	}

	BaseMiddleware struct {
		next Middleware
	}
)

func (bm *BaseMiddleware) SetNext(m Middleware) {
	bm.next = m
}

func (bm *BaseMiddleware) Next(ctx Context) error {
	return bm.next.Handle(ctx)
}

type xMiddleware struct {
	BaseMiddleware
	IsEnd bool
}

func (x *xMiddleware) Handle(ctx Context) error {

	n := ctx.RouteNode()
	len := len(n.Middlewares())
	if x.IsEnd {
		return ctx.Handler()(ctx)
	}
	if x.next == nil {
		if len == 0 {
			return ctx.Handler()(ctx)
		}
		if reflect.TypeOf(ctx.RouteNode().Middlewares()[len-1]).String() != "*jiaweb.xMiddleware" {
			ctx.RouteNode().Use(&xMiddleware{IsEnd: true})
		}
		return ctx.RouteNode().Middlewares()[0].Handle(ctx)

	}
	return x.Next(ctx)

}

type RequestLogMiddleware struct {
	BaseMiddleware
}

func (m *RequestLogMiddleware) Handle(ctx Context) error {
	m.Next(ctx)

	timeTaken := int64(time.Now().Sub(ctx.(*HttpContext).startTime) / time.Millisecond)
	log := fmt.Sprintf("%s %s", ctx.Request().Url(), httpLog(ctx, timeTaken))
	logger.Logger().Debug(log, LogTarget_HttpRequest)
	return nil
}

func httpLog(ctx Context, timeTaken int64) string {
	var reqByteLen, resByteLen, method, proto, status, userip string
	if ctx != nil {
		reqByteLen = utils.Int642String(ctx.Request().ContentLength)
		resByteLen = utils.Int642String(ctx.Response().Size)
		method = ctx.Request().Method
		proto = ctx.Request().Proto
		status = strconv.Itoa(ctx.Response().Status)
		userip = ctx.RemoteIP()
	}
	return fmt.Sprintf(
		"%s %s %s %s %s %s %d",
		method,
		userip,
		proto,
		status,
		reqByteLen,
		resByteLen,
		timeTaken,
	)
}
