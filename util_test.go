// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"regexp"
	"testing"

	"github.com/caixw/lib.go/assert"
)

func TestParseCaptures(t *testing.T) {
	a := assert.New(t)

	type test struct {
		str string
		exp *regexp.Regexp
		val map[string]string
	}

	tests := []*test{
		&test{
			str: "hello world",
			exp: regexp.MustCompile("h(?P<n1>e)[a-z]+(?P<n2> )world"),
			val: map[string]string{"n1": "e", "n2": " "},
		},
		&test{
			str: "hello world",
			exp: regexp.MustCompile("(?P<1>\\w+)\\s(?P<2>\\w+)"),
			val: map[string]string{"1": "hello", "2": "world"},
		},
		// 带一个未命名的捕获组
		&test{
			str: "hello world",
			exp: regexp.MustCompile("(?P<1>\\w+)(\\s+)(?P<2>\\w+)"),
			val: map[string]string{"1": "hello", "2": "world"},
		},
	}

	fn := func(pt *test, index int) {
		m := parseCaptures(pt.exp, pt.str)
		a.Equal(m, pt.val, "第[%v]个元素的值不相等:\nv1=%v\nv2=%v\n", index, m, pt.val)
	}

	for index, v := range tests {
		fn(v, index)
	}
}
