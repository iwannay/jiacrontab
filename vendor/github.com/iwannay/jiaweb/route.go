package jiaweb

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iwannay/jiaweb/base"
	"github.com/iwannay/jiaweb/logger"
	"github.com/iwannay/jiaweb/utils"
)

const (
	HTTPMethod_Any       = "ANY"
	HTTPMethod_GET       = "GET"
	HTTPMethod_POST      = "POST"
	HTTPMethod_PUT       = "PUT"
	HTTPMethod_DELETE    = "DELETE"
	HTTPMethod_PATCH     = "PATCH"
	HTTPMethod_HiJack    = "HIJACK"
	HTTPMethod_WebSocket = "WEBSOCKET"
	HTTPMethod_HEAD      = "HEAD"
	HTTPMethod_OPTIONS   = "OPTIONS"
)

type (
	// Middleware func(httpCtx *HttpContext)
	Router interface {
		ServeHTTP(ctx *HttpContext)
		ServerFile(path, fileRoot string) RouteNode
		RegisterRoute(method string, path string, handle HttpHandle) RouteNode
		HEAD(path string, handle HttpHandle) RouteNode
		POST(path string, handle HttpHandle) RouteNode
		GETPOST(path string, handle HttpHandle)
		Any(path string, handle HttpHandle)
		GET(path string, handle HttpHandle) RouteNode
		PUT(path string, handle HttpHandle) RouteNode
		DELETE(path string, handle HttpHandle) RouteNode
		PATCH(path string, handle HttpHandle) RouteNode
		OPTIONS(path string, handle HttpHandle) RouteNode
	}

	RouteNode interface {
		Use(m ...Middleware) *Node
		Middlewares() []Middleware
		Node() *Node
	}

	route struct {
		handleMap             map[string]HttpHandle
		NodeMap               map[string]*Node
		rwMutex               sync.RWMutex
		RedirectTrailingSlash bool
		server                *HttpServer
		RedirectFixedPath     bool
		HandleOPTIONS         bool
	}

	RouteHandle func(ctx *HttpContext)
)

var (
	SupportHTTPMethod map[string]bool
)

func NewRoute(server *HttpServer) *route {
	return &route{
		handleMap:             make(map[string]HttpHandle),
		NodeMap:               make(map[string]*Node),
		RedirectTrailingSlash: true,
		RedirectFixedPath:     true,
		HandleOPTIONS:         true,
		server:                server,
	}
}

func (r *route) RegisterHandler(name string, handler HttpHandle) {
	r.rwMutex.Lock()
	r.handleMap[name] = handler
	r.rwMutex.Unlock()
}

func (r *route) GetHandler(name string) (HttpHandle, bool) {
	r.rwMutex.RLock()
	h, ok := r.handleMap[name]
	r.rwMutex.RUnlock()
	return h, ok
}

func (r *route) ServeHTTP(ctx *HttpContext) {

	req := ctx.Request().Request
	rw := ctx.Response().ResponseWriter()
	path := req.URL.Path
	if root := r.NodeMap[req.Method]; root != nil {
		if node, handler, params := root.GetValue(path); handler != nil {
			ctx.params = params
			ctx.routeNode = node
			handler(ctx)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301
			if req.Method != "GET" {
				code = 307
			}

			if r.RedirectTrailingSlash && filepath.Ext(path) == "" {

				if len(path) > 1 && path[len(path)-1] != '/' {
					req.URL.Path = path + "/"
					http.Redirect(rw, req, req.URL.String(), code)
					return
				}

			}

			if r.RedirectFixedPath {
				// TODO 自动补全斜线
			}

		}

	}

	if req.Method == "OPTIONS" {
		if r.HandleOPTIONS {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				rw.Header().Set("Allow", allow)
				return
			}
			if r.RedirectTrailingSlash && filepath.Ext(path) == "" {
				if allow := r.allowed(path+"/", req.Method); len(allow) > 0 {
					rw.Header().Set("Allow", allow)
					return
				}
			}

		}
	} else {
		// 405
		if allow := r.allowed(path, req.Method); len(allow) > 0 {

			ctx.Response().SetHeader("Allow", allow)
			ctx.Response().SetStatusCode(http.StatusMethodNotAllowed)

			r.server.JiaWeb.MethodNotAllowedHandler(ctx)
			return

		}
	}

	// Handle 404
	ctx.Response().WriteHeader(http.StatusNotFound)
	r.server.JiaWeb.NotFoundHandler(ctx)

}

