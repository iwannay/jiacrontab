package jiaweb

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/iwannay/jiaweb/base"

	// "github.com/iwannay/view"

	"github.com/iwannay/jiaweb/config"
	"github.com/iwannay/jiaweb/logger"
)

type (
	HttpServer struct {
		stdServer *http.Server
		pool      *pool
		route     Router
		JiaWeb    *JiaWeb
		Jwt       *base.MJwt
		Modules   []*HttpModule
		Render    Viewer
		// modelView *view.ModelView
		end bool
	}

	pool struct {
		request  sync.Pool
		response sync.Pool
		context  sync.Pool
	}
	LogJson struct {
		RequestUrl string
		HttpHeader string
		HttpBody   string
	}
)

func NewHttpServer() *HttpServer {
	s := &HttpServer{
		end: false,
		pool: &pool{

			context: sync.Pool{
				New: func() interface{} {
					return &HttpContext{}
				},
			},

			request: sync.Pool{
				New: func() interface{} {
					return &Request{}
				},
			},

			response: sync.Pool{
				New: func() interface{} {
					return &Response{}
				},
			},
		},
	}
	s.stdServer = &http.Server{
		Handler: s,
	}
	s.route = NewRoute(s)
	s.Render = NewView(s)

	return s
}

func (s *HttpServer) ListenAndServe(addr string) error {
	s.stdServer.Addr = addr
	logger.Logger().Debug("jiaWeb:HttpServer ListenAndServe ["+addr+"]", LogTarget_HttpServer)
	return s.stdServer.ListenAndServe()
}

func (s *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set(HeaderServer, DefaultServerName)
	httpctx := s.pool.context.Get().(*HttpContext)
	request := s.pool.request.Get().(*Request)
	response := s.pool.response.Get().(*Response)

	httpctx.reset(request, response, s)
	request.reset(req, httpctx)
	response.reset(rw)

	for _, module := range s.Modules {
		if module.OnBeginRequest != nil {
			module.OnBeginRequest(httpctx)
		}
	}

	if !httpctx.IsEnd() {
		s.Route().ServeHTTP(httpctx)
	}

	for _, module := range s.Modules {
		if module.OnEndRequest != nil {
			module.OnEndRequest(httpctx)
		}
	}

	base.GlobalState.AddRequestCount(httpctx.Request().Path(), httpctx.Response().HttpStatus(), 1)
	response.release()
	s.pool.response.Put(response)
	request.release()
	s.pool.request.Put(request)

	httpctx.release()
	s.pool.context.Put(httpctx)
}

func (s *HttpServer) SetEnableJwt() {
	if jwtConf := s.JwtConfig(); jwtConf != nil {
		jwtConf.EnableJwt = true
		logger.Logger().Debug(
			fmt.Sprintf(
				"JiaWeb:HttpServer SetJwtConfig expire:%d name:%s cookieMaxAge:%d",
				jwtConf.Expire, jwtConf.Name, jwtConf.CookieMaxAge), LogTarget_HttpServer)
		s.Jwt = base.NewJwt(jwtConf.Domain, jwtConf.Expire, jwtConf.Name, []byte(jwtConf.SignKey), jwtConf.CookieMaxAge)
		return
	}

	logger.Logger().Debug("JiaWeb:HttpServer SetJwtConfig failed config nil", LogTarget_HttpServer)
}

func (s *HttpServer) JwtConfig() *config.JwtNode {
	return s.JiaWeb.Config.Jwt
}

func (s *HttpServer) SetEnableIgnoreFavicon(enable bool) {
	s.ServerConfig().EnableIgnoreFavicon = enable
	logger.Logger().Debug("JiaWeb:HttpServer ignore favicon", LogTarget_HttpServer)
	s.RegisterModule(ignoreFaviconModule())
}

func (s *HttpServer) RegisterModule(module *HttpModule) {
	s.Modules = append(s.Modules, module)
	logger.Logger().Debug(fmt.Sprintf("JiaWeb:HttpServer RegisterModule [%s]", module.Name), LogTarget_HttpServer)
}

func (s *HttpServer) SetEnableDetailRequestData(enable bool) {
	s.ServerConfig().EnableDetailRequestData = enable
	logger.Logger().Debug(fmt.Sprintf("JiaWeb:HttpServer SetEnableDetailRequest [%b]", enable), LogTarget_HttpRequest)
}

func (s *HttpServer) SetJiaWeb(jiaweb *JiaWeb) {
	s.JiaWeb = jiaweb
}

func (s *HttpServer) Route() Router {
	return s.route
}

func (s *HttpServer) ServerConfig() *config.ServerNode {
	return s.JiaWeb.Config.Server
}

func (s *HttpServer) TemplateConfig() *config.TemplateNode {
	return s.JiaWeb.Config.Template
}

func (s *HttpServer) Group(prefix string) Group {
	return NewGroup(prefix, s)
}
