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

// 各个请求方法的值
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

	none methodType = 0

	// 以上各个值进行组合之后的数量。
	max = get + post + del + put + patch + options + head + connect + trace + 1
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

	// 当前支持的所有请求方法
	supported = make([]string, 0, len(methodMap))

	// any 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为除 http.MethodOptions 之外的所有 supported 中的元素
	any = make([]string, 0, len(methodMap)-1)

	// 所有的 OPTIONS 请求的 allow 报头字符串
	optionsStrings = make([]string, max, max)
)

func init() {
	// 生成 supported 和 any
	for typ := range methodMap {
		supported = append(supported, typ)
		if typ != http.MethodOptions {
			any = append(any, typ)
		}
	}

	makeOptionsStrings()
}

func makeOptionsStrings() {
	methods := make([]string, 0, len(supported))
	for i := methodType(0); i < max; i++ {
		if i&get == get {
			methods = append(methods, http.MethodGet)
		}
		if i&post == post {
			methods = append(methods, http.MethodPost)
		}
		if i&del == del {
			methods = append(methods, http.MethodDelete)
		}
		if i&put == put {
			methods = append(methods, http.MethodPut)
		}
		if i&patch == patch {
			methods = append(methods, http.MethodPatch)
		}
		if i&options == options {
			methods = append(methods, http.MethodOptions)
		}
		if i&head == head {
			methods = append(methods, http.MethodHead)
		}
		if i&connect == connect {
			methods = append(methods, http.MethodConnect)
		}
		if i&trace == trace {
			methods = append(methods, http.MethodTrace)
		}

		sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
		optionsStrings[i] = strings.Join(methods, ", ")
		methods = methods[:0]
	}
}
