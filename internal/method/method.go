// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package method 指定 mux 允许使用的请求方法。
package method

import (
	"net/http"
	"strings"
)

// Type 以数值的方式表示请求方法的类型
type Type int16

// 各个主求方法的值
const (
	None Type = 0
	Get  Type = 1 << iota
	Post
	Delete
	Put
	Patch
	Options
	Head
	Connect
	Trace

	Max = Trace // 最大值
)

var (
	methodMap = map[string]Type{
		http.MethodGet:     Get,
		http.MethodPost:    Post,
		http.MethodDelete:  Delete,
		http.MethodPut:     Put,
		http.MethodPatch:   Patch,
		http.MethodOptions: Options,
		http.MethodHead:    Head,
		http.MethodConnect: Connect,
		http.MethodTrace:   Trace,
	}

	methodStringMap = map[Type]string{
		Get:     http.MethodGet,
		Post:    http.MethodPost,
		Delete:  http.MethodDelete,
		Put:     http.MethodPut,
		Patch:   http.MethodPatch,
		Options: http.MethodOptions,
		Head:    http.MethodHead,
		Connect: http.MethodConnect,
		Trace:   http.MethodTrace,
	}
)

func (t Type) String() string {
	return methodStringMap[t]
}

// String 将一个 Type 类型转换成相对应的字符串
func String(v Type) string {
	return v.String()
}

// Int 将一个字符串转换成相对应的 Type 数值
func Int(str string) Type {
	return methodMap[str]
}

var (
	// Supported 支持的请求方法
	Supported = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
		http.MethodConnect,
		http.MethodTrace,
	}

	// Default 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为除 http.MethodOptions 之外的所有 Supported 中的元素
	Default = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		// http.MethodOptions,
		http.MethodHead,
		http.MethodConnect,
		http.MethodTrace,
	}
)

// IsSupported 是否支持该请求方法
func IsSupported(method string) bool {
	method = strings.ToUpper(method)
	return Int(method) != None
}
