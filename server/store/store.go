package store

import (
	"encoding/json"
	"errors"
	"fmt"
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

type Result struct {
	Value interface{}
	Err   error
}
type request struct {
	key     string
	state   string
	handler handle
	body    interface{}

	response chan<- Result
}
type data map[string]data

type handle func(s *Store)

func NewStore(path string) *Store {
	s := &Store{
		dataFile: path,
		swapFile: filepath.Join(filepath.Dir(path), ".swap"),
		requests: make(chan request),
		Data:     make(map[string]interface{}),
	}
	go s.server()
	return s
}

type Store struct {
	Data     map[string]interface{}
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
	var ret Result

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
				ret.Err = err
			}
		}
	}

	if req.key == "" {
		if req.state == stateSelectCopy {
			ret.Value = libs.DeepCopy2(s.Data)
		} else {
			ret.Value = s.Data
		}

		req.response <- ret
		return
	}

	ret.Value = libs.DeepFind(s.Data, req.key)
	if req.state == stateSelectCopy {
		ret.Value = libs.DeepCopy2(ret.Value)
	}

	req.response <- ret

}

func (s *Store) Get(key string, v interface{}) error {
	ret := s.Query(key, stateSelect, nil, "")
	if ret.Value != nil {
		b, err := json.Marshal(ret.Value)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, v)
	}

	return fmt.Errorf("failed to get %s", key)
}

func (s *Store) GetCopy(key string, v interface{}) error {
	ret := s.Query(key, stateSelectCopy, nil, "")
	if ret.Value != nil {
		b, err := json.Marshal(ret.Value)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, v)
	}

	return fmt.Errorf("failed to get %s", key)
}

func (s *Store) Wrap(fn handle) *Store {
	s.Query("", stateUpdate, fn, "")
	return s
}

func (s *Store) Sync() Result {
	return s.Query("", stateSync, nil, "")
}

func (s *Store) Load() Result {
	return s.Query("", stateLoad, nil, "")
}

func (s *Store) Query(key string, state string, fn handle, body interface{}) Result {
	response := make(chan Result)
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

	b, err := json.MarshalIndent(s.Data, "", "  ")

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
	err = json.Unmarshal(b, &s.Data)

	return err
}
