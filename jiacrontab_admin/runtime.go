package admin

import (
	"jiacrontab/pkg/proto"
)

func SystemInfo(ctx *myctx) {
	var (
		err     error
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
