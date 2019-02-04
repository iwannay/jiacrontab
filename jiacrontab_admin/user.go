package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris"
)

type CustomerClaims struct {
	jwt.StandardClaims
	Username string
	Group    int
	Root     bool
}

func login(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody struct {
			Username string `json:"username"`
			Passwd   string `json:"passwd"`
			Remember bool   `json:"remember"`
		}
		user           models.User
		customerClaims CustomerClaims
	)

	if !user.Verify(reqBody.Username, reqBody.Passwd) {
		ctx.respError(proto.Code_FailedAuth, "帐号或密码不正确", nil)
		return
	}

	customerClaims.ExpiresAt = cfg.Jwt.Expires + time.Now().Unix()
	customerClaims.Username = reqBody.Username
	customerClaims.Group = user.Group
	customerClaims.Root = user.Root

	if reqBody.Remember {
		customerClaims.ExpiresAt = time.Now().Add(24 * 30 * time.Hour).Unix()
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &customerClaims).SignedString([]byte(cfg.Jwt.SigningKey))

	if err != nil {
		ctx.respError(proto.Code_FailedAuth, "无法生成访问凭证", nil)
		return
	}

	ctx.respSucc("", token)
}

func signUp(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		user    models.User
		reqBody userReqParams
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	user.Name = reqBody.Name
	user.Passwd = reqBody.Passwd
	user.Group = reqBody.Group
	user.Root = reqBody.Root

	if err = user.SignUp(); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", true)
}
