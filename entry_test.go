// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestEntry_New(t *testing.T) {
	a := assert.New(t)

	// 普通情况
	e := newEntry("abc", nil)
	a.Equal(e.pattern, "abc")

	// 首字符为?的非正则匹配
	e = newEntry("\\?abc", nil)
	a.Equal("?abc", e.pattern)

	// 首字符为?的非正则匹配
	e = newEntry(`\?abc`, nil)
	a.Equal("?abc", e.pattern)

	// 正则匹配
	e = newEntry("?abc", nil)
	a.Equal("abc", e.pattern)
}

func TestEntry_match_getNamedCapture(t *testing.T) {
	a := assert.New(t)

	type test struct {
		str   string
		entry *entry
		val   map[string]string
	}

	tests := []*test{
		&test{
			str:   "hello world",
			entry: newEntry("?h(?P<n1>e)[a-z]+(?P<n2> )world", nil),
			val:   map[string]string{"n1": "e", "n2": " "},
		},

		&test{
			str:   "hello world",
			entry: newEntry("?(?P<1>\\w+)\\s(?P<2>\\w+)", nil),
			val:   map[string]string{"1": "hello", "2": "world"},
		},

		&test{ // 带一个未命名的捕获组
			str:   "hello world",
			entry: newEntry("?(?P<1>\\w+)(\\s+)(?P<2>\\w+)", nil),
			val:   map[string]string{"1": "hello", "2": "world"},
		},

		&test{
			str:   "bj.example.com",
			entry: newEntry("?(?P<city>\\w+)\\.example\\.com", nil),
			val:   map[string]string{"city": "bj"},
		},
		&test{
			str:   "hz.zj.example.com",
			entry: newEntry("?(?P<city>[a-z]*)\\.(?P<prov>[a-z]*)\\.example\\.com", nil),
			val:   map[string]string{"city": "hz", "prov": "zj"},
		},
	}

	for index, v := range tests {
		ok := v.entry.match(v.str)
		a.True(ok)
		mapped := v.entry.getNamedCapture(v.str)
		a.Equal(mapped, v.val, "第[%v]个元素的值不相等:\nv1=%v\nv2=%v\n", index, mapped, v.val)
	}
}
