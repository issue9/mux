// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

var (
	// Methods 所有支持请求方法
	Methods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodConnect,
		http.MethodTrace,
		http.MethodOptions, // OPTIONS 和 HEAD 必须在最后，后面的 addAny 需要此规则。
		http.MethodHead,
	}

	addAny = Methods[:len(Methods)-2] // 添加请求方法时，所采用的默认值。

	methodIndexMap map[string]int // Methods 的反向对照表

	methodIndexes = map[int]methodIndexEntity{}
)

type headResponse struct {
	size int
	http.ResponseWriter
}

func init() {
	methodIndexMap = make(map[string]int, len(Methods))
	for i, m := range Methods {
		methodIndexMap[m] = 1 << i
	}
}

type methodIndexEntity struct {
	methods []string
	options string
}

func buildMethodIndexes(index int) {
	if _, found := methodIndexes[index]; found {
		return
	}

	methods := make([]string, 0, len(Methods))
	for method, i := range methodIndexMap {
		if index&i == i {
			methods = append(methods, method)
		}
	}
	sort.Strings(methods)

	entity := methodIndexEntity{
		methods: methods,
		options: strings.Join(methods, ", "),
	}
	methodIndexes[index] = entity
}

func (n *Node) buildMethods() {
	n.methodIndex = 0
	for method := range n.handlers {
		n.methodIndex += methodIndexMap[method]
	}
	buildMethodIndexes(n.methodIndex)
}

func (n *Node) optionsServeHTTP(w http.ResponseWriter, _ *http.Request) {
	optionsHandle(w, n.Options())
}

func optionsHandle(w http.ResponseWriter, opt string) { w.Header().Set("Allow", opt) }

// Options 获取当前支持的请求方法列表字符串
func (n *Node) Options() string { return methodIndexes[n.methodIndex].options }

// Methods 当前节点支持的请求方法
func (n *Node) Methods() []string { return methodIndexes[n.methodIndex].methods }

// 添加一个处理函数
func (n *Node) addMethods(h http.Handler, methods ...string) error {
	for _, m := range methods {
		if m == http.MethodHead || m == http.MethodOptions {
			return fmt.Errorf("无法手动添加 OPTIONS/HEAD 请求方法")
		}

		if _, found := methodIndexMap[m]; !found {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := n.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}
		n.handlers[m] = h

		if m == http.MethodGet { // 如果是 GET，则顺便添加 HEAD
			n.handlers[http.MethodHead] = n.headServeHTTP(h)
		}
	}

	// 查看是否需要添加 OPTIONS
	if _, found := n.handlers[http.MethodOptions]; !found {
		n.handlers[http.MethodOptions] = http.HandlerFunc(n.optionsServeHTTP)
	}

	n.buildMethods()
	n.root.buildMethods(1, methods...)

	return nil
}

func (tree *Tree) buildMethods(v int, methods ...string) {
	if len(methods) == 0 {
		methods = addAny
	}

	for _, m := range methods {
		tree.methods[m] += v
	}

	tree.node.methodIndex = methodIndexMap[http.MethodOptions]
	for m, num := range tree.methods {
		if num > 0 {
			tree.node.methodIndex += methodIndexMap[m]
			if m == http.MethodGet {
				tree.node.methodIndex += methodIndexMap[http.MethodHead]
			}
		}
	}

	buildMethodIndexes(tree.node.methodIndex)
}

func (resp *headResponse) Write(bs []byte) (int, error) {
	l := len(bs)
	resp.size += l

	resp.Header().Set("Content-Length", strconv.Itoa(resp.size))
	return l, nil
}

func (n *Node) headServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&headResponse{ResponseWriter: w}, r)
	})
}
