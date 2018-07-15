package conf

var JwtService *Jwt

type Jwt struct {
	JWTSigningKey     string
	TokenCookieName   string
	TokenExpires      int64
	TokenCookieMaxAge int64
}

func LoadJwtService() {
	jwt := cf.Section("jwt")
	JwtService = &Jwt{
		JWTSigningKey:     jwt.Key("signing_key").MustString("eyJhbGciOiJIUzI1"),
		TokenCookieName:   jwt.Key("token_cookie_name").MustString("access_token"),
		TokenExpires:      jwt.Key("expires").MustInt64(3000),
		TokenCookieMaxAge: jwt.Key("token_cookie_maxage").MustInt64(3000),
	}
}
