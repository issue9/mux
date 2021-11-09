// SPDX-License-Identifier: MIT

// Package params 路由参数的相关声明
package params

import "errors"

// ErrParamNotExists 表示地址参数中并不存在该名称的值
var ErrParamNotExists = errors.New("不存在该参数")

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

	// Clone 复制当前参数对象
	Clone() Params

	// Map 转换成 map[string]string
	Map() map[string]string
}
