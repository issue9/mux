// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"sort"
)

var (
	methodMap = map[string]int{
		http.MethodGet:     1,
		http.MethodPost:    2,
		http.MethodDelete:  4,
		http.MethodPut:     8,
		http.MethodPatch:   16,
		http.MethodOptions: 32,
		http.MethodHead:    64,
		http.MethodConnect: 128,
		http.MethodTrace:   256,
	}

	// 除 OPTIONS 和 HEAD 之外的所有支持的元素
	// 在 Add 方法中用到。
	addAny = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodConnect,
		http.MethodTrace,
	}
)

// Methods 返回所有支持的请求方法名称
func Methods() []string {
	methods := make([]string, 0, len(methodMap))
	for method := range methodMap {
		methods = append(methods, method)
	}

	sort.Strings(methods)
	return methods
}
