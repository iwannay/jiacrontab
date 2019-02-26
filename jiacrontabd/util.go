package jiacrontabd

import (
	"jiacrontab/pkg/rpc"
	"jiacrontab/pkg/util"
	"os"

	"github.com/iwannay/log"
)

func rpcCall(serviceMethod string, args, reply interface{}) error {
	return rpc.Call(cfg.AdminAddr, serviceMethod, args, reply)
}

func writeFile(fPath string, content *[]byte) {
	f, err := util.TryOpen(fPath, os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Errorf("writeLog: %v", err)
		return
	}
	defer f.Close()
	f.Write(*content)
}
