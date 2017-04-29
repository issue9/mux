// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"strings"
)

var (
	// 支持的请求方法
	supportedMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
		http.MethodTrace,
	}

	// 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为添加除 htp.MethodOptions 之外的所有 supportedMethods 中的元素
	defaultMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		// http.MethodOptions,
		http.MethodHead,
		http.MethodTrace,
	}
)

// SupportedMethods 获得 mux 包支持的请求方法列表。
func SupportedMethods() []string {
	return supportedMethods
}

// MethodIsSupported 是否支持该请求方法
func MethodIsSupported(method string) bool {
	method = strings.ToUpper(method)
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}
