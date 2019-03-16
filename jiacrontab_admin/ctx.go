package admin

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/iwannay/log"

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

func (ctx *myctx) respNotAllowed() {
	ctx.respError(proto.Code_NotAllowed, proto.Msg_NotAllowed)
}

func (ctx *myctx) respJWTError(err error) {
	ctx.respError(proto.Code_JWTError, err)
}

func (ctx *myctx) respBasicError(err error) {
	ctx.respError(proto.Code_Error, err)
}

func (ctx *myctx) respRPCError(err error) {
	ctx.respError(proto.Code_RPCError, err)
}

func (ctx *myctx) respError(code int, err interface{}, v ...interface{}) {

	var (
		sign   string
		bts    []byte
		msgStr string
		data   interface{}
	)

	if err == nil {
		msgStr = "error"
	}

	msgStr = fmt.Sprintf("%s", err)
	if len(v) >= 1 {
		data = v[0]
	}

	bts, err = json.Marshal(data)
	if err != nil {
		log.Error("errorResp:", err)
	}

	sign = fmt.Sprintf("%x", md5.Sum(append(bts, []byte(cfg.App.SigningKey)...)))

	ctx.JSON(proto.Resp{
		Code: code,
		Msg:  msgStr,
		Data: string(bts),
		Sign: sign,
	})
}

func (ctx *myctx) respSucc(msg string, v interface{}) {
	if msg == "" {
		msg = "success"
	}

	bts, err := json.Marshal(v)
	if err != nil {
		log.Error("errorResp:", err)
	}

	sign := fmt.Sprintf("%x", md5.Sum(append(bts, []byte(cfg.App.SigningKey)...)))

	ctx.JSON(proto.Resp{
		Code: proto.SuccessRespCode,
		Msg:  msg,
		Data: string(bts),
		Sign: sign,
	})
}

func (ctx *myctx) getGroupIDFromToken() (uint, error) {
	cla, err := ctx.getClaimsFromToken()
	if err != nil {
		return 0, err
	}
	return cla.GroupID, nil
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

func (ctx *myctx) getGroupNodes() ([]models.Node, error) {
	var nodes []models.Node
	gid, err := ctx.getGroupIDFromToken()

	if err != nil {
		return nil, err
	}

	err = models.DB().Find(&nodes, "group_id=?", gid).Error
	return nodes, err
}

func (ctx *myctx) getGroupAddr() ([]string, error) {
	var addrs []string
	nodes, err := ctx.getGroupNodes()
	if err != nil {
		return nil, err
	}

	for _, v := range nodes {
		addrs = append(addrs, v.Addr)
	}
	return addrs, nil

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
