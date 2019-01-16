package main

import (
	"context"
	"log"
	"testing"
	"time"
)

func Test_wrapExecScript(t *testing.T) {
	var content []byte
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := wrapExecScript(ctx, "test.log", [][]string{[]string{"/Users/mac/script/sleep.sh", "a"}, []string{"grep", "hello"}}, "data", &content); err != nil {
		t.Errorf("wrapExecScript() error = %v, wantErr %v,content %s", err, nil, string(content))
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	globalConfig = newConfig()
}
