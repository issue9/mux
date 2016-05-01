// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"sort"
	"strings"
)

const (
	get int16 = 1 << iota
	post
	delete
	put
	patch
	options
	head
	trace
)

var (
	// 支持的所有请求方法
	supportedMethods = []string{
		"GET",
		"POST",
		"DELETE",
		"PUT",
		"PATCH",
		"OPTIONS",
		"HEAD",
		"TRACE",
	}

	tostr = map[int16]string{
		get:     "GET",
		post:    "POST",
		delete:  "DELETE",
		put:     "PUT",
		patch:   "PATCH",
		options: "OPTIONS",
		head:    "HEAD",
		trace:   "TRACE",
	}

	toint = map[string]int16{
		"GET":     get,
		"POST":    post,
		"DELETE":  delete,
		"PUT":     put,
		"PATCH":   patch,
		"OPTIONS": options,
		"HEAD":    head,
		"TRACE":   trace,
	}
)

// 根据数值取得其对应的Allow报头值。
func getAllowString(val int16) string {
	var ret []string
	for k, v := range tostr {
		if k&val > 0 {
			ret = append(ret, v)
		}
	}

	sort.Strings(ret)
	return strings.Join(ret, " ")
}

// 将一组method方法名称，转换成一个数值，若方法名称不存在，则不计算在内
func methodsToInt(methods ...string) int16 {
	var ret int16
	for _, method := range methods {
		ret |= toint[method]
	}
	return ret
}

func inStringSlice(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}

	return false
}
