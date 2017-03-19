// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"net/http"
	"strconv"
)

type contextKey int

const contextKeyParams contextKey = 0

// ErrParamNotExists 表示地址参数中并不存在该名称的值。
var ErrParamNotExists = errors.New("不存在该值")

// Params 用以保存请求地址中的参数内容
type Params map[string]string

// GetParams 从 r 中获取路由参数。
//
// 以下情况两个参数都会返回 nil：
//  非正则路由；
//  正则路由，但是所有匹配参数都是未命名的；
func GetParams(r *http.Request) (Params, error) {
	params := r.Context().Value(contextKeyParams)
	if params == nil {
		return nil, nil
	}

	if ret, ok := params.(Params); ok {
		return ret, nil
	}
	return nil, errors.New("无法转换成 Params 类型")
}

// String 获取地址参数中的名为 key 的变量，并将其转换成 string
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
