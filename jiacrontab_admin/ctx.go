package admin

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"jiacrontab/models"
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

func (ctx *myctx) getClaimsFromToken() (CustomerClaims, error) {
	var data CustomerClaims
	token, ok := ctx.Values().Get("jwt").(*jwt.Token)
	if !ok {
		return data, errors.New("claims is nil")
	}
	bts, err := json.Marshal(token.Claims)
	if err != nil {
		return data, err
	}
	json.Unmarshal(bts, &data)
	return data, err
}

func (ctx *myctx) pubEvent(desc, nodeAddr string, v interface{}) {
	content := ""
	claims, err := ctx.getClaimsFromToken()
	if err != nil {
		return
	}

	if v != nil {
		bts, err := json.Marshal(v)
		if err != nil {
			return
		}
		content = string(bts)
	}

	e := models.Event{
		GroupID:   claims.GroupID,
		UserID:    claims.UserID,
		Username:  claims.Username,
		EventDesc: desc,
		NodeAddr:  nodeAddr,
		Content:   content,
	}
	e.Pub()
}
