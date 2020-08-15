package admin

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"net/http"
	"sync/atomic"

	"jiacrontab/pkg/version"

	"github.com/iwannay/log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
)

type myctx struct {
	iris.Context
	adm    *Admin
	claims CustomerClaims
}

func wrapCtx(ctx iris.Context, adm *Admin) *myctx {
	key := "__ctx__"
	if v := ctx.Values().Get(key); v != nil {
		return v.(*myctx)
	}

	c := &myctx{
		Context: ctx,
		adm:     adm,
	}
	if atomic.LoadInt32(&adm.initAdminUser) == 1 {
		ctx.SetCookieKV("ready", "true", func(ctx iris.Context, c *http.Cookie, op uint8) {
			if op == 1 {
				c.HttpOnly = false
			}
		})
	} else {
		ctx.SetCookieKV("ready", "false", func(ctx iris.Context, c *http.Cookie, op uint8) {
			if op == 1 {
				c.HttpOnly = false
			}
		})
	}
	ctx.Values().Set(key, c)
	return c
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
		cfg    = ctx.adm.getOpts()
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
		Code:    code,
		Msg:     msgStr,
		Data:    string(bts),
		Sign:    sign,
		Version: version.String(cfg.App.AppName),
	})
}

func (ctx *myctx) respSucc(msg string, v interface{}) {

	cfg := ctx.adm.getOpts()
	if msg == "" {
		msg = "success"
	}

	bts, err := json.Marshal(v)
	if err != nil {
		log.Error("errorResp:", err)
	}

	sign := fmt.Sprintf("%x", md5.Sum(append(bts, []byte(cfg.App.SigningKey)...)))

	ctx.JSON(proto.Resp{
		Code:    proto.SuccessRespCode,
		Msg:     msg,
		Data:    string(bts),
		Sign:    sign,
		Version: version.String(cfg.App.AppName),
	})
}

func (ctx *myctx) isSuper() bool {
	return ctx.claims.GroupID == models.SuperGroup.ID
}

func (ctx *myctx) parseClaimsFromToken() error {
	var ok bool

	if (ctx.claims != CustomerClaims{}) {
		return nil
	}

	token, ok := ctx.Values().Get("jwt").(*jwt.Token)
	if !ok {
		return errors.New("claims is nil")
	}
	bts, err := json.Marshal(token.Claims)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bts, &ctx.claims)
	if err != nil {
		return fmt.Errorf("unmarshal claims error(%s)", err)
	}
	var user models.User
	if err := models.DB().Take(&user, "id=?", ctx.claims.UserID).Error; err != nil {
		return fmt.Errorf("validate user from db error(%s)", err)
	}
	if ctx.claims.Mail != user.Mail || ctx.claims.GroupID != user.GroupID || ctx.claims.Root != user.Root || ctx.claims.Version != user.Version {
		return fmt.Errorf("token validate error")
	}

	if ctx.claims.GroupID == models.SuperGroup.ID {
		ctx.claims.Root = true
	}

	return nil
}

func (ctx *myctx) getGroupNodes() ([]models.Node, error) {
	var nodes []models.Node
	err := models.DB().Find(&nodes, "group_id=?", ctx.claims.GroupID).Error
	return nodes, err
}

func (ctx *myctx) verifyNodePermission(addr string) bool {
	var node models.Node
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

func (ctx *myctx) pubEvent(targetName, desc string, source interface{}, v interface{}) {
	var content string

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
		Content:    content,
	}

	switch v := source.(type) {
	case models.EventSourceName:
		e.SourceName = string(v)
	case models.EventSourceUsername:
		e.SourceUsername = string(v)
	}

	e.Pub()
}
