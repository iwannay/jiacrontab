package jiaweb

import (
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	defaultMemory int64 = 32 << 20
)

type (
	Request struct {
		*http.Request
		httpCtx     *HttpContext
		body        []byte
		isReadyBody bool
	}
)

func (req *Request) reset(r *http.Request, ctx *HttpContext) {
	req.Request = r
	req.httpCtx = ctx
}

func (req *Request) Get(key string) string {
	return req.URL.Query().Get(key)
}

func (req *Request) IsPost() bool {
	return req.Method == http.MethodPost
}

func (req *Request) Post(key string) string {
	return req.PostFormValue(key)
}

func (req *Request) parseForm() error {
	if strings.HasPrefix(req.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return err
		}
	} else {
		if err := req.ParseForm(); err != nil {
			return err
		}
	}
	return nil
}

func (req *Request) Body() []byte {
	if !req.isReadyBody {
		bts, err := ioutil.ReadAll(req.Request.Body)
		if err != nil {
			return []byte{}
		}
		req.isReadyBody = true
		req.body = bts
	}
	return req.body
}

func (req *Request) FormValues() map[string][]string {
	req.parseForm()
	return map[string][]string(req.Form)
}

func (req *Request) release() {
	req.Request = nil
	req.isReadyBody = false
	req.body = nil
}

func (req *Request) RemoteIP() string {
	return req.Request.RemoteAddr
}

func (req *Request) Url() string {
	return req.URL.String()
}

func (req *Request) Path() string {
	return req.URL.Path
}

func (req *Request) IsAJAX() bool {
	return req.Header.Get(HeaderXRequestedWith) == "XMLHttpRequest"
}
