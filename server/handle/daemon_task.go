package handle

import (
	"github.com/kataras/iris"
)

func ListDaemonTask(ctx iris.Context) {

	//	ctx.Params()
	//	proto.DaemonTask{

	//	}

	// ctx.FormFile()

	ctx.View("daemon/list.html")
}

func EditDaemonTask(ctx iris.Context) {
	ctx.View("daemon/edit.html")
}
