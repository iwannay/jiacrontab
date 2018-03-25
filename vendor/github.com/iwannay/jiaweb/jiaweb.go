package jiaweb

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"

	"github.com/iwannay/jiaweb/config"

	"github.com/iwannay/jiaweb/base"
	"github.com/iwannay/jiaweb/logger"
	"github.com/iwannay/jiaweb/utils"
)

type (
	JiaWeb struct {
		HttpServer              *HttpServer
		Config                  *config.Config
		Logger                  logger.JiaLogger
		Middlewares             []Middleware
		ExceptionHandler        ExceptionHandle
		NotFoundHandler         StandardHandle
		MethodNotAllowedHandler StandardHandle

		mutex sync.RWMutex
	}

	// 自定义异常处理
	ExceptionHandle func(Context, error)

	StandardHandle func(Context)
	HttpHandle     func(httpCtx Context) error
)

var appConfig *config.Config

const (
	DefaultHTTPPort    = 8080
	RunModeDevelopment = "development"
	RunModeProduction  = "production"
)

func LoadConfig(configFile, configType string) {
	var err error
	appConfig, err = config.InitConfig(configFile, configType)
	if err != nil {
		appConfig = config.New()
	}

}

func New(fn func(app *JiaWeb)) *JiaWeb {

	if appConfig == nil {
		appConfig = config.New()
	}

	app := &JiaWeb{
		HttpServer:  NewHttpServer(),
		Config:      appConfig,
		Middlewares: make([]Middleware, 0),
	}
	app.HttpServer.SetJiaWeb(app)
	app.init(fn)
	return app
}

func Classic(fn func(app *JiaWeb)) *JiaWeb {
	app := New(fn)
	app.UseRequestLog()
	logger.Logger().Debug("JiaWeb start New AppServer", LogTarget_HttpServer)

	return app

}

func (app *JiaWeb) SetEnableLog(enableLog bool) {
	logger.SetEnableLog(enableLog)

}

func (app *JiaWeb) SetLogPath(path string) {
	logger.SetLogPath(path)
}

func (app *JiaWeb) Use(m ...Middleware) {
	step := len(app.Middlewares) - 1
	for i := range m {
		if m[i] != nil {
			if step >= 0 {
				app.Middlewares[step].SetNext(m[i])
			}
			app.Middlewares = append(app.Middlewares, m[i])
			step++
		}
	}
}

func (app *JiaWeb) SetExceptionHandle(handler ExceptionHandle) {
	app.ExceptionHandler = handler
}

func (app *JiaWeb) SetNotFoundHandle(handler StandardHandle) {
	app.NotFoundHandler = handler
}

func (app *JiaWeb) SetMethodNotAllowedHandle(handler StandardHandle) {
	app.MethodNotAllowedHandler = handler
}

// func (app *JiaWeb) RegisterMiddlewareFunc(name string, middleFunc MiddlewareFunc) {
// 	app.mutex.Lock()
// 	app.mutex[name] = middleFunc
// 	app.mutex.Unlock()
// }

func (app *JiaWeb) SetPProfConfig(enablePProf bool, port int) {
	app.Config.App.EnablePProf = enablePProf
	app.Config.App.PProfPort = port
	logger.Logger().Debug("JiaWeb SetPProfConfig ["+strconv.FormatBool(enablePProf)+", "+strconv.Itoa(port)+"]", LogTarget_HttpServer)
}

func (app *JiaWeb) UseRequestLog() {
	app.Use(&RequestLogMiddleware{})
}

func (app *JiaWeb) SetEnableDetailRequestData(enable bool) {
	base.GlobalState.EnableDetailRequestData = enable
}

func (app *JiaWeb) StartServer(port int) error {
	addr := ":" + strconv.Itoa(port)
	return app.ListenAndServe(addr)
}

