package main

// type beforeReqHandle func(rw http.ResponseWriter, r *http.Request, m *modelView) bool

type ResponseData struct {
	Code int
	Msg  string
	Data interface{}
}

// func wrapHandler(fn func(rw http.ResponseWriter, r *http.Request, m *modelView), b []beforeReqHandle) http.HandlerFunc {
// 	return func(rw http.ResponseWriter, r *http.Request) {
// 		startNa := float64(time.Now().UnixNano())
// 		startMs := startNa / 1000000

// 		modelV := newModelView(rw, globalStore)
// 		modelV.startTime = startNa

// 		defer func() {
// 			k := fmt.Sprintf("%s|%s", getHttpClientIp(r), r.Header.Get("User-Agent"))
// 			endMs := float64(time.Now().UnixNano()) / 1000000
// 			log.Printf("%s %s %s %s %fms", k, r.Method, r.URL.Path, r.URL.RawQuery, endMs-startMs)
// 			if e, ok := recover().(error); ok {
// 				if globalConfig.debug == true {
// 					_, f, l, ok := runtime.Caller(0)

// 					errStr := fmt.Sprintf("%s %d %t %s\n%s", f, l, ok, e.Error(), string(debug.Stack()))
// 					// http.Error(rw, errStr, http.StatusInternalServerError)
// 					modelV.renderHtml2([]string{"public/error"}, map[string]interface{}{
// 						"error": errStr,
// 					}, nil)
// 				} else {
// 					// http.Error(rw, e.Error(), http.StatusInternalServerError)
// 					modelV.renderHtml2([]string{"public/error"}, map[string]interface{}{
// 						"error": e.Error(),
// 					}, nil)
// 				}

// 			}
// 		}()

// 		pass := true
// 		for _, f := range b {
// 			pass = f(rw, r, modelV)
// 			if !pass {
// 				return
// 			}
// 		}
// 		fn(rw, r, modelV)
// 	}
// }

// func initServer() {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", wrapHandler(index, []beforeReqHandle{omit, filterReq, checkLogin}))
// 	mux.HandleFunc("/list", wrapHandler(listTask, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/log", wrapHandler(recentLog, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/update", wrapHandler(updateTask, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/stop", wrapHandler(stopTask, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/start", wrapHandler(startTask, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/login", wrapHandler(login, []beforeReqHandle{filterReq}))
// 	mux.HandleFunc("/logout", wrapHandler(logout, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/readme", wrapHandler(readme, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/quickStart", wrapHandler(quickStart, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/reloadConfig", wrapHandler(reloadConfig, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/deleteClient", wrapHandler(deleteClient, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/viewConfig", wrapHandler(viewConfig, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/stopAllTask", wrapHandler(stopAllTask, []beforeReqHandle{filterReq, checkLogin}))
// 	mux.HandleFunc("/model", wrapHandler(model, []beforeReqHandle{filterReq, checkLogin}))

// 	if globalConfig.debug {
// 		mux.HandleFunc("/debug/pprof/", pprof.Index)
// 		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
// 		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
// 		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
// 	}

// 	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

// 	log.Println("listen to ", globalConfig.addr)
// 	log.Fatal(http.ListenAndServe(globalConfig.addr, mux))
// }
