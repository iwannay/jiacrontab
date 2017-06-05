package main

import (
	"fmt"
	"html/template"
	"jiacrontab/libs/proto"
	"jiacrontab/server/rpc"
	"jiacrontab/server/store"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type modelView struct {
	startTime float64
	shareData map[string]interface{}
	locals    map[string]interface{}
	s         *store.Store
	rw        http.ResponseWriter
}

func newModelView(rw http.ResponseWriter, s *store.Store) *modelView {
	return &modelView{
		shareData: make(map[string]interface{}),
		locals:    make(map[string]interface{}),
		rw:        rw,
		s:         s,
	}
}

func (self *modelView) rpcCall(addr string, method string, args interface{}, reply interface{}) error {

	c, err := rpc.NewRpcClient(addr)
	if err != nil {
		self.s.Wrap(func(s *store.Store) {
			if tmp, ok := s.Data["RPCClientList"].(map[string]proto.ClientConf); ok {
				tmp[addr] = proto.ClientConf{
					Addr:  addr,
					State: 0,
				}
				s.Data["RPCClientList"] = tmp
			}

		}).Sync()
		log.Println(err)
		return err
	}

	if err := c.Call(method, args, reply); err != nil {
		err = fmt.Errorf("failded to call %s %s %s", method, args, err)
		log.Println(err)
	}
	return err

}

func (self *modelView) renderHtml(viewPath []string, locals map[string]interface{}, funcMap template.FuncMap) error {
	var fp []string
	var tplName string
	var tempStart = float64(time.Now().UnixNano())
	for _, v := range viewPath {
		tmp := filepath.Join(".", globalConfig.tplDir, v+globalConfig.tplExt)
		fp = append(fp, tmp)
		if tplName == "" {
			tplName = filepath.Base(tmp)
		}

	}
	if locals == nil {
		locals = make(map[string]interface{})
	}
	for k, v := range self.locals {
		locals[k] = v
	}
	t := template.Must(template.New(fp[0]).Funcs(funcMap).ParseFiles(fp...))

	endTime := float64(time.Now().UnixNano())
	tempCostTime := fmt.Sprintf("%.5fms", (endTime-tempStart)/1000000)

	locals["pageCostTime"] = fmt.Sprintf("%.5fms", (endTime-self.startTime)/1000000)
	locals["tempCostTime"] = tempCostTime
	err := t.ExecuteTemplate(self.rw, tplName, locals)
	if err != nil {
		log.Println(err)
	}
	self.rw.Header().Set("Content-Type", "text/html")
	return err
}

// include user info and template header footer
func (self *modelView) renderHtml2(viewPath []string, locals map[string]interface{}, funcMap template.FuncMap) error {
	var fp []string
	var tplName string
	var tempStart = float64(time.Now().UnixNano())
	// pubViews := []string{viewPath, "header", "footer"}
	viewPath = append(viewPath, []string{"public/head", "public/header", "public/footer"}...)
	for _, v := range viewPath {
		tmp := filepath.Join(".", globalConfig.tplDir, v+globalConfig.tplExt)
		fp = append(fp, tmp)
		if tplName == "" {
			tplName = filepath.Base(tmp)
		}

	}

	if locals == nil {
		locals = make(map[string]interface{})
	}
	for k, v := range self.locals {
		locals[k] = v
	}

	t := template.Must(template.New(fp[0]).Funcs(funcMap).ParseFiles(fp...))

	endTime := float64(time.Now().UnixNano())
	tempCostTime := fmt.Sprintf("%.5fms", (endTime-tempStart)/1000000)

	locals["pageCostTime"] = fmt.Sprintf("%.5fms", (endTime-self.startTime)/1000000)
	locals["tempCostTime"] = tempCostTime

	err := t.ExecuteTemplate(self.rw, tplName, locals)

	if err != nil {
		log.Println(err)
	}
	self.rw.Header().Set("Content-Type", "text/html")

	return err
}
