package jiaweb

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

type (
	Response struct {
		rw        http.ResponseWriter
		Status    int
		Size      int64
		body      []byte
		committed bool
		header    http.Header
		isEnd     bool
	}

	gzipResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

func NewResponse(rw http.ResponseWriter) (r *Response) {
	return &Response{
		rw:     rw,
		header: rw.Header(),
	}
}

func (r *Response) ResponseWriter() http.ResponseWriter {
	return r.rw
}

func (r *Response) Write(code int, b []byte) (n int, err error) {
	if !r.committed {
		r.WriteHeader(code)
	}
	n, err = r.rw.Write(b)
	r.Size += int64(n)
	r.body = append(r.body, b...)
	return
}

func (r *Response) reset(rw http.ResponseWriter) {
	r.rw = rw
	r.header = rw.Header()
	r.Status = http.StatusOK
	r.Size = 0
	r.committed = false
}

func (r *Response) release() {
	r.rw = nil
	r.header = nil
	r.Status = http.StatusOK
	r.Size = 0
	r.committed = false
	r.body = []byte{}
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) QueryHeader(key string) string {
	return r.Header().Get(key)
}

func (r *Response) Redirect(code int, targetUrl string) error {
	r.Header().Set(HeaderCacheControl, "no-cache")
	r.Header().Set(HeaderLocation, targetUrl)
	return r.WriteHeader(code)
}

func (r *Response) SetContentType(contentType string) {
	r.SetHeader(HeaderContentType, contentType)
}

func (r *Response) SetStatusCode(code int) error {
	return r.WriteHeader(code)
}

func (r *Response) BodyString() string {
	return string(r.body)
}

func (r *Response) Body() []byte {
	return r.body
}

func (r *Response) SetHeader(key, val string) {
	r.Header().Set(key, val)
}

func (r *Response) WriteHeader(code int) error {
	if r.committed {
		return errors.New("response already set statuws")
	}

	r.Status = code
	r.rw.WriteHeader(code)
	r.committed = true
	return nil
}

func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.rw.(http.Hijacker).Hijack()
}

func (r *Response) HttpStatus() int {
	return r.Status
}
