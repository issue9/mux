// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"strings"
)

var (
	// 支持的所有请求方法
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

	// 调用 Any 添加的列表，默认不添加 OPTIONS 请求
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
