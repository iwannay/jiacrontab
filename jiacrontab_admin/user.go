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
	Role     int
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
		return ctx.respError(proto.Code_FailedAuth, "帐号或密码不正确", nil)
	}

	customerClaims.ExpiresAt = cfg.Jwt.Expires + time.Now().Unix()
	customerClaims.Username = reqBody.Username
	customerClaims.Role = user.Role
	customerClaims.Root = user.Root

	if reqBody.Remember {
		customerClaims.ExpiresAt = time.Now().Add(24 * 30 * time.Hour).Unix()
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &customerClaims).SignedString([]byte(cfg.Jwt.SigningKey))

	if err != nil {
		return ctx.respError(proto.Code_FailedAuth, "无法生成访问凭证", nil)
	}

	return ctx.respSucc("", map[string]string{
		"token": token,
	})

}
