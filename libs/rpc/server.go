package rpc

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"runtime/debug"
	"time"
)

var DefaultTimeout = 20

// TimeoutCoder 超时检测
func TimeoutCoder(f func(interface{}) error, e interface{}, msg string) error {
	endChan := make(chan error, 1)
	go func() { endChan <- f(e) }()
	timer := time.NewTimer(time.Duration(DefaultTimeout) * time.Second)
	select {
	case e := <-endChan:
		return e
	case <-timer.C:
		timer.Stop()
		return fmt.Errorf("Timeout %s", msg)
	}
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *gobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return TimeoutCoder(c.dec.Decode, r, "server read request header")
}

func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
	return TimeoutCoder(c.dec.Decode, body, "server read request body")
}

func (c *gobServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = TimeoutCoder(c.enc.Encode, r, "server write response"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = TimeoutCoder(c.enc.Encode, body, "server write response body"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

// listen Start rpc server
func listen(addr string, srcvr ...interface{}) error {
	var err error
	for _, v := range srcvr {
		if err = rpc.Register(v); err != nil {
			return err
		}
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error: accept rpc connection", err.Error(), l.Addr().String())
			continue
		}
		go func(conn net.Conn) {
			defer func() {
				if err := recover(); err != nil {
					log.Println("Error: Rpc Call Recover", err, string(debug.Stack()))
				}
			}()

			log.Println(conn.RemoteAddr().String(), "->", conn.LocalAddr().String())

			buf := bufio.NewWriter(conn)
			srv := &gobServerCodec{
				rwc:    conn,
				dec:    gob.NewDecoder(conn),
				enc:    gob.NewEncoder(buf),
				encBuf: buf,
			}
			defer srv.Close()
			err = rpc.ServeRequest(srv)

			if err != nil {
				log.Print("Error: rpc server", err.Error())
			}

		}(conn)
	}
}

// ListenAndServe  run rpc server
func ListenAndServe(addr string, srcvr ...interface{}) {
	err := listen(addr, srcvr...)
	if err != nil {
		panic(err)
	}
	log.Println("rpc server listen:", addr)

}
