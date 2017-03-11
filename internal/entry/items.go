// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"sort"
	"strings"
)

// 所有 Entry 实现的公用部分。
type items struct {
	// 请求方法及其对应的 Handler
	handlers map[string]http.Handler

	// 缓存的 OPTIONS 请求头的 allow 报头内容，每次更新 handlers 时更新。
	optionsAllow string

	// 固定 optionsAllow 不再修改，
	// 调用 SetAllow() 进行强制修改之后为 true。
	fixedOptionsAllow bool

	// 固定 handlers[http.MethodOptions] 不再修改，
	// 显示地调用 items.Add(http.MethodOptions,...) 进行赋值之后为 true。
	fixedOptionsHandler bool
}

func newItems() *items {
	ret := &items{
		handlers: make(map[string]http.Handler, 10),
	}

	// 添加默认的 OPTIONS 请求内容
	ret.handlers[http.MethodOptions] = http.HandlerFunc(ret.optionsServeHTTP)
	ret.optionsAllow = ret.getOptionsAllow()

	return ret
}

// 实现 Entry.Add() 接口方法。
func (i *items) Add(method string, h http.Handler) error {
	if method == http.MethodOptions { // 强制修改 OPTIONS 方法的处理方式
		if i.fixedOptionsHandler { // 被强制修改过，不能再受理。
			return ErrMethodExists
		}

		i.handlers[method] = h
		i.fixedOptionsHandler = true
		return nil
	}

	// 非 OPTIONS 请求
	if _, found := i.handlers[method]; found {
		return ErrMethodExists
	}
	i.handlers[method] = h

	// 重新生成 optionsAllow 字符串
	if !i.fixedOptionsAllow {
		i.optionsAllow = i.getOptionsAllow()
	}
	return nil
}

func (i *items) optionsServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", i.optionsAllow)
}

func (i *items) getOptionsAllow() string {
	methods := make([]string, 0, len(i.handlers))
	for method := range i.handlers {
		methods = append(methods, method)
	}

	sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
	return strings.Join(methods, ", ")
}

// 返回值表示，是否还有路由项
func (i *items) Remove(methods ...string) bool {
	for _, method := range methods {
		delete(i.handlers, method)
		if method == http.MethodOptions { // 不恢复方法，只恢复了 fixOptionsHandler
			i.fixedOptionsHandler = false
		}
	}

	// 删完了
	if len(i.handlers) == 0 {
		i.optionsAllow = ""
		return true
	}

	// 只有一个 OPTIONS 了，且未经外界强制修改，则将其也一并删除。
	if len(i.handlers) == 1 && i.handlers[http.MethodOptions] != nil {
		if !i.fixedOptionsAllow && !i.fixedOptionsHandler {
			delete(i.handlers, http.MethodOptions)
			i.optionsAllow = ""
			return true
		}

	}

	if !i.fixedOptionsAllow {
		i.optionsAllow = i.getOptionsAllow()
	}
	return false
}

// SetAllow 设置 Allow 报头的内容。
func (i *items) SetAllow(optionsAllow string) {
	i.optionsAllow = optionsAllow
	i.fixedOptionsAllow = true
}

func (i *items) Handler(method string) http.Handler {
	return i.handlers[method]
}
