// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/issue9/mux/v10/types"
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
		http.MethodTrace, // OPTIONS/HEAD/TRACE 必须在最后，AnyMethods 从此处分割。
		http.MethodHead,
		http.MethodOptions,
	}

	AnyMethods = Methods[:len(Methods)-3] // 添加请求方法时，所采用的默认值。

	methodIndexes  map[string]int            // 将请求方法转换为一个对应的唯一数值
	optionsIndexes = map[int]optionsEntity{} // 各类请求方法组合下对应的 OPTIONS 值
)

const methodNotAllowed = "" // 表示 405 的处理方法在各个节点上的名称

type optionsEntity struct {
	methods []string
	options string
}

func init() {
	methodIndexes = make(map[string]int, len(Methods))
	for i, m := range Methods {
		methodIndexes[m] = 1 << i
	}
}

func buildOptionsIndexes(index int) {
	if _, found := optionsIndexes[index]; found {
		return
	}

	methods := make([]string, 0, len(Methods))
	for method, i := range methodIndexes {
		if index&i == i {
			methods = append(methods, method)
		}
	}
	slices.Sort(methods)

	optionsIndexes[index] = optionsEntity{
		methods: methods,
		options: strings.Join(methods, ", "),
	}
}

func (n *node[T]) buildMethods() {
	n.optionsIndex = 0
	for method := range n.handlers {
		n.optionsIndex += methodIndexes[method]
	}
	if n.root.hasTrace {
		n.optionsIndex += methodIndexes[http.MethodTrace]
	}
	buildOptionsIndexes(n.optionsIndex)
}

func (n *node[T]) AllowHeader() string { return optionsIndexes[n.optionsIndex].options }

// Methods 当前节点支持的请求方法
func (n *node[T]) Methods() []string { return optionsIndexes[n.optionsIndex].methods }

// 添加一个处理函数
func (n *node[T]) addMethods(h T, pattern string, ms []types.Middleware[T], methods ...string) error {
	for _, m := range methods {
		if m == http.MethodOptions || m == http.MethodHead || (n.root.hasTrace && m == http.MethodTrace) {
			return fmt.Errorf("无法手动添加 OPTIONS/HEAD/TRACE 请求方法")
		}
		if _, found := methodIndexes[m]; !found {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := n.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}

		if m == http.MethodGet {
			n.handlers[http.MethodHead] = ApplyMiddleware(h, http.MethodHead, pattern, n.root.Name(), ms...)
		}

		n.handlers[m] = ApplyMiddleware(h, m, pattern, n.root.Name(), ms...)
	}

	// 查看是否需要添加 OPTIONS
	if _, found := n.handlers[http.MethodOptions]; !found {
		n.handlers[http.MethodOptions] = ApplyMiddleware(n.root.optionsBuilder(n), http.MethodOptions, pattern, n.root.Name(), ms...)
	}

	if _, found := n.handlers[methodNotAllowed]; !found {
		n.handlers[methodNotAllowed] = ApplyMiddleware(n.root.methodNotAllowedBuilder(n), "", pattern, n.root.Name(), ms...)
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
	tree.node.optionsIndex = methodIndexes[http.MethodOptions]
	if tree.hasTrace {
		tree.node.optionsIndex += methodIndexes[http.MethodTrace]
	}

	for m, num := range tree.methods {
		if num > 0 {
			tree.node.optionsIndex += methodIndexes[m]
		}
	}

	buildOptionsIndexes(tree.node.optionsIndex)
}
