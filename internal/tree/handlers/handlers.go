// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package handlers 用于处理节点下请求方法与处理函数的对应关系
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/mux/internal/method"
)

type optionsState int8

const (
	optionsStateDefault      optionsState = iota // 默认情况
	optionsStateFixedString                      // 设置为固定的字符串
	optionsStateFixedHandler                     // 设置为固定的 http.Handler
	optionsStateDisable                          // 禁用
)

// Handlers 用于表示某节点下各个请求方法对应的处理函数。
type Handlers struct {
	handlers     map[string]http.Handler // 请求方法及其对应的 http.Handler
	optionsAllow string                  // 缓存的 OPTIONS 请求头的 allow 报头内容。
	optionsState optionsState            // OPTIONS 报头的处理方式
}

// New 声明一个新的 Handlers 实例
func New() *Handlers {
	ret := &Handlers{
		handlers:     make(map[string]http.Handler, len(method.Supported)),
		optionsState: optionsStateDefault,
	}

	// 添加默认的 OPTIONS 请求内容
	ret.handlers[http.MethodOptions] = http.HandlerFunc(ret.optionsServeHTTP)
	ret.optionsAllow = ret.getOptionsAllow()

	return ret
}

// Add 添加一个处理函数
func (hs *Handlers) Add(h http.Handler, methods ...string) error {
	for _, m := range methods {
		if !method.IsSupported(m) {
			return fmt.Errorf("不支持的请求方法 %v", m)
		}

		if err := hs.addSingle(h, m); err != nil {
			return err
		}
	}

	return nil
}

func (hs *Handlers) addSingle(h http.Handler, method string) error {
	if method == http.MethodOptions { // 强制修改 OPTIONS 方法的处理方式
		if hs.optionsState == optionsStateFixedHandler { // 被强制修改过，不能再受理。
			return errors.New("该请求方法 OPTIONS 已经存在")
		}

		hs.handlers[http.MethodOptions] = h
		hs.optionsState = optionsStateFixedHandler
		return nil
	}

	// 非 OPTIONS 请求
	if _, found := hs.handlers[method]; found {
		return fmt.Errorf("该请求方法 %v 已经存在", method)
	}
	hs.handlers[method] = h

	// 重新生成 optionsAllow 字符串
	if hs.optionsState == optionsStateDefault {
		hs.optionsAllow = hs.getOptionsAllow()
	}
	return nil
}

func (hs *Handlers) optionsServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", hs.optionsAllow)
}

func (hs *Handlers) getOptionsAllow() string {
	methods := make([]string, 0, len(hs.handlers))
	for method := range hs.handlers {
		methods = append(methods, method)
	}

	sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
	return strings.Join(methods, ", ")
}

// Remove 移除某个请求方法对应的处理函数
func (hs *Handlers) Remove(methods ...string) bool {
	for _, method := range methods {
		delete(hs.handlers, method)
		if method == http.MethodOptions { // 明确指出要删除该路由项的 OPTIONS 时，表示禁止
			hs.optionsState = optionsStateDisable
		}
	}

	// 删完了
	if hs.Len() == 0 {
		hs.optionsAllow = ""
		return true
	}

	// 只有一个 OPTIONS 了，且未经外界强制修改，则将其也一并删除。
	if hs.Len() == 1 &&
		hs.handlers[http.MethodOptions] != nil &&
		hs.optionsState == optionsStateDefault {
		delete(hs.handlers, http.MethodOptions)
		hs.optionsAllow = ""
		return true
	}

	if hs.optionsState == optionsStateDefault {
		hs.optionsAllow = hs.getOptionsAllow()
	}
	return false
}

// SetAllow 设置 Options 请求头的 Allow 报头。
func (hs *Handlers) SetAllow(optionsAllow string) {
	hs.optionsAllow = optionsAllow
	hs.optionsState = optionsStateFixedString
}

// Handler 获取指定方法对应的处理函数
func (hs *Handlers) Handler(method string) http.Handler {
	return hs.handlers[method]
}

// Len 获取当前支持请求方法数量
func (hs *Handlers) Len() int {
	return len(hs.handlers)
}
