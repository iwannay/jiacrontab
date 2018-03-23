package jiaweb

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

type (
	Viewer interface {
		AppendTpl(tpl ...string)
		AppendFunc(funcMap template.FuncMap)
		RenderHtml(rw *Response, viewPath []string, locals map[string]interface{}) error
		AddLocals(val ...KValue)
		Tpls() []string
	}

	view struct {
		funcMap  template.FuncMap
		innerTpl []string
		server   *HttpServer
		locals   []KValue
		// mutex    sync.RWMutex
	}

	KValue struct {
		Key   string
		Value interface{}
	}
)

func NewView(s *HttpServer) *view {
	return &view{
		server:  s,
		funcMap: make(template.FuncMap),
	}
}

func (v *view) Tpls() []string {
	return v.innerTpl
}

func (v *view) AddLocals(val ...KValue) {
	v.locals = append(v.locals, val...)
}

func (v *view) AppendTpl(tpl ...string) {
	var tplPaths []string
	var tplPath string
	for _, item := range tpl {
		tplPath = filepath.Join(".", v.server.TemplateConfig().TplDir, item+v.server.TemplateConfig().TplExt)
		tplPaths = append(tplPaths, tplPath)
	}

	v.innerTpl = append(v.innerTpl, tplPaths...)
}

func (v *view) AppendFunc(funcMap template.FuncMap) {
	v.funcMap = funcMap
}

func (v *view) RenderHtml(rw *Response, viewPath []string, locals map[string]interface{}) error {
	var tplPaths []string
	var tplName string
	var tplPath string
	var startTime = time.Now()

	for _, item := range viewPath {
		tplPath = filepath.Join(".", v.server.TemplateConfig().TplDir, item+v.server.TemplateConfig().TplExt)
		tplPaths = append(tplPaths, tplPath)
		if tplName == "" {
			tplName = filepath.Base(tplPath)
		}
	}

	tplPaths = append(v.innerTpl, tplPaths...)

	if locals == nil {
		locals = make(map[string]interface{})

	}
	for _, item := range v.locals {
		locals[item.Key] = item.Value
	}

	t := template.Must(template.New(tplName).Funcs(v.funcMap).ParseFiles(tplPaths...))

	subTime := time.Now().Sub(startTime).Nanoseconds()

	locals["costTime"] = fmt.Sprintf("%.5fms", float64(subTime))
	err := t.ExecuteTemplate(rw.ResponseWriter(), tplName, locals)
	return err
}
