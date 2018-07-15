package model

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"jiacrontab/libs"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	stateUpdate     = "update"
	stateSync       = "sync"
	stateSelect     = "select"
	stateSelectCopy = "selectCopy"
	stateLoad       = "load"
	stateSearch     = "Search"
)

type result struct {
	value interface{}
	err   error
}
type request struct {
	key     string
	state   string
	handler handle
	body    string

	response chan<- result
}

type handle func(s *Store)

func NewStore(path string) *Store {
	s := &Store{
		dataFile: path,
		swapFile: filepath.Join(filepath.Dir(path), ".swap"),
		requests: make(chan request),
		// RpcClientList: make(map[string]proto.ClientConf),
	}
	go s.server()
	return s
}

type Store struct {
	// RpcClientList map[string]proto.ClientConf
	dataFile string
	swapFile string
	requests chan request
}

func (s *Store) server() {
	t := time.Tick(time.Duration(10 * time.Minute))
	for {
		select {
		case req := <-s.requests:
			s.requestHandle(req)
		case <-t:
			s.sync(s.swapFile)
		}

	}
}

func (s *Store) requestHandle(req request) {
	var ret result

	if req.handler != nil {
		req.handler(s)
	}
	if req.state == stateUpdate {
		// s.sync()
	}

	if req.state == stateSync {
		s.sync(s.dataFile)
	}

	if req.state == stateLoad {
		if err := s.load(s.dataFile); err != nil {
			log.Printf("failed recover %s", err)
			if err = s.load(s.swapFile); err != nil {
				log.Printf("failed recover %s", err)
				ret.err = err
			}
		}
	}

	switch req.key {
	// case "RpcClientList":
	// 	if req.state == stateSearch && req.body != "" {
	// 		if v, ok := s.RpcClientList[req.body]; ok {
	// 			req.response <- result{value: v}
	// 		} else {
	// 			req.response <- result{value: nil}
	// 		}
	// 	} else {
	// 		var rpcClientList map[string]proto.ClientConf
	// 		if b, err := json.Marshal(s.RpcClientList); err == nil {
	// 			json.Unmarshal(b, &rpcClientList)
	// 		}
	// 		req.response <- result{value: rpcClientList}
	// 	}

	case "dataFile":
		req.response <- result{value: s.dataFile}
	default:
		req.response <- result{value: nil}
	}

}

// func (s *Store) getRPCClientList() (map[string]proto.ClientConf, bool) {
// 	ret, ok := (s.Get("RpcClientList")).value.(map[string]proto.ClientConf)
// 	return ret, ok
// }

// func (s *Store) searchRPCClientList(args string) (proto.ClientConf, bool) {
// 	ret, ok := s.Search("RpcClientList", args).value.(proto.ClientConf)
// 	return ret, ok
// }

func (s *Store) Get(key string) result {
	return s.Query(key, stateSelect, nil, "")

}
func (s *Store) Search(key, args string) result {
	return s.Query(key, stateSearch, nil, args)
}

func (s *Store) Wrap(fn handle) *Store {
	s.Query("", stateUpdate, fn, "")
	return s
}

func (s *Store) Sync() result {
	return s.Query("", stateSync, nil, "")
}

func (s *Store) Load() result {
	return s.Query("", stateLoad, nil, "")
}

func (s *Store) Query(key string, state string, fn handle, body string) result {
	response := make(chan result)
	s.requests <- request{key, state, fn, body, response}
	return <-response
}

func (s *Store) sync(fpath string) error {

	f, err := libs.TryOpen(fpath, os.O_CREATE|os.O_RDWR|os.O_TRUNC)
	defer func() {
		f.Close()
	}()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(s, "", "  ")

	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err

}

func (s *Store) load(fpath string) error {

	f, err := libs.TryOpen(fpath, os.O_CREATE|os.O_RDWR)
	defer func() {
		f.Close()
	}()

	if err != nil {

		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {

		return err
	}

	if len(b) == 0 {
		err = errors.New("nothing to read from " + fpath)

		return err
	}
	err = json.Unmarshal(b, &s)

	return err
}