func (app *JiaWeb) initAppConfig() {
	config := app.Config
	if config.App.LogPath != "" {
		logger.SetLogPath(config.App.LogPath)
	}
	logger.SetEnableLog(config.App.EnableLog)

	if app.Config.App.RunMode != RunModeProduction {
		app.Config.App.RunMode = RunModeDevelopment
	} else {
		app.Config.App.RunMode = RunModeProduction
	}

	// TODO CROS Config

	// TODO 设置维护状态

	// TODO set session

	if config.Server.EnableDetailRequestData {
		base.GlobalState.EnableDetailRequestData = true
	}
}

func (app *JiaWeb) initRegisterMiddleware() {

}

func (app *JiaWeb) initRegisterRoute() {

}

func (app *JiaWeb) initRegisterGroup() {

}

func (app *JiaWeb) initInnnerRouter() {
	inner := app.HttpServer.Group("/jiaweb")
	inner.GET("/debug/pprof/<key:.+>", initPProf)
	inner.GET("/debug/freememory", freeMemory)
	inner.GET("/debug/state", showServerState)
	inner.GET("/debug/query/<key:[^/]*>", showQuery)
}

func (app *JiaWeb) init(fn func(app *JiaWeb)) {
	app.Logger = logger.Logger()
	app.initAppConfig()
	if fn != nil {
		fn(app)
	}
	printLogo()
	app.initRegisterMiddleware()
	app.initRegisterRoute()
	app.initRegisterGroup()
	app.initServerEnvironment()
	app.initInnnerRouter()
}

func (app *JiaWeb) ListenAndServe(addr string) error {
	app.Use(&xMiddleware{})
	err := app.HttpServer.ListenAndServe(addr)
	return err

}

// func (app *JiaWeb) RegisterMiddlewareFunc(name string,)

func (app *JiaWeb) initServerEnvironment() {
	if app.ExceptionHandler == nil {
		app.SetExceptionHandle(app.DefaultHTTPErrorHandler)
	}
	if app.NotFoundHandler == nil {
		app.SetNotFoundHandle(app.DefaultNotFoundHandler)
	}

	if app.MethodNotAllowedHandler == nil {
		app.SetMethodNotAllowedHandle(app.DefaultMethodNotAllowedHandler)
	}

	// TODO init session

	// TODO init cache

	// TODO init render

	if app.Config.App.EnablePProf {
		logger.Logger().Debug("JiaWeb:StartPProfServer["+strconv.Itoa(app.Config.App.PProfPort)+"] Begin", LogTarget_HttpServer)

		go func() {
			err := http.ListenAndServe(":"+strconv.Itoa(app.Config.App.PProfPort), nil)
			if err != nil {
				logger.Logger().Error("JiaWbe:StartPProfServer["+strconv.Itoa(app.Config.App.PProfPort)+"] errror:"+err.Error(), LogTarget_HttpServer)
				panic(err)
			}
		}()
	}
}

func (app *JiaWeb) RunMode() string {
	if app.Config.App.RunMode != RunModeDevelopment && app.Config.App.RunMode != RunModeProduction {
		app.Config.App.RunMode = RunModeDevelopment
	}
	return app.Config.App.RunMode
}

func (app *JiaWeb) IsDevelopmentMode() bool {
	return app.RunMode() == RunModeDevelopment
}

func (app *JiaWeb) SetDevelopmentMode() {
	app.Config.App.RunMode = RunModeDevelopment
	app.SetEnableLog(true)
	logger.SetEnableConsole(true)
}

func (app *JiaWeb) SetProductionMode() {
	app.Config.App.RunMode = RunModeProduction
	app.SetEnableLog(true)
	logger.SetEnableConsole(false)
}

