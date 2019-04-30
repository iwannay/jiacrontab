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
	claims CustomerClaims
}

func wrapCtx(ctx iris.Context) *myctx {
	return &myctx{
		Context: ctx,
	}
}

func (ctx *myctx) respNotAllowed() {
	ctx.respError(proto.Code_NotAllowed, proto.Msg_NotAllowed)
}

func (ctx *myctx) respAuthFailed(err error) {
	ctx.respError(proto.Code_FailedAuth, err)
}

func (ctx *myctx) respDBError(err error) {
	ctx.respError(proto.Code_DBError, err)
}

func (ctx *myctx) respJWTError(err error) {
	ctx.respError(proto.Code_JWTError, err)
}

func (ctx *myctx) respBasicError(err error) {
	ctx.respError(proto.Code_Error, err)
}

func (ctx *myctx) respParamError(err error) {
	ctx.respError(proto.Code_ParamsError, err)
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
	err := ctx.parseClaimsFromToken()
	if err != nil {
		return 0, err
	}
	return ctx.claims.GroupID, nil
}

func (ctx *myctx) isSuper() bool {
	ok, err := ctx.getGroupIDFromToken()
	return ok == models.SuperGroup.ID && err == nil
}

func (ctx *myctx) parseClaimsFromToken() error {

	if (ctx.claims != CustomerClaims{}) {
		return nil
	}

	var ok bool
	token, ok := ctx.Values().Get("jwt").(*jwt.Token)
	if !ok {
		return errors.New("claims is nil")
	}
	bts, err := json.Marshal(token.Claims)
	if err != nil {
		return err
	}
	json.Unmarshal(bts, &ctx.claims)
	return err
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

func (ctx *myctx) verifyNodePermission(addr string) bool {
	var node models.Node
	if err := ctx.parseClaimsFromToken(); err != nil {
		return false
	}
	return node.VerifyUserGroup(ctx.claims.UserID, ctx.claims.GroupID, addr)
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

func (ctx *myctx) Valid(i Parameter) error {
	if err := ctx.ReadJSON(i); err != nil {
		return err
	}

	if err := i.Verify(ctx); err != nil {
		return err
	}

	if err := validStructRule(i); err != nil {
		return err
	}
	return nil
}

func (ctx *myctx) pubEvent(targetName, desc, sourceName string, v interface{}) {
	var content string
	if (ctx.claims == CustomerClaims{}) {
		err := ctx.parseClaimsFromToken()
		if err != nil {
			return
		}
	}

	if v != nil {
		bts, err := json.Marshal(v)
		if err != nil {
			return
		}
		content = string(bts)
	}

	e := models.Event{
		GroupID:    ctx.claims.GroupID,
		UserID:     ctx.claims.UserID,
		Username:   ctx.claims.Username,
		EventDesc:  desc,
		TargetName: targetName,
		SourceName: sourceName,
		Content:    content,
	}
	e.Pub()
}
