package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"jiacrontab/libs"
	"os"
	"sync"
)

func newStore(path string) *Store {
	s := &Store{
		dataFile:      path,
		RpcClientList: make(map[string]*mrpcClient),
	}
	go s.server()
	return s
}

type Store struct {
	RpcClientList map[string]*mrpcClient
	rpcClient     *mrpcClient

	dataFile string
	lock     sync.RWMutex

	requests chan request
}
type result struct {
	value interface{}
}
type request struct {
	key      string
	response chan<- result
}

func (s *Store) server() {
	for {

		select {
		case req := <-s.requests:
			switch req.key {
			case "RpcClient":
				req.response <- result{value: s.rpcClient}
			case "rpcClient":
				req.response <- result{value: s.rpcClient}
			case "dataFile":
				req.response <- result{value: s.dataFile}
			}
		}

	}
}

func (s *Store) get(key string) result {
	response := make(chan result)
	s.requests <- request{key, response}
	return <-response
}

func (s *Store) getRpcClientList(key string) (map[string]*mrpcClient, bool) {
	ret, ok := (s.get(key)).value.(map[string]*mrpcClient)
	return ret, ok
}

func (s *Store) getRpcClient(key string) (*mrpcClient, bool) {
	ret, ok := (s.get(key)).value.(*mrpcClient)
	return ret, ok
}

func (s *Store) getDataFile() (string, bool) {
	ret, ok := (s.get("dataFile")).value.(string)
	return ret, ok
}

func (s *Store) Update(fn func()) error {
	if fn != nil {
		fn()
	}
	return s.sync()
}

func (s *Store) sync() error {

	s.lock.Lock()

	f, err := libs.TryOpen(s.dataFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC)
	defer func() {
		f.Close()
		s.lock.Unlock()
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

func (s *Store) Load() error {

	s.lock.Lock()

	f, err := libs.TryOpen(s.dataFile, os.O_CREATE|os.O_RDWR)
	defer func() {
		f.Close()
		s.lock.Unlock()
	}()

	if err != nil {

		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {

		return err
	}

	if len(b) == 0 {
		err = errors.New("nothing to read")

		return err
	}

	err = json.Unmarshal(b, s)
	return err
}
