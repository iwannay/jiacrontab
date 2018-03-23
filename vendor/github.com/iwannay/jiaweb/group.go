package jiaweb

import (
	"github.com/iwannay/jiaweb/logger"
)

type (
	Group interface {
		Use(m ...Middleware) Group
		Group(prefix string, m ...Middleware) Group
		DELETE(path string, h HttpHandle) RouteNode
		GET(path string, h HttpHandle) RouteNode
		GETPOST(path string, h HttpHandle)
		HEAD(path string, h HttpHandle) RouteNode
		OPTIONS(path string, h HttpHandle) RouteNode
		PATCH(path string, h HttpHandle) RouteNode
		POST(path string, h HttpHandle) RouteNode
		PUT(path string, h HttpHandle) RouteNode
		RegisterRoute(method, path string, h HttpHandle) RouteNode
	}

	xGroup struct {
		prefix      string
		middlewares []Middleware
		server      *HttpServer
	}
)

func NewGroup(prefix string, server *HttpServer) Group {
	g := &xGroup{prefix: prefix, server: server}
	logger.Logger().Debug("JiaWbe:Gropu NewGroup ["+prefix+"]", LogTarget_HttpServer)
	return g
}

func (g *xGroup) Group(prefix string, m ...Middleware) Group {
	return NewGroup(g.prefix+prefix, g.server).Use(g.middlewares...).Use(m...)
}

func (g *xGroup) Use(m ...Middleware) Group {
	g.middlewares = append(g.middlewares, m...)
	return g
}

func (g *xGroup) DELETE(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_DELETE, path, h)
}

func (g *xGroup) GET(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_GET, path, h)
}
func (g *xGroup) POST(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_POST, path, h)
}
func (g *xGroup) GETPOST(path string, h HttpHandle) {
	g.add(HTTPMethod_POST, path, h)
	g.add(HTTPMethod_GET, path, h)
}

func (g *xGroup) PUST(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_DELETE, path, h)
}
func (g *xGroup) PATCH(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_DELETE, path, h)
}
func (g *xGroup) HEAD(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_DELETE, path, h)
}
func (g *xGroup) OPTIONS(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_OPTIONS, path, h)
}

func (g *xGroup) PUT(path string, h HttpHandle) RouteNode {
	return g.add(HTTPMethod_OPTIONS, path, h)
}

func (g *xGroup) add(method, path string, handler HttpHandle) RouteNode {
	node := g.server.Route().RegisterRoute(method, g.prefix+path, handler).Use(g.middlewares...)
	return node
}

func (g *xGroup) RegisterRoute(method, path string, handler HttpHandle) RouteNode {
	return g.add(method, path, handler)
}