func (r *route) allowed(path, reqMethod string) (allow string) {
	if path == "*" {
		for method := range r.NodeMap {
			if method == "OPTIONS" {
				continue
			}

			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else {
		for method := range r.NodeMap {
			if method == reqMethod || method == "OPTIONS" {
				continue
			}
			_, h, _ := r.NodeMap[method].GetValue(path)
			if h != nil {
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}

	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

func (r *route) wrapRouteHandle(handler HttpHandle, isHijack bool) RouteHandle {
	return func(ctx *HttpContext) {
		ctx.handler = handler

		// TODO do feature

		if isHijack {
			// TODO Hijack
			_, err := ctx.Hijack()
			if err != nil {
				ctx.Response().WriteHeader(http.StatusInternalServerError)
				ctx.Response().Header().Set(HeaderContentType, CharsetUTF8)

			}
		}

		defer func() {
			if err := recover(); err != nil {
				errMsg := base.FormatError("HttpServer::RouterHandle", LogTarget_HttpServer, err)
				if r.server.JiaWeb.ExceptionHandler != nil {
					r.server.JiaWeb.ExceptionHandler(ctx, fmt.Errorf("%v", err))
				}

				if logger.EnableLog {
					headinfo := fmt.Sprintln(ctx.Response().Header())
					logJson := LogJson{
						RequestUrl: ctx.Request().RequestURI,
						HttpHeader: headinfo,
						HttpBody:   errMsg,
					}
					logString := utils.GetJsonString(logJson)
					logger.Logger().Error(logString, LogTarget_HttpServer)
				}

				base.GlobalState.AddErrorCount(ctx.Request().Path(), fmt.Errorf("%v", err), 1)
			}

			// TODO Release FeatureTool
			if ctx.cancel != nil {
				ctx.cancel()
			}

		}()

		// do user handle
		var err error
		if len(r.server.JiaWeb.Middlewares) > 0 {
			err = r.server.JiaWeb.Middlewares[0].Handle(ctx)
		} else {
			err = handler(ctx)
		}

		if err != nil {
			if r.server.JiaWeb.ExceptionHandler != nil {
				r.server.JiaWeb.ExceptionHandler(ctx, err)
				base.GlobalState.AddErrorCount(ctx.Request().Path(), err, 1)
			}

		}

	}
}

func (r *route) ServerFile(path string, fileroot string) RouteNode {
	var root http.FileSystem
	root = http.Dir(fileroot)
	if !r.server.ServerConfig().EnableListDir {
		root = &base.HideDirFS{root}
	}
	fileServer := http.FileServer(root)
	// fileServer = http.StripPrefix("/dist/", fileServer)
	node := r.add(HTTPMethod_GET, path, r.wrapFileHandle(fileServer))
	return node
}

func (r *route) add(method, path string, handle RouteHandle, m ...Middleware) *Node {
	if path == "" || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.NodeMap == nil {
		r.NodeMap = make(map[string]*Node)
	}

	root := r.NodeMap[method]
	if root == nil {
		root = NewTree()
		r.NodeMap[method] = root
	}

	root.insertChild(path, handle)

	return root
}

func (r *route) wrapFileHandle(fHandler http.Handler) RouteHandle {
	return func(httpCtx *HttpContext) {
		startTime := time.Now()
		// TODO not supprot read dir by params
		fHandler.ServeHTTP(httpCtx.Response().rw, httpCtx.Request().Request)
		timeTaken := int64(time.Now().Sub(startTime) / time.Millisecond)
		logger.Logger().Debug(httpCtx.Request().Url()+" "+logRequest(httpCtx, timeTaken), LogTarget_HttpRequest)
	}
}

func handlerName(h HttpHandle) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

func (r *route) Any(path string, handle HttpHandle) {
	r.RegisterRoute(HTTPMethod_DELETE, path, handle)
	r.RegisterRoute(HTTPMethod_GET, path, handle)
	r.RegisterRoute(HTTPMethod_HEAD, path, handle)
	r.RegisterRoute(HTTPMethod_OPTIONS, path, handle)
	r.RegisterRoute(HTTPMethod_PUT, path, handle)
	r.RegisterRoute(HTTPMethod_POST, path, handle)
	r.RegisterRoute(HTTPMethod_PATCH, path, handle)
}

func (r *route) HEAD(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_HEAD, path, handle)
}

func (r *route) OPTIONS(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_OPTIONS, path, handle)
}

func (r *route) POST(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_POST, path, handle)
}

func (r *route) GETPOST(path string, handle HttpHandle) {
	r.RegisterRoute(HTTPMethod_POST, path, handle)
	r.RegisterRoute(HTTPMethod_GET, path, handle)
}

func (r *route) PUT(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_PUT, path, handle)
}

func (r *route) PATCH(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_PATCH, path, handle)
}

