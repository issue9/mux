// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"strconv"
)

type contextKey int

// ContextKeyParams 表示从 context 中获取的参数列表的关键字。
const ContextKeyParams contextKey = 0

// Params 用以保存请求地址中的参数内容
type Params map[string]string

// ErrParamNotExists 表示地址参数中并不存在该名称的值。
var ErrParamNotExists = errors.New("不存在该值")

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
