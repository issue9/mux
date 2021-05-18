// SPDX-License-Identifier: MIT

// Package route 处理路由节点的逻辑
package route

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/issue9/sliceutil"
)

const defaultSize = 4 // Route.handlers 的初始容量

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

// Route 用于表示某节点下各个请求方法对应的处理函数
type Route struct {
	handlers    map[string]http.Handler // 请求方法及其对应的 http.Handler
	disableHead bool
	methodIndex int
}

// New 声明一个新的 Route 实例
//
// disableHead 是否禁止自动添加 HEAD 请求内容
func New(disableHead bool) *Route {
	return &Route{
		handlers:    make(map[string]http.Handler, defaultSize),
		disableHead: disableHead,
	}
}

// Add 添加一个处理函数
func (r *Route) Add(h http.Handler, methods ...string) error {
	if len(methods) == 0 {
		methods = addAny
	}

	for _, m := range methods {
		if err := r.add(h, m); err != nil {
			return err
		}
	}

	// 查看是否需要添加 OPTIONS
	if _, found := r.handlers[http.MethodOptions]; !found {
		r.handlers[http.MethodOptions] = http.HandlerFunc(r.optionsServeHTTP)
	}

	r.buildMethods()
	return nil
}

func (r *Route) add(h http.Handler, m string) error {
	if m == http.MethodHead || m == http.MethodOptions {
		return fmt.Errorf("无法手动添加 OPTIONS/HEAD 请求方法")
	}

	if sliceutil.Index(Methods, func(i int) bool { return Methods[i] == m }) == -1 {
		return fmt.Errorf("该请求方法 %s 不被支持", m)
	}

	if _, found := r.handlers[m]; found {
		return fmt.Errorf("该请求方法 %s 已经存在", m)
	}
	r.handlers[m] = h

	if m == http.MethodGet && !r.disableHead { // 如果是 GET，则顺便添加 HEAD
		r.handlers[http.MethodHead] = r.headServeHTTP(h)
	}

	return nil
}

func (r *Route) optionsServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Allow", r.Options())
}

// Remove 移除某个请求方法对应的处理函数
//
// 返回值表示是否已经被清空。
func (r *Route) Remove(methods ...string) (bool, error) {
	if len(methods) == 0 {
		r.handlers = make(map[string]http.Handler, defaultSize)
		r.buildMethods()
		return r.Len() == 0, nil
	}

	for _, m := range methods {
		delete(r.handlers, m)

		if m == http.MethodOptions {
			return false, errors.New("不能手动删除 OPTIONS。")
		} else if m == http.MethodGet { // HEAD 跟随 GET 一起删除
			delete(r.handlers, http.MethodHead)
		}
	}

	if _, found := r.handlers[http.MethodOptions]; found && r.Len() == 1 { // 只有一个 OPTIONS 了
		delete(r.handlers, http.MethodOptions)
	}

	r.buildMethods()
	return r.Len() == 0, nil
}

// Handler 获取指定方法对应的处理函数
func (r *Route) Handler(method string) http.Handler { return r.handlers[method] }

// Options 获取当前支持的请求方法列表字符串
func (r *Route) Options() string { return indexes[r.methodIndex].options }

// Len 获取当前支持请求方法数量
func (r *Route) Len() int { return len(r.handlers) }

// Methods 当前节点支持的请求方法
func (r *Route) Methods() []string { return indexes[r.methodIndex].methods }
