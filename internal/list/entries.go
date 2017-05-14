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
	"github.com/issue9/mux/internal/syntax"
)

type entries struct {
	mu             sync.RWMutex
	entries        []entry.Entry
	disableOptions bool
}

func newEntries(disableOptions bool) *entries {
	return &entries{
		disableOptions: disableOptions,
		entries:        make([]entry.Entry, 0, 100),
	}
}

// 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (es *entries) clean(prefix string) {
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

// 移除指定的路由项。
func (es *entries) remove(pattern string, methods ...string) {
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

// 添加一条路由数据。
func (es *entries) add(pattern string, h http.Handler, methods ...string) error {
	ety, err := es.entry(pattern)
	if err != nil {
		return err
	}

	return ety.Add(h, methods...)
}

// 查找指定匹配模式下的 Entry，不存在，则声明新的
func (es *entries) entry(pattern string) (entry.Entry, error) {
	if ety := es.findEntry(pattern); ety != nil {
		return ety, nil
	}

	s, err := syntax.New(pattern)
	if err != nil {
		return nil, err
	}

	ety, err := entry.New(s)
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

func (es *entries) findEntry(pattern string) entry.Entry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, e := range es.entries {
		if e.Pattern() == pattern {
			return e
		}
	}

	return nil
}

// 查找与 path 最匹配的路由项以及对应的参数
func (es *entries) match(path string) (entry.Entry, map[string]string) {
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
