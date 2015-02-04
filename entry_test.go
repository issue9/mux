// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestEntry_New(t *testing.T) {
	a := assert.New(t)

	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	// 普通情况
	e, err := newEntry("abc", fn)
	a.NotError(err).Equal(e.pattern, "abc")

	// 首字符为?的非正则匹配
	e, err = newEntry("\\?abc", fn)
	a.NotError(err).Equal("?abc", e.pattern)

	// 首字符为?的非正则匹配
	e, err = newEntry("\\?", fn)
	a.NotError(err).Equal("?", e.pattern)

	// 首字符为?的非正则匹配
	e, err = newEntry(`\?abc`, fn)
	a.NotError(err).Equal("?abc", e.pattern)

	// 正则匹配
	e, err = newEntry("?abc", fn)
	a.NotError(err).Equal("abc", e.pattern)

	// handler为空
	e, err = newEntry("?abc", nil)
	a.Error(err).Nil(e)

	// pattern为空
	e, err = newEntry("", fn)
	a.Error(err).Nil(e)

	// pattern正则内容为空
	e, err = newEntry("?", fn)
	a.Error(err).Nil(e)
}

func TestEntry_match_getNamedCapture(t *testing.T) {
	a := assert.New(t)

	makeEntry := func(pattern string) *entry {
		fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
		entry, err := newEntry(pattern, fn)
		a.NotError(err)
		return entry
	}

	type test struct {
		str   string
		entry *entry
		val   map[string]string
	}

	tests := []*test{
		&test{
			str:   "hello world",
			entry: makeEntry("?h(?P<n1>e)[a-z]+(?P<n2> )world"),
			val:   map[string]string{"n1": "e", "n2": " "},
		},

		&test{
			str:   "hello world",
			entry: makeEntry("?(?P<1>\\w+)\\s(?P<2>\\w+)"),
			val:   map[string]string{"1": "hello", "2": "world"},
		},

		&test{ // 带一个未命名的捕获组
			str:   "hello world",
			entry: makeEntry("?(?P<1>\\w+)(\\s+)(?P<2>\\w+)"),
			val:   map[string]string{"1": "hello", "2": "world"},
		},

		&test{
			str:   "bj.example.com",
			entry: makeEntry("?(?P<city>\\w+)\\.example\\.com"),
			val:   map[string]string{"city": "bj"},
		},
		&test{
			str:   "hz.zj.example.com",
			entry: makeEntry("?(?P<city>[a-z]*)\\.(?P<prov>[a-z]*)\\.example\\.com"),
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
