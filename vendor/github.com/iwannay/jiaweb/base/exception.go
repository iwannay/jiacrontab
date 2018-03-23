package base

import (
	"fmt"
	"os"
	"runtime/debug"
)

func FormatError(title string, logtarget string, err interface{}) (errmsg string) {
	errmsg = fmt.Sprintln(err)
	stack := string(debug.Stack())
	os.Stdout.Write([]byte(title + " error! => " + errmsg + " => " + stack))
	return title + " error! => " + errmsg + " => " + stack
}
