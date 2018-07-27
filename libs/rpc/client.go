package rpc

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	return TimeoutCoder(c.dec.Decode, r, "client read response header")
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return TimeoutCoder(c.dec.Decode, body, "client read response body")
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

// Call 调用
func Call(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	bts, _ := json.Marshal(args)
	log.Printf("RPC call %s %s %s ", addr, serviceMethod, string(bts))
	conn, err := net.DialTimeout("tcp4", addr, time.Second*10)
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
	}
	return errC

}
