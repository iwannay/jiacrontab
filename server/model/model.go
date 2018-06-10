package model

import (
	"fmt"
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
	"log"
)

var innerStore *Store

func InitStore(path string) {
	innerStore = NewStore(path)
}

func recordError(err error) {
	if err != nil {
		log.Println(err)
	}
}

type Model struct {
	// shareData map[string]interface{}
	// locals    map[string]interface{}
	// s         *store.Store
	// rw        http.ResponseWriter
}

// func NewModelView(s *store.Store) *Model {
// 	return &Model{
// 		shareData: make(map[string]interface{}),
// 		locals:    make(map[string]interface{}),
// 		s:         s,
// 	}
// }

func NewModel() *Model {
	return &Model{}
}

func (self *Model) InnerStore() *Store {
	return innerStore
}

func (self *Model) GetRPCClientList() (map[string]proto.ClientConf, bool) {
	return innerStore.getRPCClientList()
}

func (self *Model) SearchRPCClientList(args string) (proto.ClientConf, bool) {
	return innerStore.searchRPCClientList(args)
}

func (self *Model) RpcCall(addr string, method string, args interface{}, reply interface{}) (err error) {
	defer recordError(err)

	v, ok := innerStore.searchRPCClientList(addr)
	if !ok {
		return fmt.Errorf("cannot found %s", addr)
	}

	err = rpc.Call(addr, method, args, reply)

	if err != nil {
		innerStore.Wrap(func(s *Store) {
			v.State = 0
			s.RpcClientList[addr] = v

		}).Sync()
		return err
	}
	return nil
}
