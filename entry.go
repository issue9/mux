// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"net/http"
	"regexp"
)

// 支持正则表达式的匹配内容。
type entry struct {
	pattern string
	expr    *regexp.Regexp
	handler http.Handler
}

// 当pattern或是h为空时，将返回错误信息。
func newEntry(pattern string, h http.Handler) (*entry, error) {
	if len(pattern) == 0 {
		return nil, errors.New("newEntry:参数pattern不能为空值")
	}

	if h == nil {
		return nil, errors.New("newEntry:参数h不能为空值")
	}

	entry := &entry{
		pattern: pattern,
		handler: h,
		expr:    nil,
	}

	switch {
	case pattern[0] == '?':
		if len(pattern) == 1 {
			return nil, errors.New("newEntry:pattern正则内容不能为空")
		}
		entry.pattern = pattern[1:]
		entry.expr = regexp.MustCompile(pattern[1:])
	case pattern[:2] == "\\?":
		entry.pattern = pattern[1:]
	default:
	}

	return entry, nil
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
