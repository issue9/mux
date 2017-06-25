// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/issue9/mux/internal/method"
)

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

// 所有的 OPTIONS 请求的 allow 报头字符串
var optionsStrings = make(map[methodType]string, len(method.Supported))

func init() {
	methods := make([]string, 0, len(method.Supported))
	for i := methodType(0); i < all; i++ {
		if i&get == get {
			methods = append(methods, methodStringMap[get])
		}
		if i&post == post {
			methods = append(methods, methodStringMap[post])
		}
		if i&del == del {
			methods = append(methods, methodStringMap[del])
		}
		if i&put == put {
			methods = append(methods, methodStringMap[put])
		}
		if i&patch == patch {
			methods = append(methods, methodStringMap[patch])
		}
		if i&options == options {
			methods = append(methods, methodStringMap[options])
		}
		if i&head == head {
			methods = append(methods, methodStringMap[head])
		}
		if i&connect == connect {
			methods = append(methods, methodStringMap[connect])
		}
		if i&trace == trace {
			methods = append(methods, methodStringMap[trace])
		}

		sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
		optionsStrings[i] = strings.Join(methods, ", ")
		methods = methods[:0]
	}
}
