// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/entry"
)

// 初始化时，默认的路由项数量大小，在一定的情况下，可以减少后期的内存多次分配操作
const defaultEntriesSize = 1000

// Entries entry.Entry 的存放列表
type Entries struct {
	mu             sync.RWMutex
	entries        []entry.Entry
	disableOptions bool
}

// NewEntries 声明一个 Entries 实例
func NewEntries(disableOptions bool) *Entries {
	return &Entries{
		disableOptions: disableOptions,
		entries:        make([]entry.Entry, 0, defaultEntriesSize),
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (es *Entries) Clean(prefix string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if len(prefix) == 0 {
		es.entries = es.entries[:0]
		return
	}

	dels := []string{}
	for _, ety := range es.entries {
		pattern := ety.Pattern()
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
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, e := range es.entries {
		if e.Pattern() != pattern {
			continue
		}

		if empty := e.Remove(methods...); empty { // 空了，则当整个路由项移除
			es.entries = removeEntries(es.entries, e.Pattern())
		}
		return // 只可能有一相完全匹配，找到之后，即可返回
	}
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 为可以匹配的请求方法，默认为 method.Default 中的所有元素，
// 可以为 method.Supported 中的所有元素。
// 当 h 或是 pattern 为空时，将触发 panic。
func (es *Entries) Add(pattern string, h http.Handler, methods ...string) error {
	ety, err := es.Entry(pattern)
	if err != nil {
		return err
	}

	return ety.Add(h, methods...)
}

// Entry 查找指定匹配模式下的 Entry，不存在，则声明新的
func (es *Entries) Entry(pattern string) (entry.Entry, error) {
	if ety := es.findEntry(pattern); ety != nil {
		return ety, nil
	}

	ety, err := entry.New(pattern)
	if err != nil {
		return nil, err
	}

	if es.disableOptions { // 禁用 OPTIONS
		ety.Remove(http.MethodOptions)
	}

	es.mu.Lock()
	defer es.mu.Unlock()
	es.entries = append(es.entries, ety)
	sort.SliceStable(es.entries, func(i, j int) bool { return es.entries[i].Priority() < es.entries[j].Priority() })

	return ety, nil
}

func (es *Entries) findEntry(pattern string) entry.Entry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, e := range es.entries {
		if e.Pattern() == pattern {
			return e
		}
	}

	return nil
}

// Match 查找与 path 最匹配的路由项以及对应的参数
func (es *Entries) Match(path string) (entry.Entry, map[string]string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, ety := range es.entries {
		if matched, params := ety.Match(path); matched {
			return ety, params
		}
	}

	return nil, nil
}

func removeEntries(es []entry.Entry, pattern string) []entry.Entry {
	lastIndex := len(es) - 1
	for index, e := range es {
		if e.Pattern() != pattern {
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
