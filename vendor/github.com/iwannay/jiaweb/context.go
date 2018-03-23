package jiaweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/iwannay/jiaweb/base"
)

type (
	Context interface {
		HttpServer() *HttpServer
		Response() *Response
		Request() *Request
		Store() *base.Store
		RouteNode() RouteNode
		Handler() HttpHandle
		Hijack() (*HijackConn, error)
		RemoteIP() string
		IsHijack() bool
		End()
		NotFound()
		QueryRouteParam(key string) string
		RenderHtml(viewPath []string, locals map[string]interface{}) error

		WriteJSON(i interface{}) (int, error)
		WriteJSONAndStatus(status int, i interface{}) (int, error)
		WriteJSONBlob(b []byte) (int, error)
		WriteJSONBlobAndStatus(status int, b []byte) (int, error)
		WriteJSONP(callback string, i interface{}) (int, error)
		WriteJSONPBlob(callback string, b []byte) (size int, err error)

		WriteHtml(content ...interface{}) (int, error)
		WriteHtmlAndStatus(status int, content ...interface{}) (int, error)
		WriteString(content ...interface{}) (int, error)
		WriteStringAndStatus(status int, content ...interface{}) (int, error)
		WriteBlob(contentType string, b []byte) (int, error)
		WriteBlobAndStatus(status int, contentType string, b []byte) (int, error)

		GenerateToken(v jwt.MapClaims)
		CleanToken()
		GenerateSeesionToken(v jwt.MapClaims)
		StartTime() time.Time
		CostTime() string
		VerifyToken(v *map[string]interface{}) bool
		Redirect(url string, code int)
	}
	HttpContext struct {
		request     *Request
		response    *Response
		cancel      context.CancelFunc
		httpServer  *HttpServer
		handler     HttpHandle
		hiJackConn  *HijackConn
		routeNode   RouteNode
		isHijack    bool
		isWebsocket bool
		isEnd       bool
		store       *base.Store
		startTime   time.Time
		params      map[string]string
		mutex       sync.RWMutex
	}
)

func (ctx *HttpContext) reset(r *Request, rw *Response, httpServer *HttpServer) {
	ctx.request = r
	ctx.response = rw
	ctx.isHijack = false
	ctx.isWebsocket = false
	ctx.isEnd = false
	ctx.store = base.NewStore()
	ctx.httpServer = httpServer
	ctx.startTime = time.Now()
}

func (ctx *HttpContext) release() {
	ctx.request = nil
	ctx.response = nil
	ctx.routeNode = nil
	ctx.params = nil
	ctx.hiJackConn = nil
	ctx.isHijack = false
	ctx.httpServer = nil
	ctx.isEnd = false
	ctx.handler = nil
	ctx.store = nil
	ctx.startTime = time.Time{}
	ctx.mutex = sync.RWMutex{}
}

func (ctx *HttpContext) NotFound() {
	ctx.httpServer.JiaWeb.NotFoundHandler(ctx)
}

func (ctx *HttpContext) Store() *base.Store {
	return ctx.store
}

func (ctx *HttpContext) Redirect(url string, code int) {
	http.Redirect(ctx.Response().ResponseWriter(), ctx.Request().Request, url, code)
}

func (ctx *HttpContext) QueryRouteParam(key string) string {
	ctx.mutex.Lock()
	val, ok := ctx.params[key]
	ctx.mutex.Unlock()

	if ok {
		return val
	}
	return ""
}

func (ctx *HttpContext) CostTime() string {
	dur := time.Now().Sub(ctx.startTime)
	return fmt.Sprintf("%.3f ms", float64(dur.Nanoseconds())/float64(1000000))
}

func (ctx *HttpContext) StartTime() time.Time {
	return ctx.startTime
}

func (ctx *HttpContext) RenderHtml(viewPath []string, locals map[string]interface{}) error {
	return ctx.httpServer.Render.RenderHtml(ctx.response, viewPath, locals)
}

func (ctx *HttpContext) Request() *Request {
	return ctx.request
}

func (ctx *HttpContext) End() {
	ctx.isEnd = true
}

func (ctx *HttpContext) IsEnd() bool {
	return ctx.isEnd
}

func (ctx *HttpContext) Response() *Response {
	return ctx.response
}

func (ctx *HttpContext) RouteNode() RouteNode {
	return ctx.routeNode
}

func (ctx *HttpContext) HttpServer() *HttpServer {
	return ctx.httpServer
}

func (ctx *HttpContext) Handler() HttpHandle {
	return ctx.handler
}

func (ctx *HttpContext) WriteString(content ...interface{}) (int, error) {
	return ctx.WriteStringAndStatus(http.StatusOK, content...)
}

