package admin

import (
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

func runtimeInfo(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		info    map[string]interface{}
		reqBody runtimeInfoReqParams
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err = rpcCall(reqBody.Addr, "Admin.SystemInfo", "", &info); err != nil {
		goto failed
	}

	ctx.respSucc("", info)
	return

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)
}
