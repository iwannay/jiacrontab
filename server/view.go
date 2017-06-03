package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type modelView struct {
	startTime float64
	shareData map[string]interface{}
	locals    map[string]interface{}
	rw        http.ResponseWriter
}

func newModelView(rw http.ResponseWriter) *modelView {
	return &modelView{
		shareData: make(map[string]interface{}),
		locals:    make(map[string]interface{}),
		rw:        rw,
	}
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
