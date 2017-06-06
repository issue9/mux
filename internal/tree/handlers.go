// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

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

type handlers struct {
	handlers     map[string]http.Handler // 请求方法及其对应的 Handler
	optionsAllow string                  // 缓存的 OPTIONS 请求头的 allow 报头内容。
	optionsState optionsState            // OPTIONS 报头的处理方式
}

func newHandlers() *handlers {
	ret := &handlers{
		handlers:     make(map[string]http.Handler, len(method.Supported)),
		optionsState: optionsStateDefault,
	}

	// 添加默认的 OPTIONS 请求内容
	ret.handlers[http.MethodOptions] = http.HandlerFunc(ret.optionsServeHTTP)
	ret.optionsAllow = ret.getOptionsAllow()

	return ret
}

func (hs *handlers) add(h http.Handler, methods ...string) error {
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

func (hs *handlers) addSingle(h http.Handler, method string) error {
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

func (hs *handlers) optionsServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", hs.optionsAllow)
}

func (hs *handlers) getOptionsAllow() string {
	methods := make([]string, 0, len(hs.handlers))
	for method := range hs.handlers {
		methods = append(methods, method)
	}

	sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
	return strings.Join(methods, ", ")
}

func (hs *handlers) remove(methods ...string) bool {
	for _, method := range methods {
		delete(hs.handlers, method)
		if method == http.MethodOptions { // 明确指出要删除该路由项的 OPTIONS 时，表示禁止
			hs.optionsState = optionsStateDisable
		}
	}

	// 删完了
	if len(hs.handlers) == 0 {
		hs.optionsAllow = ""
		return true
	}

	// 只有一个 OPTIONS 了，且未经外界强制修改，则将其也一并删除。
	if len(hs.handlers) == 1 && hs.handlers[http.MethodOptions] != nil {
		if hs.optionsState == optionsStateDefault {
			delete(hs.handlers, http.MethodOptions)
			hs.optionsAllow = ""
			return true
		}
	}

	if hs.optionsState == optionsStateDefault {
		hs.optionsAllow = hs.getOptionsAllow()
	}
	return false
}

func (hs *handlers) setAllow(optionsAllow string) {
	hs.optionsAllow = optionsAllow
	hs.optionsState = optionsStateFixedString
}

func (hs *handlers) handler(method string) http.Handler {
	return hs.handlers[method]
}
