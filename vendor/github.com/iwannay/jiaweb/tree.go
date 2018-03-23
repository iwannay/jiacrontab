package jiaweb

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iwannay/jiaweb/utils"
)

const (
	nodeTypeStatic = iota
	nodeTypeParam
)

const (
	pathDelimiter  = '/'
	paramStart     = '<'
	paramDelimiter = ':'
	paramEnd       = '>'
)

type (
	Node struct {
		Path       string
		NodeType   int
		hander     RouteHandle
		middleware []Middleware
		Children   []*Node
		Params     Params
		reg        *regexp.Regexp
	}

	Param struct {
		Value string
	}

	Params []Param
)

func NewTree() *Node {
	return &Node{
		Path:     "/",
		NodeType: nodeTypeStatic,
	}
}

// Use 添加中间件
func (n *Node) Use(m ...Middleware) *Node {
	if len(m) < 0 {
		return n
	}

	step := len(n.middleware) - 1

	for _, v := range m {
		if v != nil {
			if step >= 0 {
				n.middleware[step].SetNext(v)
			}
			n.middleware = append(n.middleware, v)
			step++
		}
	}
	return n

}

func (n *Node) insertChild(path string, handler RouteHandle, m ...Middleware) {

	// fullpath := path
	var paramsCheck uint8
	var paramStartPos int
	var newChild *Node

	if handler == nil {
		panic("[route] handler function cannot be nil")
	}

walk:
	for {
		i := 0

		maxLength := utils.IntMin(len(path), len(n.Path))
		paramsCheck = 0
		for i < maxLength && path[i] == n.Path[i] {
			if path[i] == paramStart {
				paramStartPos = i
				paramsCheck++
			}
			if path[i] == paramEnd {
				paramsCheck--
			}
			i++

		}

		if paramsCheck != 0 {
			i = paramStartPos
		}

		// 当比树中原有结点path短时替换该结点为子结点
		if i < len(n.Path) && i != 0 {
			child := *n
			child.Path = child.Path[i:]
			n.Path = path[:i]
			n.middleware = []Middleware{}
			n.hander = nil

			// 检查新父节点是否包含参数
			if has, reg, params := parseParam(n.Path); has {
				n.NodeType = nodeTypeParam
				n.Params = params
				n.reg = reg
			} else {
				n.NodeType = nodeTypeStatic
				n.Params = nil
				n.reg = nil
			}

			// 检查新生成的子节点是否包含参数
			if has, reg, params := parseParam(child.Path); has {
				child.NodeType = nodeTypeParam
				child.Params = params
				child.reg = reg
			} else {
				child.NodeType = nodeTypeStatic
				child.Params = nil
				child.reg = nil
			}

			// 检查插入的path新结点是否包含参数
			if has, reg, params := parseParam(path[i:]); has {

				newChild = &Node{
					Path:     path[i:],
					NodeType: nodeTypeParam,
					hander:   handler,
					Params:   params,
					reg:      reg,
				}
				newChild.Use(m...)
				n.Children = []*Node{&child, newChild}
				return
			}

			newChild = &Node{
				Path:     path[i:],
				NodeType: nodeTypeStatic,
				hander:   handler,
			}
			newChild.Use(m...)
			n.Children = []*Node{&child, newChild}

		} else {

			path = path[i:]

			if len(n.Children) > 0 && path != "" {
				for _, v := range n.Children {
					if v.Path == path {
						n.hander = handler
						// panic("[route] " + fullpath + " already exists")
						return
					}

					// 深度遍历
					if v.Path[0] == path[0] {
						maxLength = utils.IntMin(len(path), len(v.Path))
						paramsCheck = 0
						i = 0
						for i < maxLength && path[i] == v.Path[i] {
							if path[i] == paramStart {
								paramsCheck++
							}
							if path[i] == paramEnd {
								paramsCheck--
							}
							i++

						}
						if paramsCheck == 0 || v.Path[0] != paramStart {
							n = v
							continue walk
						}

					}

				}

			}

			if path == "" {
				n.hander = handler
				n.Use(m...)
				// panic("[route] " + fullpath + " already exists")
				return
			}

			if has, reg, params := parseParam(path); has {
				newChild = &Node{
					Path:     path,
					reg:      reg,
					NodeType: nodeTypeParam,
					hander:   handler,
					Params:   params,
				}
				newChild.Use(m...)
				n.Children = append(n.Children, newChild)
				return
			}

			newChild = &Node{
				Path:     path,
				NodeType: nodeTypeStatic,
				hander:   handler,
			}
			newChild.Use(m...)
			n.Children = append(n.Children, newChild)

		}
		return
	}

}

