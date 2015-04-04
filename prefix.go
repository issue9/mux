// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// 按前缀分组。
type Prefix struct {
	mu             sync.Mutex
	items          map[string]http.Handler
	defaultHandler http.Handler
}

// 声明一个Prefix实例。
func NewPrefix() *Prefix {
	return &Prefix{
		items: map[string]http.Handler{},
	}
}

// 添加一个分组。
//
// 参数h不能为空值，否则会返回错误信息。
// prefix用于指定前缀，若为空，则在其它都无法匹配的情况下，
// 使用该接口用于默认的操作。
func (p *Prefix) Add(prefix string, h http.Handler) error {
	if len(prefix) == 0 {
		if p.defaultHandler != nil {
			return errors.New("Add:已经指定一个默认处理函数")
		}
		p.defaultHandler = h
		return nil
	}

	if h == nil {
		return errors.New("Add:参数h不能为空值")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, found := p.items[prefix]; found {
		return fmt.Errorf("Add:该前缀名称[%v]已经存在", prefix)
	}

	p.items[prefix] = h
	return nil
}

func (p *Prefix) AddFunc(prefix string, f func(http.ResponseWriter, *http.Request)) error {
	return p.Add(prefix, http.HandlerFunc(f))
}

// implement net/http.Handler.ServeHTTP(...)
func (p *Prefix) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for prefix, h := range p.items {
		if !strings.HasPrefix(req.URL.Path, prefix) {
			continue
		}

		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		h.ServeHTTP(w, req)
		return
	}

	if p.defaultHandler != nil {
		p.defaultHandler.ServeHTTP(w, req)
		return
	}

	panic(fmt.Sprintf("没有找到与之匹配的Prefix"))
}
