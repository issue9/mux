// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMethods(t *testing.T) {
	a := assert.New(t)

	// methodMap、methodStringMap
	for typ, method := range methodMap {
		a.Equal(typ, methodStringMap[method])
	}
	a.Equal(len(methodMap), len(methodStringMap))

	// supported、any
	for _, m := range any {
		for _, mm := range supported {
			if mm == m {
				continue
			}
		}
		a.False(false, "supported 中 未包含 any 中的 %s", m)
	}

	// optionsString
	for index, allow := range optionsStrings {
		if index == 0 {
			a.Empty(allow)
		} else {
			a.NotEmpty(allow, "索引 %d 的值为空", index)
		}
	}
}

func TestOptionsStrings(t *testing.T) {
	a := assert.New(t)
	test := func(key methodType, str string) {
		a.Equal(optionsStrings[key], str, "key:%d,str:%s", key, str)
	}

	test(get+post, "GET, POST")
	test(get+post+options, "GET, OPTIONS, POST")
	test(get+post+options+del+trace, "DELETE, GET, OPTIONS, POST, TRACE")
	test(get+post+options+del+trace+head+patch, "DELETE, GET, HEAD, OPTIONS, PATCH, POST, TRACE")
}
