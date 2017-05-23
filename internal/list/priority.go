// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/syntax"
)

// 按 Entry.Priority() 为优先级保存的 entry.Entry 列表。
type priority struct {
	mu             sync.RWMutex
	entries        []entry.Entry
	disableOptions bool
}

func newPriority(disableOptions bool) *priority {
	return &priority{
		disableOptions: disableOptions,
		entries:        make([]entry.Entry, 0, 100),
	}
}

// entries.clean
func (es *priority) clean(prefix string) {
	es.mu.Lock()
	defer es.mu.Unlock()

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

// entries.remove
func (es *priority) remove(pattern string, methods ...string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, e := range es.entries {
		if e.Pattern() != pattern {
			continue
		}

		if empty := e.Remove(methods...); empty { // 空了，则整个路由项都移除
			es.entries = removeEntries(es.entries, e.Pattern())
		}
		return // 只可能有一相完全匹配，找到之后，即可返回
	}
}

// entries.add
func (es *priority) add(s *syntax.Syntax, h http.Handler, methods ...string) error {
	ety, err := es.entry(s)
	if err != nil {
		return err
	}

	return ety.Add(h, methods...)
}

// entries.entry
func (es *priority) entry(s *syntax.Syntax) (entry.Entry, error) {
	if ety := es.findEntry(s.Pattern); ety != nil {
		return ety, nil
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

func (es *priority) findEntry(pattern string) entry.Entry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, e := range es.entries {
		if e.Pattern() == pattern {
			return e
		}
	}

	return nil
}

// entries.match
func (es *priority) match(path string) (entry.Entry, map[string]string) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	for _, ety := range es.entries {
		if matched, params := ety.Match(path); matched {
			return ety, params
		}
	}

	return nil, nil
}

// entries.len
func (es *priority) len() int {
	return len(es.entries)
}

func (es *priority) toSlash() (entries, error) {
	ret := newSlash(es.disableOptions)
	for _, ety := range es.entries {
		if err := ret.addEntry(ety); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// printDeep
func (es *priority) printDeep(deep int) {
	fmt.Println(strings.Repeat(" ", deep*4), "*********priority")
	for _, item := range es.entries {
		fmt.Println(strings.Repeat(" ", deep*4), item.Pattern())
	}
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
