// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import "net/http"

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

	methodTypeMap = map[methodType]string{
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

	// 除 http.MethodOptions 之外的所有支持的元素
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
