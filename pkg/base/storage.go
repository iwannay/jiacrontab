package base

import (
	"sync"
)

type (
	Storage struct {
		sync.Map
	}
)

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) All() map[string]interface{} {
	data := make(map[string]interface{})
	s.Range(func(key, value interface{}) bool {

		data[key.(string)] = value
		return true
	})
	return data
}

func (s *Storage) Exists(key interface{}) bool {

	_, ok := s.Load(key)

	return ok
}

func (s *Storage) GetUint64(key interface{}) (uint64, bool) {
	val, ok := s.Load(key)
	if !ok {
		return 0, false
	}

	ret, ok := val.(uint64)
	return ret, ok
}

func (s *Storage) Len() uint {
	var count uint
	s.Range(func(key, val interface{}) bool {
		count++
		return true
	})

	return count

}
