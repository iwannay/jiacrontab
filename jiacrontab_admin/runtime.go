package admin

import (
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

func SystemInfo(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		info    map[string]interface{}
		reqBody SystemInfoReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if err = rpcCall(reqBody.Addr, "Srv.SystemInfo", proto.EmptyArgs{}, &info); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.respSucc("", info)
}
