package admin

import (
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/version"
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
	info["version"] = version.String("jiacrontab")
	ctx.respSucc("", info)
}
