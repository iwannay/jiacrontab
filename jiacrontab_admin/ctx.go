package admin

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/proto"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
)

type myctx struct {
	iris.Context
}

func wrapCtx(ctx iris.Context) *myctx {
	return &myctx{
		ctx,
	}
}

func (ctx *myctx) respError(code int, msg string, v interface{}) {

	var (
		sign string
		bts  []byte
		err  error
	)

	if msg == "" {
		msg = "error"
	}

	if v == nil {
		goto end
	}

	bts, err = json.Marshal(v)
	if err != nil {
		log.Error("errorResp:", err)
	}

	sign = fmt.Sprintf("%x", md5.Sum(bts))

end:
	ctx.JSON(proto.Resp{
		Code: code,
		Msg:  msg,
		Data: string(bts),
		Sign: sign,
	})
}

func (ctx *myctx) respSucc(msg string, v interface{}) {
	if msg == "" {
		msg = "success"
	}
	ctx.JSON(proto.Resp{
		Code: proto.SuccessRespCode,
		Msg:  msg,
		Data: v,
	})
}

func (ctx *myctx) getGroupIDFromToken() (int, error) {
	var data CustomerClaims
	token := ctx.Values().Get("jwt").(*jwt.Token)
	bts, err := json.Marshal(token.Claims)
	if err != nil {
		return 0, err
	}
	json.Unmarshal(bts, &data)
	log.Infof("%+v", data)
	return int(data.GroupID), nil
}
