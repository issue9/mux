// SPDX-License-Identifier: MIT

// Package handlers 用于处理节点下与处理函数相关的逻辑
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const defaultSize = 4 // Handlers.handlers 的初始容量

// Methods 所有支持请求方法
var Methods = []string{
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

// Add 未指定请求方法时，所采用的默认值。
var addAny = Methods[:len(Methods)-2]

// Handlers 用于表示某节点下各个请求方法对应的处理函数
type Handlers struct {
	handlers     map[string]http.Handler // 请求方法及其对应的 http.Handler
	disableHead  bool
	methods      []string
	optionsAllow string
}

// New 声明一个新的 Handlers 实例
//
// disableHead 是否禁止自动添加 HEAD 请求内容
func New(disableHead bool) *Handlers {
	return &Handlers{
		handlers:    make(map[string]http.Handler, defaultSize),
		disableHead: disableHead,
		methods:     make([]string, 0, defaultSize),
	}
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

	// 查看是否需要添加 OPTIONS
	if _, found := hs.handlers[http.MethodOptions]; !found {
		hs.handlers[http.MethodOptions] = http.HandlerFunc(hs.optionsServeHTTP)
	}

	hs.buildMethods()
	return nil
}

func (hs *Handlers) addSingle(h http.Handler, m string) error {
	if m == http.MethodHead || m == http.MethodOptions {
		return fmt.Errorf("无法手动添加 OPTIONS/HEAD 请求方法")
	}

	if !methodExists(m) {
		return fmt.Errorf("该请求方法 %s 不被支持", m)
	}

	if _, found := hs.handlers[m]; found {
		return fmt.Errorf("该请求方法 %s 已经存在", m)
	}
	hs.handlers[m] = h

	if m == http.MethodGet && !hs.disableHead { // 如果是 GET，则顺便添加 HEAD
		hs.handlers[http.MethodHead] = hs.headServeHTTP(h)
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
		h.ServeHTTP(&headResponse{ResponseWriter: w}, r)
	})
}

func (hs *Handlers) buildMethods() {
	hs.methods = hs.methods[:0]
	for method := range hs.handlers {
		hs.methods = append(hs.methods, method)
	}
	sort.Strings(hs.methods)
	hs.optionsAllow = strings.Join(hs.methods, ", ")
}

// Remove 移除某个请求方法对应的处理函数
//
// 返回值表示是否已经被清空。
func (hs *Handlers) Remove(methods ...string) (bool, error) {
	if len(methods) == 0 {
		hs.handlers = make(map[string]http.Handler, defaultSize)
		hs.buildMethods()
		return hs.Len() == 0, nil
	}

	for _, m := range methods {
		delete(hs.handlers, m)

		if m == http.MethodOptions {
			return false, errors.New("不能手动删除 OPTIONS。")
		} else if m == http.MethodGet { // HEAD 跟随 GET 一起删除
			delete(hs.handlers, http.MethodHead)
		}
	}

	if _, found := hs.handlers[http.MethodOptions]; found && hs.Len() == 1 { // 只有一个 OPTIONS 了
		delete(hs.handlers, http.MethodOptions)
	}

	hs.buildMethods()
	return hs.Len() == 0, nil
}

// Handler 获取指定方法对应的处理函数
func (hs *Handlers) Handler(method string) http.Handler { return hs.handlers[method] }

// Options 获取当前支持的请求方法列表字符串
func (hs *Handlers) Options() string { return hs.optionsAllow }

// Len 获取当前支持请求方法数量
func (hs *Handlers) Len() int { return len(hs.handlers) }

// Methods 当前节点支持的请求方法
func (hs *Handlers) Methods() []string { return hs.methods }
