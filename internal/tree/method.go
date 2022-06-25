// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"net/http"
	"sort"
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
		http.MethodHead,
		http.MethodOptions, // OPTIONS 必须在最后，后面的 addAny 需要此规则。
	}

	addAny = Methods[:len(Methods)-1] // 添加请求方法时，所采用的默认值。

	methodIndexMap map[string]int // 各个请求方法对应的数值

	methodIndexes = map[int]methodIndexEntity{}
)

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

	methodIndexes[index] = methodIndexEntity{
		methods: methods,
		options: strings.Join(methods, ", "),
	}
}

func (n *Node[T]) buildMethods() {
	n.methodIndex = 0
	for method := range n.handlers {
		n.methodIndex += methodIndexMap[method]
	}
	buildMethodIndexes(n.methodIndex)
}

// Options 获取当前支持的请求方法列表字符串
func (n *Node[T]) Options() string { return methodIndexes[n.methodIndex].options }

// Methods 当前节点支持的请求方法
func (n *Node[T]) Methods() []string { return methodIndexes[n.methodIndex].methods }

// 添加一个处理函数
func (n *Node[T]) addMethods(h T, methods ...string) error {
	for _, m := range methods {
		if m == http.MethodOptions {
			return fmt.Errorf("无法手动添加 OPTIONS 请求方法")
		}
		if _, found := methodIndexMap[m]; !found {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := n.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}
		n.handlers[m] = h
	}

	// 查看是否需要添加 OPTIONS
	if _, found := n.handlers[http.MethodOptions]; !found {
		n.handlers[http.MethodOptions] = n.root.optionsBuilder(n)
	}

	n.buildMethods()
	n.root.buildMethods(1, methods...)

	return nil
}

// num 表示为该请求方法加上的计数
func (tree *Tree[T]) buildMethods(num int, methods ...string) {
	if len(methods) == 0 {
		methods = addAny
	}

	for _, m := range methods {
		tree.methods[m] += num
	}

	// 即使所有接口都没了，也有 OPTIONS * 存在，所以始终有 OPTIONS。
	tree.node.methodIndex = methodIndexMap[http.MethodOptions]
	for m, num := range tree.methods {
		if num > 0 {
			tree.node.methodIndex += methodIndexMap[m]
		}
	}

	buildMethodIndexes(tree.node.methodIndex)
}
