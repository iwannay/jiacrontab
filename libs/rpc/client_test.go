package rpc

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
)

type Person struct {
}

func (p *Person) Say(args string, reply *string) error {
	fmt.Println("hello people")
	*reply = "hello boy"
	return nil
}
func TestCall(t *testing.T) {
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
		log.Println("start server")
		err := listen(":6478", &Person{})
		if err != nil {
			t.Fatal("server error:", err)
		}

	}()
	<-done

	// 等待server启动
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			var ret string
			// var args string
			err := Call(":6478", "Person.Say", "", &ret)
			if err != nil {
				t.Log(i, "error:", err)
			}
			t.Log(i, ret)

		}(i)

	}

	go func() {
		t.Log("listen :6060")
		t.Log(http.ListenAndServe(":6060", nil))
	}()

	wg.Wait()
}
