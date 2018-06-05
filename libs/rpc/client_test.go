package rpc

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

type Person struct {
}

func (p *Person) Say(args string, reply *string) error {
	fmt.Println("hello people")
	*reply = "hahahaha"
	return nil
}
func TestCall(t *testing.T) {
	go func() {
		// t.Fatal(Start(":1234", &Person{}))

	}()

	for i := 0; i < 1000; i++ {
		go func(i int) {
			var ret string
			err := Call(":1234", "Person.Say", "", &ret)
			if err != nil {
				t.Log(err)
			}
			t.Log(i)

		}(i)

	}

	done := make(chan struct{})

	go func() {
		t.Log("listen :6060")
		t.Log(http.ListenAndServe(":6060", nil))
	}()
	<-done
}
