// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import "net/http"

type methodType int16

// 表示请求方法的常量，必须要与 github.com/issue9/mux/internal/method.Supported
// 中的各个元素一一对应。
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

	all = trace + connect + head + options + patch + put + del + post + get
)

var methodMap = map[string]methodType{
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

var methodStringMap = map[methodType]string{
	get:     http.MethodGet,
	post:    http.MethodPost,
	del:     http.MethodDelete,
	put:     http.MethodPut,
	patch:   http.MethodPatch,
	options: http.MethodOptions,
	head:    http.MethodHead,
	connect: http.MethodConnect,
	trace:   http.MethodTrace,
}
