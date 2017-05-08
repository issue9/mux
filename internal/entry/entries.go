// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/method"
)

// Entries 列表
type Entries struct {
	mu sync.RWMutex

	// 是否禁用自动产生 OPTIONS 请求方法。
	// 该值不能中途修改，否则会出现部分有 OPTIONS，部分没有的情况。
	disableOptions bool

	// 路由项，按资源进行分类。
	entries []Entry
}

// NewEntries 声明一个 Entries 实例
func NewEntries(disableOptions bool) *Entries {
	return &Entries{
		disableOptions: disableOptions,
		entries:        make([]Entry, 0, 1000),
	}
}

// Clean 清除所有的路由项
func (es *Entries) Clean(prefix string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if len(prefix) == 0 {
		es.entries = es.entries[:0]
		return
	}

	dels := []string{}
	for _, ety := range es.entries {
		pattern := ety.pattern()
		if strings.HasPrefix(pattern, prefix) {
			dels = append(dels, pattern)
		}
	} // end for

	for _, pattern := range dels {
		es.entries = removeEntries(es.entries, pattern)
	}
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (es *Entries) Remove(pattern string, methods ...string) {
	if len(methods) == 0 { // 删除所有 method 下匹配的项
		methods = method.Supported
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	for _, e := range es.entries {
		if e.pattern() != pattern {
			continue
		}

		if empty := e.Remove(methods...); empty { // 该 Entry 下已经没有路由项了
			es.entries = removeEntries(es.entries, e.pattern())
		}
		return // 只可能有一相完全匹配，找到之后，即可返回
	}
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// methods 参数应该只能为 method.Default 中的字符串，若不指定，默认为所有，
// 当 h 或是 pattern 为空时，将触发 panic。
func (es *Entries) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}
	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	ety := es.Entry(pattern)
	if ety == nil { // 不存在相同的资源项，则声明新的。
		var err error
		if ety, err = New(pattern, h); err != nil {
			return err
		}

		if es.disableOptions { // 禁用 OPTIONS
			ety.Remove(http.MethodOptions)
		}

		es.mu.Lock()
		defer es.mu.Unlock()
		es.entries = append(es.entries, ety)
		sort.SliceStable(es.entries, func(i, j int) bool { return es.entries[i].priority() < es.entries[j].priority() })
	}

	return ety.Add(h, methods...)
}

// Entry 查找指定匹配模式下的 Entry
func (es *Entries) Entry(pattern string) Entry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, e := range es.entries {
		if e.pattern() == pattern {
			return e
		}
	}

	return nil
}

// Match 查找与 path 最匹配的路由项
//
// e 为当前匹配的 Entry 实例。
func (es *Entries) Match(path string) (e Entry) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, ety := range es.entries {
		if ety.match(path) {
			return ety
		}
	} // end for

	return nil
}

func removeEntries(es []Entry, pattern string) []Entry {
	lastIndex := len(es) - 1
	for index, e := range es {
		if e.pattern() != pattern {
			continue
		}

		switch {
		case len(es) == 1: // 只有一个元素
			return es[:0]
		case index == lastIndex: // 最后一个元素
			return es[:lastIndex]
		default:
			return append(es[:index], es[index+1:]...)
		}
	} // end for

	return es
}
