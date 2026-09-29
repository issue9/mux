// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package types 类型的前置声明
package types

import "errors"

var errParamNotExists = errors.New("不存在该参数")

// ErrParamNotExists 用于表示 [Params] 中参数不存在的错误
func ErrParamNotExists() error { return errParamNotExists }

// Node 路由节点
type Node interface {
	// Pattern 路由上的匹配内容
	Pattern() string

	// Methods 当前节点支持的方法列表
	//
	// NOTE: 返回切片不能修改
	Methods() []string

	// AllowHeader Allow 报头的内容
	AllowHeader() string
}

// BuildNodeHandler 定义了为节点 node 生成路由处理方法的类型
//
// node 为节点类型，该值可能为空值，表示不针对特定的类型。
// 返回的类型 T 为路由项的处理函数。
type BuildNodeHandler[T any] func(node Node) T

// Middleware 中间件
type Middleware[T any] interface {
	// Middleware 调整路由项 next 的行为
	//
	// next 路由项的处理函数；
	// method 当前路由的请求方法；
	// pattern 当前路由的匹配项；
	// router 路由名称，即 [mux.Router.Name] 的值；
	//
	// NOTE: method 和 pattern 在某些特殊的路由项中会有特殊的值：
	//  - 404 method 和 pattern 均为空；
	//  - 405 method 为空，pattern 正常；
	//  - TRACE 请求则 pattern 为空；
	//
	// NOTE: 此方法本身仅执行一次，返回的对象才会在每次请求时都执行。
	Middleware(next T, method, pattern, router string) T
}

// MiddlewareFunc 中间件
type MiddlewareFunc[T any] func(next T, method, pattern, router string) T

func (f MiddlewareFunc[T]) Middleware(next T, method, pattern, router string) T {
	return f(next, method, pattern, router)
}

// InterceptorFunc 拦截器的处理函数
type InterceptorFunc func(string) bool
