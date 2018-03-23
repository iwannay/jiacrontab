package jiaweb

type HttpModule struct {
	Name           string
	OnBeginRequest func(Context)
	OnEndRequest   func(Context)
}

func ignoreFaviconModule() *HttpModule {
	return &HttpModule{
		Name: "IgnoreFAvicon",
		OnBeginRequest: func(ctx Context) {
			if ctx.Request().URL.Path == "/favicon.ico" {
				ctx.End()
			}
		},
	}
}
