// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/issue9/assert"
)

// 以上各个值进行组合之后的数量。
const max = get + post + del + put + patch + options + head + connect + trace

func TestMethods(t *testing.T) {
	a := assert.New(t)

	// 检测 methodMap 是否完整
	var val methodType
	for _, v := range methodMap {
		val += v
	}
	a.Equal(max, val, "methodMap 中的值与 max 不相同！")

	// removeAll 的内容是否都存在于 mehtodMap
	for _, m := range removeAll {
		_, found := methodMap[m]
		a.True(found)
	}

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

	test := func(key methodType, str string) {
		a.Equal(optionsStrings[key], str, "key:%d,str:%s", key, str)
	}

	test(0, "")
	test(get, "GET")
	test(get+post, "GET, POST")
	test(get+post+options, "GET, OPTIONS, POST")
	test(get+post+options+del+trace, "DELETE, GET, OPTIONS, POST, TRACE")
	test(get+post+options+del+trace+head+patch, "DELETE, GET, HEAD, OPTIONS, PATCH, POST, TRACE")
	test(max, "CONNECT, DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT, TRACE")
}
