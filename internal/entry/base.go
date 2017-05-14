// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

type optionsState int8

const (
	optionsStateDefault      optionsState = iota // 默认情况
	optionsStateFixedString                      // 设置为固定的字符串
	optionsStateFixedHandler                     // 设置为固定的 http.Handler
	optionsStateDisable                          // 禁用
)

// 所有 Entry 实现的公用部分。
type base struct {
	pattern      string                  // 匹配模式字符串
	wildcard     bool                    // 是否包含通配符
	handlers     map[string]http.Handler // 请求方法及其对应的 Handler
	optionsAllow string                  // 缓存的 OPTIONS 请求头的 allow 报头内容。
	optionsState optionsState            // OPTIONS 报头的处理方式
}

func newBase(s *syntax.Syntax) *base {
	ret := &base{
		pattern:      s.Pattern,
		handlers:     make(map[string]http.Handler, len(method.Supported)),
		wildcard:     s.Wildcard,
		optionsState: optionsStateDefault,
	}

	// 添加默认的 OPTIONS 请求内容
	ret.handlers[http.MethodOptions] = http.HandlerFunc(ret.optionsServeHTTP)
	ret.optionsAllow = ret.getOptionsAllow()

	return ret
}

// Entry.Pattern()
func (b *base) Pattern() string {
	return b.pattern
}

// Entry.Add()
func (b *base) Add(h http.Handler, methods ...string) error {
	for _, m := range methods {
		if !method.IsSupported(m) {
			return fmt.Errorf("不支持的请求方法 %v", m)
		}

		if err := b.addSingle(h, m); err != nil {
			return err
		}
	}

	return nil
}

func (b *base) addSingle(h http.Handler, method string) error {
	if method == http.MethodOptions { // 强制修改 OPTIONS 方法的处理方式
		if b.optionsState == optionsStateFixedHandler { // 被强制修改过，不能再受理。
			return errors.New("该请求方法 OPTIONS 已经存在") // 与以下的错误提示相同
		}

		b.handlers[http.MethodOptions] = h
		b.optionsState = optionsStateFixedHandler
		return nil
	}

	// 非 OPTIONS 请求
	if _, found := b.handlers[method]; found {
		return fmt.Errorf("该请求方法 %v 已经存在", method)
	}
	b.handlers[method] = h

	// 重新生成 optionsAllow 字符串
	if b.optionsState == optionsStateDefault {
		b.optionsAllow = b.getOptionsAllow()
	}
	return nil
}

func (b *base) optionsServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", b.optionsAllow)
}

func (b *base) getOptionsAllow() string {
	methods := make([]string, 0, len(b.handlers))
	for method := range b.handlers {
		methods = append(methods, method)
	}

	sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
	return strings.Join(methods, ", ")
}

// Entry.Remove()
func (b *base) Remove(methods ...string) bool {
	for _, method := range methods {
		delete(b.handlers, method)
		if method == http.MethodOptions { // 明确指出要删除该路由项的 OPTIONS 时，表示禁止
			b.optionsState = optionsStateDisable
		}
	}

	// 删完了
	if len(b.handlers) == 0 {
		b.optionsAllow = ""
		return true
	}

	// 只有一个 OPTIONS 了，且未经外界强制修改，则将其也一并删除。
	if len(b.handlers) == 1 && b.handlers[http.MethodOptions] != nil {
		if b.optionsState == optionsStateDefault {
			delete(b.handlers, http.MethodOptions)
			b.optionsAllow = ""
			return true
		}
	}

	if b.optionsState == optionsStateDefault {
		b.optionsAllow = b.getOptionsAllow()
	}
	return false
}

// Entry.SetAllow()
func (b *base) SetAllow(optionsAllow string) {
	b.optionsAllow = optionsAllow
	b.optionsState = optionsStateFixedString
}

// Entry.Handler()
func (b *base) Handler(method string) http.Handler {
	return b.handlers[method]
}
