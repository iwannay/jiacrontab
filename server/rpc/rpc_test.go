package rpc

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"
)

type Person struct {
}

func (p *Person) Say(args string, reply *string) error {
	fmt.Println("hello people")
	*reply = "hahahaha"
	return nil
}

func TestMrpcClient_Call(t *testing.T) {
	go InitSrvRpc("/__myrpc__", "/debug/rpc", ":2003", &Person{})
	time.Sleep(1 * time.Second)
	c, err := NewRpcClient(":2003")

	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		go func() {
			var ret string
			err := c.Call("Person.Say", "", &ret)
			if err != nil {
				t.Error(err)
			}
		}()

	}

	done := make(chan struct{})
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	<-done

}
