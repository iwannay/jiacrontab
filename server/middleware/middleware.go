package middleware
//
//import (
//	"crypto/md5"
//	"fmt"
//	"net/http"
//	"strings"
//
//	"github.com/iwannay/jiaweb"
//)
//
//
//
//func  Auth(ctx iris.Context) error {
//	var data = make(map[string]interface{})
//	var passURL = map[string]bool{
//		"/login": true,
//	}
//	clientFeature := ctx.RemoteIP() + "-" + ctx.Request().Header.Get("User-Agent")
//
//	clientSign := fmt.Sprintf("%x", md5.Sum([]byte(clientFeature)))
//	path := ctx.Request().Path()
//
//	if strings.HasPrefix(path, "/jiaweb") || strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/viewImage") {
//		return m.Next(ctx)
//	}
//
//	if ctx.VerifyToken(&data) {
//		if sign, ok := data["clientSign"]; ok && sign == clientSign {
//			ctx.HttpServer().Render.AddLocals(jiaweb.KValue{
//				Key:   "user",
//				Value: data,
//			})
//			return m.Next(ctx)
//		}
//	}
//
//	if _, ok := passURL[path]; ok {
//		return m.Next(ctx)
//
//	}
//
//	ctx.Redirect("/login", http.StatusFound)
//	return nil
//
//}