func (r *route) DELETE(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_DELETE, path, handle)
}

func (r *route) GET(path string, handle HttpHandle) RouteNode {
	return r.RegisterRoute(HTTPMethod_GET, path, handle)
}

func (r *route) RegisterRoute(routeMethod string, path string, handle HttpHandle) RouteNode {
	var node *Node
	routeMethod = strings.ToUpper(routeMethod)
	if _, ok := SupportHTTPMethod[routeMethod]; !ok {
		logger.Logger().Warn("JiaWeb:Router:Registe failed illegal method "+routeMethod+"["+path+"]", LogTarget_HttpServer)

		return nil
	}
	logger.Logger().Debug("JiaWbe:Router:RegisterRoute Success "+routeMethod+"["+path+"]", LogTarget_HttpServer)

	// TODO websocket

	if routeMethod == HTTPMethod_HiJack {
		r.add(HTTPMethod_GET, path, r.wrapRouteHandle(handle, true))
	} else {
		node = r.add(routeMethod, path, r.wrapRouteHandle(handle, false))
	}

	if r.server.ServerConfig().EnableAutoHEAD {
		if routeMethod == HTTPMethod_HiJack {
			r.add(HTTPMethod_HEAD, path, r.wrapRouteHandle(handle, true))
		} else if routeMethod != HTTPMethod_Any {
			r.add(HTTPMethod_HEAD, path, r.wrapRouteHandle(handle, false))
		}
	}
	return node
}

func logRequest(ctx Context, timeTaken int64) string {
	var reqByteLen, resByteLen, method, proto, status, userip string
	reqByteLen = utils.Int642String(ctx.Request().ContentLength)
	resByteLen = ""
	method = ctx.Request().Method
	proto = ctx.Request().Proto
	status = "200"
	userip = ctx.Request().RemoteIP()

	return fmt.Sprintf("%s %s %s %s %s %s %s",
		method,
		userip,
		proto,
		status,
		reqByteLen,
		resByteLen,
		utils.Int642String(timeTaken))
}

func init() {
	SupportHTTPMethod = make(map[string]bool)
	SupportHTTPMethod[HTTPMethod_Any] = true
	SupportHTTPMethod[HTTPMethod_GET] = true
	SupportHTTPMethod[HTTPMethod_POST] = true
	SupportHTTPMethod[HTTPMethod_PUT] = true
	SupportHTTPMethod[HTTPMethod_DELETE] = true
	SupportHTTPMethod[HTTPMethod_PATCH] = true
	SupportHTTPMethod[HTTPMethod_HiJack] = true
	SupportHTTPMethod[HTTPMethod_WebSocket] = true
	SupportHTTPMethod[HTTPMethod_HEAD] = true
	SupportHTTPMethod[HTTPMethod_OPTIONS] = true
}
