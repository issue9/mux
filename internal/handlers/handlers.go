// SPDX-License-Identifier: MIT

// Package handlers 用于处理节点下与处理函数相关的逻辑
package handlers

import (
	"fmt"
	"net/http"
	"sort"
)

type optionsState int8

// 表示对 OPTIONAL 请求中 Allow 报头中输出内容的处理方式。
const (
	optionsStateDefault      optionsState = iota // 默认情况
	optionsStateFixedString                      // 设置为固定的字符串
	optionsStateFixedHandler                     // 设置为固定的 http.Handler
	optionsStateDisable                          // 禁用，不会自动生 optionAllow 的值
)

type headState int8

// 表示对 HEAD 请求是否根据 GET 请求自动生成。
const (
	headStateDefault headState = iota // 不作任何额外操作
	headStateAuto                     // 自动生成，有 GET 就生成，作为默认值
	headStateFixed                    // 有固定的 Handler
)

// Methods 所有支持请求方法
var Methods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodDelete,
	http.MethodPut,
	http.MethodPatch,
	http.MethodConnect,
	http.MethodTrace,
	http.MethodOptions,
	http.MethodHead,
}

// 除 OPTIONS 和 HEAD 之外的所有支持的元素
//
// 在 Add 方法中用到。
var addAny = Methods[:len(Methods)-2]

// Handlers 用于表示某节点下各个请求方法对应的处理函数
type Handlers struct {
	handlers map[string]http.Handler // 请求方法及其对应的 http.Handler

	optionsAllow string       // 缓存的 OPTIONS 请求的 allow 报头内容。
	optionsState optionsState // OPTIONS 请求的处理方式
	headState    headState
}

// New 声明一个新的 Handlers 实例
//
// disableHead 是否自动添加 HEAD 请求内容。
func New(disableOptions, disableHead bool) *Handlers {
	ret := &Handlers{
		handlers:     make(map[string]http.Handler, 4), // 大部分不会超过 4 条数据
		optionsState: optionsStateDefault,
		headState:    headStateDefault,
	}

	if !disableHead {
		ret.headState = headStateAuto
	}

	if disableOptions {
		ret.optionsState = optionsStateDisable
	} else {
		ret.handlers[http.MethodOptions] = http.HandlerFunc(ret.optionsServeHTTP)
		ret.optionsAllow = ret.getOptionsAllow()
	}

	return ret
}

// Add 添加一个处理函数
func (hs *Handlers) Add(h http.Handler, methods ...string) error {
	if len(methods) == 0 {
		methods = addAny
	}

	for _, m := range methods {
		if err := hs.addSingle(h, m); err != nil {
			return err
		}
	}

	return nil
}

func (hs *Handlers) addSingle(h http.Handler, m string) error {
	switch m {
	case http.MethodOptions:
		if hs.optionsState == optionsStateFixedHandler { // 被强制修改过，不能再受理。
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}

		hs.handlers[m] = h
		hs.optionsState = optionsStateFixedHandler
	case http.MethodHead:
		if hs.headState == headStateFixed {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}

		hs.handlers[m] = h
		hs.headState = headStateFixed
	default:
		if !methodExists(m) {
			return fmt.Errorf("该请求方法 %s 不被支持", m)
		}

		if _, found := hs.handlers[m]; found {
			return fmt.Errorf("该请求方法 %s 已经存在", m)
		}
		hs.handlers[m] = h

		// GET 请求，且状态为 Auto 的时候，可以自动添加
		if m == http.MethodGet && hs.headState == headStateAuto {
			hs.handlers[http.MethodHead] = hs.headServeHTTP(h)
		}

		// 重新生成 optionsAllow 字符串
		if hs.optionsState == optionsStateDefault {
			hs.optionsAllow = hs.getOptionsAllow()
		}
	}
	return nil
}

func methodExists(m string) bool {
	for _, mm := range Methods {
		if mm == m {
			return true
		}
	}
	return false
}

func (hs *Handlers) optionsServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", hs.optionsAllow)
}

func (hs *Handlers) headServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&response{ResponseWriter: w}, r)
	})
}

func (hs *Handlers) getOptionsAllow() string {
	var index int
	for method := range hs.handlers {
		index += methodMap[method]
	}
	return optionsStrings[index]
}

// Remove 移除某个请求方法对应的处理函数
//
// 返回值表示是否已经被清空。
func (hs *Handlers) Remove(methods ...string) bool {
	if len(methods) == 0 {
		hs.handlers = make(map[string]http.Handler, 8)
		hs.optionsAllow = ""
		return true
	}

	for _, m := range methods {
		delete(hs.handlers, m)

		if m == http.MethodOptions { // 明确要删除 OPTIONS，则将其 optionsState 转为禁止使用
			hs.optionsState = optionsStateDisable
		} else if m == http.MethodGet && hs.headState == headStateAuto { // 跟随 get 一起删除
			delete(hs.handlers, http.MethodHead)
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

// SetAllow 设置 Options 请求头的 Allow 报头
func (hs *Handlers) SetAllow(optionsAllow string) {
	if hs.optionsState == optionsStateDisable {
		hs.handlers[http.MethodOptions] = http.HandlerFunc(hs.optionsServeHTTP)
	}
	hs.optionsAllow = optionsAllow
	hs.optionsState = optionsStateFixedString
}

// Handler 获取指定方法对应的处理函数
func (hs *Handlers) Handler(method string) http.Handler {
	return hs.handlers[method]
}

// Options 获取当前支持的请求方法列表字符串
func (hs *Handlers) Options() string {
	return hs.optionsAllow
}

// Len 获取当前支持请求方法数量
func (hs *Handlers) Len() int {
	return len(hs.handlers)
}

// Methods 当前节点支持的请求方法
func (hs *Handlers) Methods(ignoreHead, ignoreOptions bool) []string {
	methods := make([]string, 0, len(hs.handlers))

	for key := range hs.handlers {
		if (key == http.MethodOptions && ignoreOptions && hs.optionsState == optionsStateDefault) ||
			key == http.MethodHead && ignoreHead && hs.headState == headStateAuto {
			continue
		}

		methods = append(methods, key)
	}

	sort.Strings(methods)
	return methods
}
