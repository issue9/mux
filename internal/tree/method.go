// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/sliceutil"
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
		http.MethodOptions, // 必须在最后，后面的 addAny 需要此规则。
		http.MethodHead,
	}

	addAny = Methods[:len(Methods)-2] // 添加请求方法时，所采用的默认值。

	methodIndexMap map[string]int

	methodIndexes = map[int]methodIndexEntity{}

	DefaultNotFound = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	DefaultMethodNotAllowed = func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
)

type headResponse struct {
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

func (n *Node) buildMethods() {
	n.methodIndex = 0
	for method := range n.handlers {
		n.methodIndex += methodIndexMap[method]
	}

	if _, found := methodIndexes[n.methodIndex]; !found {
		methods := make([]string, 0, len(n.handlers))
		for method := range n.handlers {
			methods = append(methods, method)
		}
		sort.Strings(methods)
		entity := methodIndexEntity{methods: methods, options: strings.Join(methods, ", ")}
		methodIndexes[n.methodIndex] = entity
	}
}

func (n *Node) optionsServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Allow", n.Options())
}

// Options 获取当前支持的请求方法列表字符串
func (n *Node) Options() string { return methodIndexes[n.methodIndex].options }

// Methods 当前节点支持的请求方法
func (n *Node) Methods() []string { return methodIndexes[n.methodIndex].methods }

// 添加一个处理函数
func (n *Node) addMethods(disableHead bool, h http.Handler, methods ...string) error {
	for _, m := range methods {
		if m == http.MethodHead || m == http.MethodOptions {
			return fmt.Errorf("无法手动添加 OPTIONS/HEAD 请求方法")
		}

		if sliceutil.Index(Methods, func(i int) bool { return Methods[i] == m }) == -1 {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := n.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}
		n.handlers[m] = h

		if m == http.MethodGet && !disableHead { // 如果是 GET，则顺便添加 HEAD
			n.handlers[http.MethodHead] = n.headServeHTTP(h)
		}
	}

	// 查看是否需要添加 OPTIONS
	if _, found := n.handlers[http.MethodOptions]; !found {
		n.handlers[http.MethodOptions] = http.HandlerFunc(n.optionsServeHTTP)
	}

	n.buildMethods()
	return nil
}

func (resp *headResponse) Write([]byte) (int, error) { return 0, nil }

func (n *Node) headServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&headResponse{ResponseWriter: w}, r)
	})
}
