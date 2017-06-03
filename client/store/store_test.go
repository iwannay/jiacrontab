package store

import (
	"fmt"
	"jiacrontab/libs/proto"
	"sync"
	"testing"
)

func TestMul(t *testing.T) {
	store := NewStore("../.data/data.json")
	store.Load()
	t.Log(store.GetDataFile())
	t.Log(store.GetMail())
	var w sync.WaitGroup
	t.Log(store.SearchTaskList("12348012348192034812934"))

	for i := 0; i < 100; i++ {
		w.Add(1)
		go func(i int) {
			if i%2 == 0 {
				store.Update(func(s *Store) {
					s.TaskList[fmt.Sprint(i)] = &proto.TaskArgs{}
				}).Sync()
			} else {
				store.GetTaskList()
				// t.Logf("%#v", ret)

			}

			w.Done()
		}(i)
	}
	t.Log(store.SearchTaskList("2"))

	w.Wait()
	t.Log("end")

}
