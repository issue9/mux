// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"errors"
	"net/http"
	"strings"
)

var ErrMethodExists = errors.New("该请求方法的路由已经存在")

type items struct {
	handlers    map[string]http.Handler
	allows      string // 缓存的 allow 报头内容
	fixedAllows bool   // 固定 allows 不再修改
}

func newItems() *items {
	return &items{
		handlers: make(map[string]http.Handler, 10),
	}
}

// NOTE: 如果 mehtod == options 则，其它两个设置都不起作用
func (i *items) Add(method string, h http.Handler) error {
	if _, found := i.handlers[method]; found {
		return ErrMethodExists
	}

	i.handlers[method] = h

	// 添加 OPTIONS 的处理函数
	if _, found := i.handlers[http.MethodOptions]; !found {
		i.handlers[http.MethodOptions] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Allow", i.allows)
		})
	}

	if !i.fixedAllows {
		i.allows = i.getAllows()
	}
	return nil
}

func (i *items) getAllows() string {
	methods := make([]string, 0, len(i.handlers))
	for method := range i.handlers {
		methods = append(methods, method)
	}

	return strings.Join(methods, ", ")
}

// 返回值表示，是否还有路由项
func (i *items) Remove(methods ...string) bool {
	for _, method := range methods {
		delete(i.handlers, method)
	}

	if len(i.handlers) == 0 || // 没了
		(len(i.handlers) == 1 && i.handlers[http.MethodOptions] != nil) { // 只有一个 OPTIONS 了
		delete(i.handlers, http.MethodOptions)

		if !i.fixedAllows {
			i.allows = ""
		}

		return true
	}

	if !i.fixedAllows {
		i.allows = i.getAllows()
	}
	return false
}

func (i *items) SetAllow(allows string) {
	i.allows = allows
	i.fixedAllows = true
}

func (i *items) Handler(method string) http.Handler {
	return i.handlers[method]
}
