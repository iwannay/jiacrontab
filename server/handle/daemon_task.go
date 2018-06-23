package handle

import (
	"github.com/kataras/iris"
	"jiacrontab/libs/proto"
	"jiacrontab/libs/rpc"
)

func ListDaemonTask(ctx iris.Context) {




	ctx.View("daemon/list.html")
}

func EditDaemonTask(ctx iris.Context) {
	var err error
	name := ctx.PostValueTrim("name")
	mailNotify,err:= ctx.PostValueBool("mailNotify")
	mailTo := ctx.PostValueTrim("mailTo")
	command := ctx.PostValue("command")
	args := ctx.PostValue("args")



	rpc.Call()
	ctx.View("daemon/edit.html")
}
