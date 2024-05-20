// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/issue9/mux/v8/types"
)

var (
	// Methods 支持的请求方法
	Methods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodConnect,
		http.MethodTrace, // OPTIONS/HEAD/TRACE 必须在最后，后面的 addAny 需要此规则。
		http.MethodHead,
		http.MethodOptions,
	}

	addAny = Methods[:len(Methods)-3] // 添加请求方法时，所采用的默认值。

	methodIndexMap map[string]int // 各个请求方法对应的数值

	methodIndexes = map[int]methodIndexEntity{}
)

const methodNotAllowed = "" // 表示 405 的处理方法在各个节点上的名称。

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
	slices.Sort(methods)

	methodIndexes[index] = methodIndexEntity{
		methods: methods,
		options: strings.Join(methods, ", "),
	}
}

func (n *node[T]) buildMethods() {
	n.methodIndex = 0
	for method := range n.handlers {
		n.methodIndex += methodIndexMap[method]
	}
	if n.root.trace {
		n.methodIndex += methodIndexMap[http.MethodTrace]
	}
	buildMethodIndexes(n.methodIndex)
}

func (n *node[T]) AllowHeader() string { return methodIndexes[n.methodIndex].options }

// Methods 当前节点支持的请求方法
func (n *node[T]) Methods() []string { return methodIndexes[n.methodIndex].methods }

// 添加一个处理函数
func (n *node[T]) addMethods(h T, pattern string, ms []types.MiddlewareOf[T], methods ...string) error {
	for _, m := range methods {
		if m == http.MethodOptions || m == http.MethodHead || (n.root.trace && m == http.MethodTrace) {
			return fmt.Errorf("无法手动添加 OPTIONS/HEAD/TRACE 请求方法")
		}
		if _, found := methodIndexMap[m]; !found {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := n.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}

		if m == http.MethodGet {
			n.handlers[http.MethodHead] = ApplyMiddleware(h, http.MethodHead, pattern, ms...)
		}

		n.handlers[m] = ApplyMiddleware(h, m, pattern, ms...)
	}

	// 查看是否需要添加 OPTIONS
	if _, found := n.handlers[http.MethodOptions]; !found {
		n.handlers[http.MethodOptions] = ApplyMiddleware(n.root.optionsBuilder(n), http.MethodOptions, pattern, ms...)
	}

	if _, found := n.handlers[methodNotAllowed]; !found {
		n.handlers[methodNotAllowed] = ApplyMiddleware(n.root.methodNotAllowedBuilder(n), "", pattern, ms...)
	}

	n.buildMethods()
	n.root.buildMethods(1, methods...)

	return nil
}

// num 表示为该请求方法加上的计数
func (tree *Tree[T]) buildMethods(num int, methods ...string) {
	for _, m := range methods {
		tree.methods[m] += num
	}

	// 即使所有接口都没了，也有 OPTIONS * 存在，所以始终有 OPTIONS 和可能的 TRACE 存在。
	tree.node.methodIndex = methodIndexMap[http.MethodOptions]
	if tree.trace {
		tree.node.methodIndex += methodIndexMap[http.MethodTrace]
	}

	for m, num := range tree.methods {
		if num > 0 {
			tree.node.methodIndex += methodIndexMap[m]
		}
	}

	buildMethodIndexes(tree.node.methodIndex)
}
