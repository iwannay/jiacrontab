package rpc

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"
)

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = TimeoutCoder(c.enc.Encode, r, "client write request"); err != nil {
		return
	}
	if err = TimeoutCoder(c.enc.Encode, body, "client write request body"); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

// Call 调用
func Call(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		return err
	}
	encBuf := bufio.NewWriter(conn)
	codec := &gobClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
	c := rpc.NewClientWithCodec(codec)
	err = c.Call(serviceMethod, args, reply)
	errC := c.Close()
	if err != nil && errC != nil {
		return fmt.Errorf("%s %s", err, errC)
	}
	if err != nil {
		return err
	} else {
		return errC
	}
}
