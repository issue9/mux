// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"sort"
)

type methodType int16

// 各个请求方法的值。
// NOTE: 值类型为 methodType 的实际类型，不要溢出了。
const (
	get methodType = 1 << iota
	post
	del
	put
	patch
	options
	head
	connect
	trace
)

var (
	methodMap = map[string]methodType{
		http.MethodGet:     get,
		http.MethodPost:    post,
		http.MethodDelete:  del,
		http.MethodPut:     put,
		http.MethodPatch:   patch,
		http.MethodOptions: options,
		http.MethodHead:    head,
		http.MethodConnect: connect,
		http.MethodTrace:   trace,
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
