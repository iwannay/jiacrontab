package base

import (
	"sync"
)

type (
	Store struct {
		sync.Map
	}
)

func NewStore() *Store {
	return &Store{}
}

func (s *Store) All() map[string]interface{} {
	data := make(map[string]interface{})
	s.Range(func(key, value interface{}) bool {

		data[key.(string)] = value
		return true
	})
	return data
}

func (s *Store) Exists(key interface{}) bool {

	_, ok := s.Load(key)

	return ok
}

func (s *Store) GetUint64(key interface{}) (uint64, bool) {
	val, ok := s.Load(key)
	if !ok {
		return 0, false
	}

	ret, ok := val.(uint64)
	return ret, ok
}

func (s *Store) Len() uint {
	var count uint
	s.Range(func(key, val interface{}) bool {
		count++
		return true
	})

	return count

}
