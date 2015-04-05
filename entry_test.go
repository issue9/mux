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

	// 正则匹配
	e, err = newEntry("?abc", fn)
	a.NotError(err).Equal("abc", e.pattern)

	// pattern为空
	e, err = newEntry("", fn)
	a.Error(err).Nil(e)

	// pattern正则内容为空
	e, err = newEntry("?", fn)
	a.Error(err).Nil(e)
}

func testEntry_match(t *testing.T) {
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
			entry: makeEntry("hello world"),
			val:   nil,
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

		&test{ // 数字命名
			str:   "hz.zj.example.com",
			entry: makeEntry("?(?P<2>[a-z]*)\\.(?P<1>[a-z]*)\\.example\\.com"),
			val:   map[string]string{"2": "hz", "1": "zj"},
		},

		&test{ // 带一个未命名捕获
			str:   "yh.hz.zj.example.com",
			entry: makeEntry("?([a-z]*)\\.(?P<city>[a-z]*)\\.(?P<prov>[a-z]*)\\.example\\.com"),
			val:   map[string]string{"city": "hz", "prov": "zj"},
		},
	}

	a.NotNil(tests)

	/*for index, v := range tests {
		ok, mapped := v.entry.match(v.str)
		a.True(ok).
			Equal(mapped, v.val, "第[%v]个元素的值不相等:\nv1=%#v\nv2=%#v\n", index, mapped, v.val)
	}*/

	// 不能正确匹配
	//ok, mapped := tests[2].entry.match("hellow world")
	//a.False(ok).Nil(mapped)
}
