// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"sort"
	"strings"
)

type methodType int16

// 各个主求方法的值
const (
	none methodType = 0
	get  methodType = 1 << iota
	post
	del
	put
	patch
	options
	head
	connect
	trace
	max = trace // 最大值
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

	methodStringMap = map[methodType]string{
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
)

var (
	supported = make([]string, 0, len(methodMap))

	// Any 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为除 http.MethodOptions 之外的所有 Supported 中的元素
	any = make([]string, 0, len(methodMap)-1)
)

// 所有的 OPTIONS 请求的 allow 报头字符串
var optionsStrings = make(map[methodType]string, max)

func init() {
	for typ := range methodMap {
		supported = append(supported, typ)
		if typ != http.MethodOptions {
			any = append(any, typ)
		}
	}
}

func init() {
	methods := make([]string, 0, len(supported))
	for i := methodType(0); i <= max; i++ {
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
