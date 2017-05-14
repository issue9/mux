// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"errors"
	"net/http"
	"strings"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

const defaultEntrySlashSize = 100

// List entry.Entry 列表
type List struct {
	entries        map[int]*Entries
	disableOptions bool
}

// New 声明一个 List 实例
func New(disableOptions bool) *List {
	return &List{
		disableOptions: disableOptions,
		entries:        make(map[int]*Entries, defaultEntrySlashSize),
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (l *List) Clean(prefix string) {
	if len(prefix) == 0 {
		l.entries = make(map[int]*Entries, defaultEntrySlashSize)
		return
	}

	for _, es := range l.entries {
		es.Clean(prefix)
	}
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (l *List) Remove(pattern string, methods ...string) {
	if len(methods) == 0 {
		methods = method.Supported
	}

	cnt := strings.Count(pattern, "/")
	es, found := l.entries[cnt]
	if !found {
		return
	}

	// TODO wildcard
	es.Remove(pattern, methods...)
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 为可以匹配的请求方法，默认为 method.Default 中的所有元素，
// 可以为 method.Supported 中的所有元素。
// 当 h 或是 pattern 为空时，将触发 panic。
func (l *List) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}
	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	if len(methods) == 0 {
		methods = method.Default
	}

	cnt := strings.Count(pattern, "/")
	es, found := l.entries[cnt]
	if !found {
		es = NewEntries(l.disableOptions)
		l.entries[cnt] = es
	}

	return es.Add(pattern, h, methods...)
}

// Entry 查找指定匹配模式下的 Entry，不存在，则声明新的
func (l *List) Entry(pattern string) (entry.Entry, error) {
	cnt := strings.Count(pattern, "/")
	es, found := l.entries[cnt]
	if !found {
		es = NewEntries(l.disableOptions)
		l.entries[cnt] = es
	}

	return es.Entry(pattern)
}

// Match 查找与 path 最匹配的路由项以及对应的参数
func (l *List) Match(path string) (entry.Entry, map[string]string) {
	cnt := strings.Count(path, "/")
	es, found := l.entries[cnt]
	if !found {
		return nil, nil
	}

	return es.Match(path)
}
