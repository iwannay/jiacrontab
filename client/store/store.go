package store

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"jiacrontab/model"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	Update = "update"
	Sync   = "sync"
	Select = "select"
	Load   = "load"
	Search = "Search"
)

type result struct {
	value interface{}
}
type request struct {
	key      string
	state    string
	handler  handle
	body     string
	response chan<- result
}

type handle func(s *Store)

func NewStore(path string) *Store {
	s := &Store{
		dataFile: path,
		swapFile: filepath.Join(filepath.Dir(path), ".swap"),
		requests: make(chan request),
	}
	go s.Server()
	return s
}

type Store struct {
	Mail     proto.MailArgs
	TaskList map[string]model.CrontabTask

	swapFile string
	dataFile string
	requests chan request
}

func (s *Store) Server() {
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
	if req.handler != nil {
		req.handler(s)
	}
	if req.state == Update {
		// s.sync()
	}

	if req.state == Sync {
		s.sync(s.dataFile)
	}

	if req.state == Load {
		if err := s.load(s.dataFile); err != nil {
			log.Printf("failed recover %s", err)
			if err = s.load(s.swapFile); err != nil {
				log.Printf("failed recover %s", err)
			}
		}
	}
	switch req.key {
	case "Mail":
		req.response <- result{value: s.Mail}
	// case "TaskList":
	// 	if req.state == Search && req.body != "" {
	// 		if v, ok := s.TaskList[req.body]; ok {
	// 			req.response <- result{value: v}
	// 		} else {
	// 			req.response <- result{value: nil}
	// 		}
	// 	} else {
	// 		var taskList proto.Mdata
	// 		if b, err := json.Marshal(s.TaskList); err == nil {
	// 			json.Unmarshal(b, &taskList)
	// 		}
	// 		req.response <- result{value: taskList}
	// 	}

	case "dataFile":
		req.response <- result{value: s.dataFile}
	default:
		req.response <- result{value: nil}
	}
}

func (s *Store) Get(key string) result {
	return s.Query(key, Select, nil, "")
}

func (s *Store) Search(key, args string) result {
	return s.Query(key, Search, nil, args)
}

// func (s *Store) SearchTaskList(args string) (*proto.TaskArgs, bool) {
// 	ret, ok := s.Search("TaskList", args).value.(*proto.TaskArgs)
// 	return ret, ok
// }

func (s *Store) Update(fn handle) *Store {
	s.Query("", Update, fn, "")
	return s
}

func (s *Store) Sync() result {

	return s.Query("", Sync, nil, "")
}

func (s *Store) Load() result {
	return s.Query("", Load, nil, "")
}

func (s *Store) Query(key string, state string, fn handle, body string) result {
	response := make(chan result)
	s.requests <- request{key, state, fn, body, response}
	return <-response
}

func (s *Store) GetMail() (proto.MailArgs, bool) {
	ret, ok := (s.Get("Mail")).value.(proto.MailArgs)
	return ret, ok
}

// func (s *Store) GetRpcClient() (proto.Mdata, bool) {
// 	ret, ok := (s.Get("RpcClient")).value.(proto.Mdata)
// 	return ret, ok
// }

func (s *Store) GetDataFile() (string, bool) {
	ret, ok := (s.Get("dataFile")).value.(string)
	return ret, ok
}

// func (s *Store) GetTaskList() (proto.Mdata, bool) {
// 	ret, ok := (s.Get("TaskList")).value.(proto.Mdata)
// 	return ret, ok
// }

func (s *Store) Export2DB() {
	for _, v := range s.TaskList {
		ret := model.DB().Create(&v)
		if ret.Error == nil {
			log.Printf("import crontab %+v \n", v)
		} else {
			log.Printf("failed import crontab %+v \n", v, ret.Error)
		}

	}
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

	err = json.Unmarshal(b, s)
	if err == nil {
		for _, v := range s.TaskList {
			if v.MaxConcurrent == 0 {
				v.MaxConcurrent = 1
			}
			v.NumberProcess = 0
			v.TimerCounter = 0
		}
	}
	return err
}
