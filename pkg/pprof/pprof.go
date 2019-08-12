package pprof

import (
	"jiacrontab/pkg/file"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/iwannay/log"
)

func ListenPprof() {
	go listenSignal()
}

func cpuprofile() {
	path := filepath.Join("pprof", "cpuprofile")
	log.Debugf("profile save in %s", path)

	f, err := file.CreateFile(path)
	if err != nil {
		log.Error("could not create CPU profile: ", err)
		return
	}

	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Error("could not start CPU profile: ", err)
	} else {
		time.Sleep(time.Minute)
	}
	defer pprof.StopCPUProfile()
}

func memprofile() {
	path := filepath.Join("pprof", "memprofile")
	log.Debugf("profile save in %s", path)
	f, err := file.CreateFile(path)
	if err != nil {
		log.Error("could not create memory profile: ", err)
		return
	}

	defer f.Close()

	runtime.GC() // get up-to-date statistics

	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Error("could not write memory profile: ", err)
	}
}

func profile() {
	names := []string{
		"goroutine",
		"heap",
		"allocs",
		"threadcreate",
		"block",
		"mutex",
	}
	for _, name := range names {
		path := filepath.Join("pprof", name)
		log.Debugf("profile save in %s", path)
		f, err := file.CreateFile(path)
		if err != nil {
			log.Error(err)
			continue
		}
		pprof.Lookup(name).WriteTo(f, 0)
	}

}
