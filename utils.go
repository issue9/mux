// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

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

var tostr = map[int16]string{
	get:     "GET",
	post:    "POST",
	delete:  "DELETE",
	put:     "PUT",
	patch:   "PATCH",
	options: "OPTIONS",
	head:    "HEAD",
	trace:   "TRACE",
}

var toint = map[string]int16{
	"GET":     get,
	"POST":    post,
	"DELETE":  delete,
	"PUT":     put,
	"PATCH":   patch,
	"OPTIONS": options,
	"HEAD":    head,
	"TRACE":   trace,
}

func getAllowString(val int16) string {
	return ""
}

// 判断一个字符串是否在另一个字符串数组中。
func inStringSlice(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// 判断一个字符串是否在另一个字符串数组中。
func indexStringSlice(slice []string, val string) int {
	for k, v := range slice {
		if v == val {
			return k
		}
	}
	return -1
}

func deleteStringsSlice(slice []string, val ...string) []string {
	for _, v := range val {
		index := indexStringSlice(slice, v)
		if index >= 0 {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}
