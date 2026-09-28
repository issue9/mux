// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"iter"
	"strconv"
	"sync"
)

var routePool = &sync.Pool{New: func() any { return &Route{} }}

// Route 当前请求的路由信息
type Route struct {
	Path   string // 实际请求的路径信息
	params map[string]string
	name   string
	node   Node
}

func NewRoute() *Route {
	route := routePool.Get().(*Route)
	route.Reset()
	return route
}

func (route *Route) Reset() {
	route.Path = ""
	clear(route.params)
	route.name = ""
	route.node = nil
}

func (route *Route) SetNode(n Node) { route.node = n }

// Node 当前请求关联的节点信息
//
// 有可能返回 nil，比如请求到了 404。
func (route *Route) Node() Node { return route.node }

func (route *Route) SetRouterName(n string) { route.name = n }

// RouterName 关联的路由对象名称
//
// 在多路由环境中可以使用此值区分路由
func (route *Route) RouterName() string { return route.name }

func (route *Route) Destroy() {
	const destroyMaxSize = 30
	if route != nil && len(route.params) <= destroyMaxSize {
		routePool.Put(route)
	}
}

// Exists 查找指定名称的参数是否存在
func (route *Route) Exists(key string) bool {
	_, found := route.Get(key)
	return found
}

func (route *Route) value[T any](key string, conv func(string) (T, error)) (T, error) {
	if str, found := route.Get(key); found {
		return conv(str)
	}

	var zero T
	return zero, ErrParamNotExists()
}

func (route *Route) mustValue[T any](key string, def T, conv func(string) (T, error)) T {
	if str, found := route.Get(key); found {
		if val, err := conv(str); err == nil {
			return val
		}
	}
	return def
}

// String 获取地址参数中的名为 key 的变量并将其转换成 string
//
// 当参数不存在时，返回 [ErrParamNotExists] 错误。
func (route *Route) String(key string) (string, error) {
	return route.value(key, func(s string) (string, error) { return s, nil })
}

// MustString 获取地址参数中的名为 key 的变量并将其转换成 string
//
// 若不存在或是无法转换则返回 def。
func (route *Route) MustString(key, def string) string {
	return route.mustValue(key, def, func(s string) (string, error) { return s, nil })
}

// Int 获取地址参数中的名为 key 的变量并将其转换成 int64
//
// 当参数不存在时，返回 [ErrParamNotExists] 错误。
func (route *Route) Int(key string) (int64, error) {
	return route.value(key, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
}

// MustInt 获取地址参数中的名为 key 的变量并将其转换成 int64
//
// 若不存在或是无法转换则返回 def。
func (route *Route) MustInt(key string, def int64) int64 {
	return route.mustValue(key, def, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
}

// Uint 获取地址参数中的名为 key 的变量并将其转换成 uint64
//
// 当参数不存在时，返回 [ErrParamNotExists] 错误。
func (route *Route) Uint(key string) (uint64, error) {
	return route.value(key, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
}

// MustUint 获取地址参数中的名为 key 的变量并将其转换成 uint64
//
// 若不存在或是无法转换则返回 def。
func (route *Route) MustUint(key string, def uint64) uint64 {
	return route.mustValue(key, def, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
}

// Bool 获取地址参数中的名为 key 的变量并将其转换成 bool
//
// 当参数不存在时，返回 [ErrParamNotExists] 错误。
func (route *Route) Bool(key string) (bool, error) {
	return route.value(key, func(s string) (bool, error) { return strconv.ParseBool(s) })
}

// MustBool 获取地址参数中的名为 key 的变量并将其转换成 bool
//
// 若不存在或是无法转换则返回 def。
func (route *Route) MustBool(key string, def bool) bool {
	return route.mustValue(key, def, func(s string) (bool, error) { return strconv.ParseBool(s) })
}

// Float 获取地址参数中的名为 key 的变量并将其转换成 Float64
//
// 当参数不存在时，返回 [ErrParamNotExists] 错误。
func (route *Route) Float(key string) (float64, error) {
	return route.value(key, func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
}

// MustFloat 获取地址参数中的名为 key 的变量并将其转换成 float64
//
// 若不存在或是无法转换则返回 def。
func (route *Route) MustFloat(key string, def float64) float64 {
	return route.mustValue(key, def, func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
}

// Get 获取指定名称的参数值
func (route *Route) Get(key string) (string, bool) {
	if route.params == nil {
		return "", false
	}
	v, f := route.params[key]
	return v, f
}

// Count 返回参数的数量
func (route *Route) Count() int { return len(route.params) }

// Set 添加或是修改值
func (route *Route) Set(k, v string) {
	if route.params == nil {
		route.params = map[string]string{k: v}
		return
	}
	route.params[k] = v
}

// Delete 删除指定名称的参数
func (route *Route) Delete(k string) {
	if route.params != nil {
		delete(route.params, k)
	}
}

// Params 返回遍历每个参数的迭代器
func (route *Route) Params() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for k, v := range route.params {
			if !yield(k, v) {
				break
			}
		}
	}
}
