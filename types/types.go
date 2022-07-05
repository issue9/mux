// SPDX-License-Identifier: MIT

// Package types 类型的前置声明
package types

// Params 表示路由中的参数操作接口
type Params interface {
	// Count 返回参数的数量
	Count() int

	// Get 获取指定名称的参数值
	//
	// 如果不存在此值，第二个值返回 false。
	Get(key string) (string, bool)

	// Exists 查找指定名称的参数是否存在
	//
	// NOTE: 如果是可选参数，可能会不存在。
	Exists(key string) bool

	// String 获取地址参数中的名为 key 的变量并将其转换成 string
	//
	// 当参数不存在时，返回 ErrParamNotExists 错误。
	String(key string) (string, error)

	// MustString 获取地址参数中的名为 key 的变量并将其转换成 string
	//
	// 若不存在或是无法转换则返回 def。
	MustString(key, def string) string

	// Int 获取地址参数中的名为 key 的变量并将其转换成 int64
	//
	// 当参数不存在时，返回 ErrParamNotExists 错误。
	Int(key string) (int64, error)

	// MustInt 获取地址参数中的名为 key 的变量并将其转换成 int64
	//
	// 若不存在或是无法转换则返回 def。
	MustInt(key string, def int64) int64

	// Uint 获取地址参数中的名为 key 的变量并将其转换成 uint64
	//
	// 当参数不存在时，返回 ErrParamNotExists 错误。
	Uint(key string) (uint64, error)

	// MustUint 获取地址参数中的名为 key 的变量并将其转换成 uint64
	//
	// 若不存在或是无法转换则返回 def。
	MustUint(key string, def uint64) uint64

	// Bool 获取地址参数中的名为 key 的变量并将其转换成 bool
	//
	// 当参数不存在时，返回 ErrParamNotExists 错误。
	Bool(key string) (bool, error)

	// MustBool 获取地址参数中的名为 key 的变量并将其转换成 bool
	//
	// 若不存在或是无法转换则返回 def。
	MustBool(key string, def bool) bool

	// Float 获取地址参数中的名为 key 的变量并将其转换成 Float64
	//
	// 当参数不存在时，返回 ErrParamNotExists 错误。
	Float(key string) (float64, error)

	// MustFloat 获取地址参数中的名为 key 的变量并将其转换成 float64
	//
	// 若不存在或是无法转换则返回 def。
	MustFloat(key string, def float64) float64

	// Set 添加或是修改值
	Set(key, val string)

	// Range 依次访问每个参数
	Range(func(key, val string))

	// Node 当前请求关联的节点信息
	//
	// 有可能返回 nil，比如请求到了 404。
	Node() Node

	// RouterName RouterOf[T].Name() 的值
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

// BuildNodeHandleOf 为节点生成处理方法
type BuildNodeHandleOf[T any] func(Node) T

// MiddlewareOf 中间件对象需要实现的接口
type MiddlewareOf[T any] interface {
	Middleware(T) T
}

// MiddlewareFuncOf 中间件处理函数
type MiddlewareFuncOf[T any] func(T) T

func (f MiddlewareFuncOf[T]) Middleware(next T) T { return f(next) }
