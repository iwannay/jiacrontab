package rpc

import (
	"log"
	"net"
	"net/rpc"
)

// listen Start rpc server
func listen(addr string, srcvr ...interface{}) error {
	var err error
	for _, v := range srcvr {
		if err = rpc.Register(v); err != nil {
			return err
		}
	}

	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	defer func() {
		log.Println("listen rpc", addr, "close")
		if err := l.Close(); err != nil {
			log.Printf("listen.Close() error(%v)", err)
		}
	}()

	rpc.Accept(l)
	return nil
}

// ListenAndServe  run rpc server
func ListenAndServe(addr string, srcvr ...interface{}) {
	err := listen(addr, srcvr...)
	if err != nil {
		panic(err)
	}
	log.Println("rpc server listen:", addr)

}