func (ctx *HttpContext) WriteStringAndStatus(status int, content ...interface{}) (int, error) {
	contents := fmt.Sprint(content...)
	return ctx.WriteBlobAndStatus(status, MIMETextPlainCharsetUTF8, []byte(contents))
}

func (ctx *HttpContext) WriteHtml(content ...interface{}) (int, error) {
	return ctx.WriteHtmlAndStatus(http.StatusOK, content...)
}

func (ctx *HttpContext) WriteHtmlAndStatus(status int, content ...interface{}) (int, error) {
	contents := fmt.Sprint(content...)
	return ctx.WriteBlobAndStatus(status, MIMETextHTMLCharsetUTF8, []byte(contents))
}

func (ctx *HttpContext) WriteBlob(contentType string, b []byte) (int, error) {
	return ctx.WriteBlobAndStatus(http.StatusOK, contentType, b)
}

func (ctx *HttpContext) WriteBlobAndStatus(code int, contentType string, b []byte) (int, error) {
	if contentType != "" {
		ctx.response.SetContentType(contentType)
	}

	if ctx.IsHijack() {
		return ctx.hiJackConn.WriteBlob(b)
	}
	return ctx.response.Write(code, b)

}

func (ctx *HttpContext) WriteJSON(i interface{}) (int, error) {
	return ctx.WriteJSONAndStatus(http.StatusOK, i)
}

func (ctx *HttpContext) WriteJSONAndStatus(status int, i interface{}) (int, error) {
	if b, err := json.Marshal(i); err != nil {
		return 0, err
	} else {
		return ctx.WriteJSONBlobAndStatus(http.StatusOK, b)
	}

}

func (ctx *HttpContext) WriteJSONBlob(b []byte) (int, error) {
	return ctx.WriteJSONBlobAndStatus(http.StatusOK, b)
}

func (ctx *HttpContext) WriteJSONBlobAndStatus(status int, b []byte) (int, error) {
	return ctx.WriteBlobAndStatus(status, MIMEApplicationJSONCharsetUTF8, b)
}

func (ctx *HttpContext) WriteJSONP(callback string, i interface{}) (int, error) {
	if callback != "" {
		b, err := json.Marshal(i)
		if err != nil {
			return 0, err
		}

		return ctx.WriteJSONPBlob(callback, b)
	} else {
		return ctx.WriteJSON(i)
	}

}

func (ctx *HttpContext) WriteJSONPBlob(callback string, b []byte) (size int, err error) {
	ctx.Response().SetContentType(MIMEApplicationJavaScriptCharsetUTF8)
	if ctx.IsHijack() {
		if size, err = ctx.hiJackConn.WriteBlob([]byte(ctx.hiJackConn.header + "\r\n")); err != nil {
			return
		}
	}
	if size, err = ctx.WriteBlob("", []byte(callback+"(")); err != nil {
		return
	}
	if size, err = ctx.WriteBlob("", b); err != nil {
		return
	}
	if size, err = ctx.WriteBlob("", b); err != nil {
		return
	}
	size, err = ctx.WriteBlob("", []byte(");"))
	return
}

func (ctx *HttpContext) IsHijack() bool {
	return ctx.isHijack
}

func (ctx *HttpContext) Hijack() (*HijackConn, error) {
	hj, ok := ctx.response.ResponseWriter().(http.Hijacker)
	if !ok {
		return nil, errors.New("web server does not support Hijackng!")
	}

	conn, buf, err := hj.Hijack()
	if err != nil {
		return nil, errors.New("Hijack error:" + err.Error())
	}
	ctx.hiJackConn = &HijackConn{
		Conn:       conn,
		ReadWriter: buf,
		header:     "HTTP/1.1 200 OK\r\n",
	}
	ctx.isHijack = true
	return ctx.hiJackConn, nil

}

func (ctx *HttpContext) RemoteIP() string {
	return ctx.Request().RemoteIP()
}

func (ctx *HttpContext) GenerateToken(v jwt.MapClaims) {
	ctx.HttpServer().Jwt.GenerateToken(ctx.Response().ResponseWriter(), v)
}

func (ctx *HttpContext) CleanToken() {
	ctx.HttpServer().Jwt.CleanCookie(ctx.Response().ResponseWriter())
}

func (ctx *HttpContext) GenerateSeesionToken(v jwt.MapClaims) {
	ctx.HttpServer().Jwt.GenerateSeesionToken(ctx.Response().ResponseWriter(), v)
}
func (ctx *HttpContext) VerifyToken(v *map[string]interface{}) bool {

	return ctx.HttpServer().Jwt.VerifyToken(ctx.Request().Request, v)
}
