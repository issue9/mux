// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package types 类型的前置声明
package types

import "errors"

var errParamNotExists = errors.New("不存在该参数")

// ErrParamNotExists 用于表示 [Params] 中参数不存在的错误
func ErrParamNotExists() error { return errParamNotExists }

// Params 表示路由中的参数操作接口
type Params interface {
	// Count 返回参数的数量
	Count() int

	// Get 获取指定名称的参数值
	Get(key string) (v string, found bool)

	// Exists 查找指定名称的参数是否存在
	Exists(key string) bool

	// String 获取地址参数中的名为 key 的变量并将其转换成 string
	//
	// 当参数不存在时，返回 [ErrParamNotExists] 错误。
	String(key string) (string, error)

	// MustString 获取地址参数中的名为 key 的变量并将其转换成 string
	//
	// 若不存在或是无法转换则返回 def。
	MustString(key, def string) string

	// Int 获取地址参数中的名为 key 的变量并将其转换成 int64
	//
	// 当参数不存在时，返回 [ErrParamNotExists] 错误。
	Int(key string) (int64, error)

	// MustInt 获取地址参数中的名为 key 的变量并将其转换成 int64
	//
	// 若不存在或是无法转换则返回 def。
	MustInt(key string, def int64) int64

	// Uint 获取地址参数中的名为 key 的变量并将其转换成 uint64
	//
	// 当参数不存在时，返回 [ErrParamNotExists] 错误。
	Uint(key string) (uint64, error)

	// MustUint 获取地址参数中的名为 key 的变量并将其转换成 uint64
	//
	// 若不存在或是无法转换则返回 def。
	MustUint(key string, def uint64) uint64

	// Bool 获取地址参数中的名为 key 的变量并将其转换成 bool
	//
	// 当参数不存在时，返回 [ErrParamNotExists] 错误。
	Bool(key string) (bool, error)

	// MustBool 获取地址参数中的名为 key 的变量并将其转换成 bool
	//
	// 若不存在或是无法转换则返回 def。
	MustBool(key string, def bool) bool

	// Float 获取地址参数中的名为 key 的变量并将其转换成 Float64
	//
	// 当参数不存在时，返回 [ErrParamNotExists] 错误。
	Float(key string) (float64, error)

	// MustFloat 获取地址参数中的名为 key 的变量并将其转换成 float64
	//
	// 若不存在或是无法转换则返回 def。
	MustFloat(key string, def float64) float64

	// Set 添加或是修改值
	Set(key, val string)

	// Range 依次访问每个参数
	Range(func(key, val string))
}

// Route 当前请求的路由信息
type Route interface {
	// Params 当前请求关联的参数
	Params() Params

	// Node 当前请求关联的节点信息
	//
	// 有可能返回 nil，比如请求到了 404。
	Node() Node

	// RouterName [mux.Router.Name] 的值
	RouterName() string
}

// Node 路由节点
type Node interface {
	// Pattern 路由上的匹配内容
	Pattern() string

	// Methods 当前节点支持的方法列表
	Methods() []string

	// AllowHeader Allow 报头的内容
	AllowHeader() string
}

// BuildNodeHandler 为节点生成处理方法
type BuildNodeHandler[T any] func(Node) T

// Middleware 中间件对象需要实现的接口
type Middleware[T any] interface {
	// Middleware 包装处理 next
	//
	// next 路由项的处理函数；
	// method 当前路由的请求方法；
	// pattern 当前路由的匹配项；
	// router 路由名称，即 [mux.Router.Name] 的值；
	//
	// NOTE: method 和 pattern 在某些特殊的路由项中会有特殊的值：
	//  - 404 method 和 pattern 均为空；
	//  - 405 method 为空，pattern 正常；
	Middleware(next T, method, pattern, router string) T
}

// MiddlewareFunc 中间件处理函数
type MiddlewareFunc[T any] func(next T, method, pattern, router string) T

func (f MiddlewareFunc[T]) Middleware(next T, method, pattern, router string) T {
	return f(next, method, pattern, router)
}
