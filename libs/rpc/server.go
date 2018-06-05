package rpc

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"time"
)

// TimeoutCoder 超时检测
func TimeoutCoder(f func(interface{}) error, e interface{}, msg string) error {
	endChan := make(chan error, 1)
	go func() { endChan <- f(e) }()
	timer := time.NewTimer(time.Minute)
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

func Start(srcvr ...interface{}) {
	var err error
	server := rpc.NewServer()
	for _, v := range srcvr {
		if err = server.Register(v); err != nil {
			return err
		}
	}
	server.HandleHTTP(globalConfig.defaultRPCPath, globalConfig.defaultRPCDebugPath)

	l, err := net.Listen("tcp", globalConfig.rpcListenAddr)
	if err != nil {
		return err
	}
	log.Printf("rpc listen %s", globalConfig.rpcListenAddr)

	return http.Serve(l, nil)
}

// func ListenRPC() {
// 	rpc.Register(NewWorker())
// 	l, e := net.Listen("tcp", ":4200")
// 	if e != nil {
// 		log.Fatal("Error: listen 4200 error:", e)
// 	}
// 	go func() {
// 		for {
// 			conn, err := l.Accept()
// 			if err != nil {
// 				log.Print("Error: accept rpc connection", err.Error())
// 				continue
// 			}
// 			go func(conn net.Conn) {
// 				buf := bufio.NewWriter(conn)
// 				srv := &gobServerCodec{
// 					rwc:    conn,
// 					dec:    gob.NewDecoder(conn),
// 					enc:    gob.NewEncoder(buf),
// 					encBuf: buf,
// 				}
// 				err = rpc.ServeRequest(srv)
// 				if err != nil {
// 					log.Print("Error: server rpc request", err.Error())
// 				}
// 				srv.Close()
// 			}(conn)
// 		}
// 	}()
// }

// func main() {
// 	go ListenRPC()
// 	N := 1000
// 	mapChan := make(chan int, N)
// 	for i := 0; i < N; i++ {
// 		go func(i int) {
// 			call("localhost", "Worker.DoJob", strconv.Itoa(i), new(string))
// 			mapChan <- i
// 		}(i)
// 	}
// 	for i := 0; i<N; i++ {
// 		<-mapChan
// 	}

// }
