// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"regexp"
)

// 支持正则表达式的匹配内容。
type entry struct {
	pattern string
	expr    *regexp.Regexp
	handler http.Handler
}

// 不检测h值是否为空，调用函数需要保证h值的合法性！
func newEntry(pattern string, h http.Handler) *entry {
	entry := &entry{
		pattern: pattern,
		handler: h,
		expr:    nil,
	}

	switch {
	case pattern[0] == '?':
		entry.pattern = pattern[1:]
		entry.expr = regexp.MustCompile(pattern[1:])
	case pattern[:2] == "\\?":
		entry.pattern = pattern[1:]
	default:
	}

	return entry
}

// 当前实例是否与参数匹配。
func (entry *entry) match(pattern string) bool {
	// 简单的字符串匹配
	if entry.expr == nil {
		return entry.pattern == pattern
	}

	return entry.expr.MatchString(pattern)
}

// 获取pattern中的命名捕获
func (entry *entry) getNamedCapture(pattern string) map[string]string {
	if entry.expr == nil {
		return nil
	}

	ret := make(map[string]string)
	subexps := entry.expr.SubexpNames()
	args := entry.expr.FindStringSubmatch(pattern)
	for index, name := range subexps {
		if len(name) == 0 {
			continue
		}

		ret[name] = args[index]
	}
	return ret
}
