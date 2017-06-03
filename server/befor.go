package main

import "net/http"
import "runtime"

func filterReq(rw http.ResponseWriter, r *http.Request, m *modelView) bool {
	if r.URL.Path == "/favicon.ico" {
		return false
	}

	m.locals["action"] = r.URL.Path
	m.locals["appInfo"] = globalConfig.appName + " 当前版本:" + globalConfig.version
	m.locals["goVersion"] = runtime.Version()
	m.locals["appName"] = globalConfig.appName
	m.locals["version"] = globalConfig.version

	if err := globalReqFilter.filter(rw, r); err != nil {
		m.renderHtml2([]string{"public/error"}, map[string]interface{}{
			"error": err.Error(),
		}, nil)
		return false
	} else {
		return true
	}
}

func omit(rw http.ResponseWriter, r *http.Request, m *modelView) bool {
	if r.URL.Path == "/favicon.ico" {
		return false
	} else {
		return true
	}
}

func checkLogin(rw http.ResponseWriter, r *http.Request, m *modelView) bool {

	var userinfo map[string]interface{}
	ok := globalJwt.auth(rw, r, &userinfo)
	if ok {
		for k, v := range userinfo {
			m.shareData[k] = v
			m.locals[k] = v
		}
	}
	if !ok && r.URL.Path == "/login" {
		ok = true
	}

	if !ok {
		http.Redirect(rw, r, "/login", http.StatusFound)
	}
	return ok
}
