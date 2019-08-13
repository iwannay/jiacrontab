// +build !windows

package pprof

import (
	"os"
	"os/signal"
	"syscall"
)

func listenSignal() {
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGUSR1)
	for {
		<-signChan
		profile()
		memprofile()
		cpuprofile()
	}
}
