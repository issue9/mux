// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package list 提供了对 entry.Entry 元素的存储、匹配等功能。
package list

import (
	"errors"
	"net/http"
	"strings"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

const (
	maxSlashSize         = 255
	wildcardEntriesIndex = maxSlashSize
)

// List entry.Entry 列表，按 / 字符的多少来对 entry.Entry 实例进行分组，
// 以减少每次查询时的循环次数。
type List struct {
	entries        map[int]*entries
	disableOptions bool
}

// New 声明一个 List 实例
func New(disableOptions bool) *List {
	return &List{
		disableOptions: disableOptions,
		entries:        make(map[int]*entries, maxSlashSize),
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (l *List) Clean(prefix string) {
	if len(prefix) == 0 {
		l.entries = make(map[int]*entries, maxSlashSize)
		return
	}

	for _, es := range l.entries {
		es.clean(prefix)
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

	es, found := l.entries[getSlashSize(pattern)]
	if !found {
		return
	}

	es.remove(pattern, methods...)
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

	cnt := getSlashSize(pattern)
	es, found := l.entries[cnt]
	if !found {
		es = newEntries(l.disableOptions)
		l.entries[cnt] = es
	}

	return es.add(pattern, h, methods...)
}

// Entry 查找指定匹配模式下的 Entry，不存在，则声明新的
func (l *List) Entry(pattern string) (entry.Entry, error) {
	cnt := getSlashSize(pattern)
	es, found := l.entries[cnt]
	if !found {
		es = newEntries(l.disableOptions)
		l.entries[cnt] = es
	}

	return es.entry(pattern)
}

// Match 查找与 path 最匹配的路由项以及对应的参数
func (l *List) Match(path string) (entry.Entry, map[string]string) {
	cnt := strings.Count(path, "/")
	es, found := l.entries[cnt]
	if !found {
		return nil, nil
	}

	ety, ps := es.match(path)
	if ety == nil {
		if es, found = l.entries[wildcardEntriesIndex]; found {
			ety, ps = es.match(path)
		}
	}

	return ety, ps
}

func getSlashSize(str string) int {
	cnt := strings.Count(str, "/")
	if entry.IsWildcard(str) {
		cnt = wildcardEntriesIndex
	}

	return cnt
}
