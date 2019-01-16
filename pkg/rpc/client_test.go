package rpc

import (
	"jiacrontab/libs/proto"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

type Logic struct {
}

func (l *Logic) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}

func (p *Logic) Say(args string, reply *string) error {

	*reply = "hello boy"
	time.Sleep(100 * time.Second)
	return nil
}
func TestCall(t *testing.T) {
	done := make(chan struct{})
	go func() {

		done <- struct{}{}

		log.Println("start server")
		err := listen(":6478", &Logic{})
		if err != nil {
			t.Fatal("server error:", err)
		}
	}()
	<-done
	time.Sleep(5 * time.Second)
	// 等待server启动
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {

			defer wg.Done()
			var ret string
			// var args string
			err := Call(":6478", "Logic.Say", "", &ret)
			if err != nil {
				log.Println(i, "error:", err)
			}
			t.Log(i, ret)
		}(i)

	}

	go func() {
		t.Log("listen :6060")
		t.Log(http.ListenAndServe(":6060", nil))
	}()

	wg.Wait()
	log.Println("end")
	time.Sleep(2 * time.Minute)
}
