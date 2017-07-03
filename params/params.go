// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package params 获取和转换路由中的参数信息。
package params

import (
	"errors"
	"net/http"
	"strconv"
)

type contextKey int

// ContextKeyParams 存取路由参数的关键字
const ContextKeyParams contextKey = 0

// ErrParamNotExists 表示地址参数中并不存在该名称的值。
var ErrParamNotExists = errors.New("不存在该参数")

// Params 获取和转换路由中的参数信息。
type Params map[string]string

// Get 获得一个 Params 实例。
//
// 以下情况两个参数都会返回 nil：
//  非正则和命名路由；
//  正则路由，但是所有匹配参数都是未命名的；
func Get(r *http.Request) Params {
	params := r.Context().Value(ContextKeyParams)
	if params == nil {
		return nil
	}

	return params.(Params)
}

// Exists 查找指定名称的参数是否存在。
// 可选参数，可能会不存在。
func (p Params) Exists(key string) bool {
	_, found := p[key]
	return found
}

// String 获取地址参数中的名为 key 的变量，并将其转换成 string
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) String(key string) (string, error) {
	v, found := p[key]
	if !found {
		return "", ErrParamNotExists
	}

	return v, nil
}

// MustString 获取地址参数中的名为 key 的变量，并将其转换成 string，
// 若不存在或是无法转换则返回 def。
func (p Params) MustString(key, def string) string {
	v, found := p[key]
	if !found {
		return def
	}

	return v
}

// Int 获取地址参数中的名为 key 的变量，并将其转换成 int64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) Int(key string) (int64, error) {
	str, found := p[key]
	if !found {
		return 0, ErrParamNotExists
	}

	return strconv.ParseInt(str, 10, 64)
}

// MustInt 获取地址参数中的名为 key 的变量，并将其转换成 int64，
// 若不存在或是无法转换则返回 def。
func (p Params) MustInt(key string, def int64) int64 {
	str, found := p[key]
	if !found {
		return def
	}

	if val, err := strconv.ParseInt(str, 10, 64); err == nil {
		return val
	}

	return def
}

// Uint 获取地址参数中的名为 key 的变量，并将其转换成 uint64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) Uint(key string) (uint64, error) {
	str, found := p[key]
	if !found {
		return 0, ErrParamNotExists
	}

	return strconv.ParseUint(str, 10, 64)
}

// MustUint 获取地址参数中的名为 key 的变量，并将其转换成 uint64，
// 若不存在或是无法转换则返回 def。
func (p Params) MustUint(key string, def uint64) uint64 {
	str, found := p[key]
	if !found {
		return def
	}

	if val, err := strconv.ParseUint(str, 10, 64); err == nil {
		return val
	}

	return def
}

// Bool 获取地址参数中的名为 key 的变量，并将其转换成 bool
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) Bool(key string) (bool, error) {
	str, found := p[key]
	if !found {
		return false, ErrParamNotExists
	}

	return strconv.ParseBool(str)
}

// MustBool 获取地址参数中的名为 key 的变量，并将其转换成 bool，
// 若不存在或是无法转换则返回 def。
func (p Params) MustBool(key string, def bool) bool {
	str, found := p[key]
	if !found {
		return def
	}

	if val, err := strconv.ParseBool(str); err == nil {
		return val
	}

	return def
}

// Float 获取地址参数中的名为 key 的变量，并将其转换成 Float64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) Float(key string) (float64, error) {
	str, found := p[key]
	if !found {
		return 0, ErrParamNotExists
	}

	return strconv.ParseFloat(str, 64)
}

// MustFloat 获取地址参数中的名为 key 的变量，并将其转换成 float64，
// 若不存在或是无法转换则返回 def。
func (p Params) MustFloat(key string, def float64) float64 {
	str, found := p[key]
	if !found {
		return def
	}

	if val, err := strconv.ParseFloat(str, 64); err == nil {
		return val
	}

	return def
}