func (app *JiaWeb) DefaultHTTPErrorHandler(ctx Context, err error) {
	ctx.Response().Header().Set(HeaderContentType, CharsetUTF8)
	if app.IsDevelopmentMode() {
		stack := string(debug.Stack())
		ctx.WriteStringAndStatus(http.StatusInternalServerError, fmt.Sprintln(err)+stack)
	} else {
		ctx.WriteStringAndStatus(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (app *JiaWeb) DefaultNotFoundHandler(ctx Context) {
	ctx.Response().Header().Set(HeaderContentType, CharsetUTF8)
	ctx.WriteStringAndStatus(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (app *JiaWeb) DefaultMethodNotAllowedHandler(ctx Context) {
	ctx.Response().Header().Set(HeaderContentType, CharsetUTF8)
	ctx.WriteStringAndStatus(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}

func (app *JiaWeb) RuntimeState() base.RuntimeInfo {
	return base.GlobalState.GetRuntimeInfo()
}

func initPProf(ctx Context) error {
	querykey := ctx.QueryRouteParam("key")
	runtime.GC()
	p := pprof.Lookup(querykey)
	if p != nil {
		p.WriteTo(ctx.Response().ResponseWriter(), 1)
	} else {
		ctx.WriteString("pprof can not found key=>" + querykey)
	}

	return nil
}

func freeMemory(ctx Context) error {
	debug.FreeOSMemory()
	return nil
}

func showIntervalData(ctx Context) error {
	type data struct {
		Time         string
		RequestCount uint64
		ErrorCount   uint64
	}
	queryKey := ctx.QueryRouteParam("key")
	d := new(data)
	d.Time = queryKey
	d.RequestCount = base.GlobalState.QueryIntervalRequstData(queryKey)
	d.ErrorCount = base.GlobalState.QueryIntervalErrorData(queryKey)
	ctx.WriteJSON(d)
	return nil
}

func showServerState(ctx Context) error {
	dataType := ctx.Request().FormValue("type")
	ctx.WriteHtml(base.GlobalState.ShowData(dataType))
	return nil
}

func showQuery(ctx Context) error {
	queryKey := ctx.QueryRouteParam("key")
	switch queryKey {
	case "state":
		ctx.WriteString(utils.GetJsonString(base.GlobalState))

	case "":
		ctx.WriteString("please input key")
	default:
		ctx.WriteString("not support key =>" + queryKey)
	}
	return nil
}

func printLogo() {

	str := `                                                         
         ;;.  .:.            ;;.     ;;.    .;;            !OC        
        ;MM: .NM7            NMC    >MM!    QMC            CM&        
        !MM. ;NN-            HMC    HMM!   :MM;            HM?        
        ?MH                  QMO   >MNM>   &MC            .MM!        
        &MO  :7!   ;>?C7-    &MO   HMOM7  -MM;   .>?C>.   :MM;-?C7;   
        NM>  QMC  :MMNMMM!   OM&  !MCCM?  OMC   >MMNMMH.  7MNOMMMMN-  
       ;MM-  MM>  ;:. .&MQ   CMQ  QM;CM? ;MM;  ?MQ; .HM7  OMM&-.!MM&  
       !MM. -MM-       ?MH   ?MQ !MC ?MC CMC  -MN.   ?MO  HM&    OMN  
       ?MQ  >MM.  .7&NMMM&   7MQ QM; 7MO;MM;  OMN&&&&NMC .MM!    ?MN  
       QMC  CM&  !NMO7!NM7   >MH:MO  >MO?MC   NMQ&&&&&&: :MM;    OM&  
      .MM!  QM? ;MM!  .MM:   !MH&M;  >MONM;  .MM:        7MN    .MM>  
      CMN. .MM! >MN   >MM.   :MNMO   !MNM?    NM?        OMM;  .&MH.  
  !CCQMM:  -MM; >MM?>CMMH    -MMM-   :MMM;    7MM&7>?O   NMNN??HMN;   
  QMMM&-   >MN  .OMMMO7M&    ;MMO    -MM?      7NMMMM&  ;MH-HMMMC.   
`

	LogMsg := strings.Split(str, "\n")

	for _, v := range LogMsg {
		logger.Logger().Print(v, LogTarget_HttpServer)
	}
}

func init() {
	logger.InitJiaLog()
}
