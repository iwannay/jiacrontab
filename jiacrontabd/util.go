package jiacrontabd

import (
	"jiacrontab/pkg/util"
	"os"

	"github.com/iwannay/log"
)

func writeFile(fPath string, content *[]byte) {
	f, err := util.TryOpen(fPath, os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Errorf("writeLog: %v", err)
		return
	}
	defer f.Close()
	f.Write(*content)
}
