package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	"github.com/iwannay/jiaweb"
)

type AuthMiddleware struct {
	jiaweb.BaseMiddleware
}

func (m *AuthMiddleware) Handle(ctx jiaweb.Context) error {
	var data = make(map[string]interface{})
	var passURL = map[string]bool{
		"/login": true,
		"/":      true,
	}
	clientFeature := ctx.RemoteIP() + "-" + ctx.Request().Header.Get("User-Agent")

	clientSign := fmt.Sprintf("%x", md5.Sum([]byte(clientFeature)))
	path := ctx.Request().Path()

	if strings.HasPrefix(path, "/jiaweb") || strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/viewImage") {
		m.Next(ctx)
		return nil
	}

	if ctx.VerifyToken(&data) {
		if sign, ok := data["clientSign"]; ok && sign == clientSign {
			m.Next(ctx)
			return nil
		}
	}

	if _, ok := passURL[path]; ok {
		m.Next(ctx)
		return nil
	}

	ctx.Redirect("/login", http.StatusFound)
	return nil

}
