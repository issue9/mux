// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/issue9/assert"
)

// 以上各个值进行组合之后的数量。
var max int

func init() {
	for _, v := range methodMap {
		max += v
	}
}

func TestMethods(t *testing.T) {
	a := assert.New(t)

	// 检测 methodMap 是否完整
	var val int
	for _, v := range methodMap {
		val += v
	}
	a.Equal(max, val, "methodMap 中的值与 max 不相同！")

	// addAny 的内容是否都存在于 mehtodMap
	for _, m := range addAny {
		_, found := methodMap[m]
		a.True(found)
	}
}

func TestOptionsStrings(t *testing.T) {
	a := assert.New(t)

	for index, allow := range optionsStrings {
		if index == 0 {
			a.Empty(allow)
		} else {
			a.NotEmpty(allow, "索引 %d 的值为空", index)
		}
	}

	test := func(key int, str string) {
		a.Equal(optionsStrings[key], str, "key:%d,val:%s", key, optionsStrings[key])
	}

	test(0, "")
	test(1, "GET")
	test(1+2, "GET, POST")
	test(max, "CONNECT, DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT, TRACE")
}
