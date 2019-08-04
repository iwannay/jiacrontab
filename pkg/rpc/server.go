package rpc

import (
	"github.com/iwannay/log"
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
		log.Info("listen rpc", addr, "close")
		if err := l.Close(); err != nil {
			log.Infof("listen.Close() error(%v)", err)
		}
	}()

	rpc.Accept(l)
	return nil
}

// ListenAndServe  run rpc server
func ListenAndServe(addr string, srcvr ...interface{}) {
	log.Info("rpc server listen:", addr)
	err := listen(addr, srcvr...)
	if err != nil {
		panic(err)
	}

}
