// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
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
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
		http.MethodTrace,
	}

	tostr = map[int16]string{
		get:     http.MethodGet,
		post:    http.MethodPost,
		delete:  http.MethodDelete,
		put:     http.MethodPut,
		patch:   http.MethodPatch,
		options: http.MethodOptions,
		head:    http.MethodHead,
		trace:   http.MethodTrace,
	}

	toint = map[string]int16{
		http.MethodGet:     get,
		http.MethodPost:    post,
		http.MethodDelete:  delete,
		http.MethodPut:     put,
		http.MethodPatch:   patch,
		http.MethodOptions: options,
		http.MethodHead:    head,
		http.MethodTrace:   trace,
	}
)

// 是否支持某请求方法
func MethodIsSupported(method string) bool {
	method = strings.ToUpper(method)
	for _, m := range supportedMethods {
		if m == method {
			return true
		}
	}

	return false
}

// 根据数值取得其对应的 Allow 报头值。
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

// 将一组 method 方法名称，转换成一个数值，若方法名称不存在，则不计算在内
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
