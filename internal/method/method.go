// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package method

import (
	"net/http"
	"strings"
)

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
		http.MethodTrace,
	}

	// Default 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为添加除 http.MethodOptions 之外的所有 supportedMethods 中的元素
	Default = []string{
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

// IsSupported 是否支持该请求方法
func IsSupported(method string) bool {
	method = strings.ToUpper(method)
	for _, m := range Supported {
		if m == method {
			return true
		}
	}

	return false
}
