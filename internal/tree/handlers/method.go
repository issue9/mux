// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"sort"
	"strings"

	"github.com/issue9/mux/internal/method"
)

// 所有的 OPTIONS 请求的 allow 报头字符串
var optionsStrings = make(map[method.Type]string, len(method.Supported))

func init() {
	methods := make([]string, 0, len(method.Supported))
	for i := method.Type(0); i <= method.Max; i++ {
		if i&method.Get == method.Get {
			methods = append(methods, method.String(method.Get))
		}
		if i&method.Post == method.Post {
			methods = append(methods, method.String(method.Post))
		}
		if i&method.Delete == method.Delete {
			methods = append(methods, method.String(method.Delete))
		}
		if i&method.Put == method.Put {
			methods = append(methods, method.String(method.Put))
		}
		if i&method.Patch == method.Patch {
			methods = append(methods, method.String(method.Patch))
		}
		if i&method.Options == method.Options {
			methods = append(methods, method.String(method.Options))
		}
		if i&method.Head == method.Head {
			methods = append(methods, method.String(method.Head))
		}
		if i&method.Connect == method.Connect {
			methods = append(methods, method.String(method.Connect))
		}
		if i&method.Trace == method.Trace {
			methods = append(methods, method.String(method.Trace))
		}

		sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
		optionsStrings[i] = strings.Join(methods, ", ")
		methods = methods[:0]
	}
}