func (n *Node) Middlewares() []Middleware {
	return n.middleware
}

func (n *Node) Node() *Node {
	return n
}

func (n *Node) getNode(path string) (node *Node, paramsValue map[string]string, fullPath string) {
	var i, maxLength int
	paramsValue = make(map[string]string)

	if n.NodeType == nodeTypeStatic {

		if path == n.Path {
			fullPath += n.Path
			return n, paramsValue, fullPath
		}
		if strings.HasPrefix(path, n.Path) && len(n.Children) > 0 {

			i = 0
			maxLength = utils.IntMin(len(path), len(n.Path))
			for i < maxLength && path[i] == n.Path[i] {
				i++
			}
			fullPath += n.Path[:i]
			path = path[i:]

			for _, v := range n.Children {
				if v.NodeType == nodeTypeParam {
					if matchs := v.reg.FindStringSubmatch(path); len(matchs) > 0 {
						if n, p, f := v.getNode(path); n != nil {
							for key, val := range p {
								paramsValue[key] = val
							}
							fullPath += f
							return n, paramsValue, fullPath
						}

					}
				}

				if strings.HasPrefix(path, v.Path) {
					if n, p, f := v.getNode(path); n != nil {
						for key, val := range p {
							paramsValue[key] = val
						}
						fullPath += f
						return n, paramsValue, fullPath
					}
				}
			}

			return nil, paramsValue, fullPath

		}
	} else {
		matchs := n.reg.FindStringSubmatch(path)
		if len(matchs) > 0 {
			paternPath := matchs[0]

			for k, v := range n.Params {
				paramsValue[v.Value] = matchs[k+1]
			}

			if paternPath == path {
				fullPath += n.Path
				return n, paramsValue, fullPath
			}

			if len(n.Children) > 0 {

				i = 0
				maxLength = utils.IntMin(len(paternPath), len(path))
				for i < maxLength && paternPath[i] == path[i] {
					i++
				}
				fullPath += n.Path
				path = path[i:]

				for _, v := range n.Children {
					if v.NodeType == nodeTypeParam {
						if matchs := v.reg.FindStringSubmatch(path); len(matchs) > 0 {

							if n, p, f := v.getNode(path); n != nil {
								for key, val := range p {
									paramsValue[key] = val
								}
								fullPath += f
								return n, paramsValue, fullPath
							}

						}
					}

					if strings.HasPrefix(path, v.Path) {

						if n, p, f := v.getNode(path); n != nil {
							for key, val := range p {
								paramsValue[key] = val
							}
							fullPath += f
							return n, paramsValue, fullPath
						}
					}
				}

				return nil, paramsValue, fullPath

			}

		}
	}

	return nil, paramsValue, fullPath

}

func (n *Node) GetValue(path string) (node *Node, handle RouteHandle, paramsValue map[string]string) {
	node, params, _ := n.getNode(path)
	if node != nil {
		return node, node.hander, params
	}
	return nil, nil, nil
}

func (n *Node) Middleware() []Middleware {
	return n.middleware
}

func parseParam(path string) (bool, *regexp.Regexp, Params) {
	l := len(path)
	var params Params
	var patern string
	var paramsSli []string
	var hasParam bool
	var reg *regexp.Regexp

walk:
	for i := 0; i < l; i++ {

		if path[i] == paramStart {
			startParam := i
			for i < l {

				if path[i] == paramEnd {
					hasParam = true
					paramsSli = strings.Split(string(path[startParam+1:i]), ":")

					if len(paramsSli) != 2 {
						panic("route: subpath " + path + " is invalid!")
					}
					if patern == "" {
						patern = fmt.Sprintf("^(%s)", paramsSli[1])
					} else {
						if len(params) == 0 {
							patern = fmt.Sprintf("^%s(%s)", patern, paramsSli[1])
						} else {
							patern = fmt.Sprintf("%s(%s)", patern, paramsSli[1])
						}

					}

					params = append(params, Param{paramsSli[0]})
					continue walk
				}

				i++

			}
		}
		patern = patern + string(path[i])
	}
	if hasParam {
		reg = regexp.MustCompile(patern)
	}

	return hasParam, reg, params
}
