// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package entries 管理 entry.Entry 的添加删除匹配等工作
package entries

import (
	"container/list"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

// Entries 列表
type Entries struct {
	mu sync.RWMutex

	// 是否禁用自动产生 OPTIONS 请求方法。
	// 该值不能中途修改，否则会出现部分有 OPTIONS，部分没有的情况。
	disableOptions bool

	// 路由项，按资源进行分类。
	entries *list.List
}

// New 声明一个 Entries 实例
func New(disableOptions bool) *Entries {
	return &Entries{
		disableOptions: disableOptions,
		entries:        list.New(),
	}
}

// Clean 清除所有的路由项
func (es *Entries) Clean(prefix string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if len(prefix) == 0 {
		es.entries.Init()
		return
	}

	for item := es.entries.Front(); item != nil; {
		curr := item
		item = item.Next() // 提前记录下个元素，因为 item 有可能被删除

		ety := curr.Value.(entry.Entry)
		pattern := ety.Pattern()
		if strings.HasPrefix(pattern, prefix) {
			if empty := ety.Remove(method.Supported...); empty {
				es.entries.Remove(curr)
			}
		}
	} // end for
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

	for item := es.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() != pattern {
			continue
		}

		if empty := e.Remove(methods...); empty { // 该 Entry 下已经没有路由项了
			es.entries.Remove(item)
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
		if ety, err = entry.New(pattern, h); err != nil {
			return err
		}

		if es.disableOptions { // 禁用 OPTIONS
			ety.Remove(http.MethodOptions)
		}

		es.mu.Lock()
		defer es.mu.Unlock()
		if ety.Type() == entry.TypeRegexp { // 正则路由，在后端插入
			es.entries.PushBack(ety)
		} else {
			es.entries.PushFront(ety)
		}
	}

	return ety.Add(h, methods...)
}

// Entry 查找指定匹配模式下的 Entry
func (es *Entries) Entry(pattern string) entry.Entry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for item := es.entries.Front(); item != nil; item = item.Next() {
		e := item.Value.(entry.Entry)
		if e.Pattern() == pattern {
			return e
		}
	}

	return nil
}

// Match 查找与 path 最匹配的路由项
//
// e 为当前匹配的 entry.Entry 实例。
func (es *Entries) Match(path string) (e entry.Entry) {
	size := -1 // 匹配度，0 表示完全匹配，-1 表示完全不匹配，其它值越小匹配度越高

	es.mu.RLock()
	defer es.mu.RUnlock()

	for item := es.entries.Front(); item != nil; item = item.Next() {
		ety := item.Value.(entry.Entry)
		s := ety.Match(path)

		if s == 0 { // 完全匹配，可以中止匹配过程
			return ety
		}

		if s == -1 || (size > 0 && s >= size) { // 完全不匹配，或是匹配度没有当前的高
			continue
		}

		// 匹配度比当前的高，则保存下来
		size = s
		e = ety
	} // end for

	if size < 0 {
		return nil
	}
	return e
}
