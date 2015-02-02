// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"sync"
)

// 用于匹配域名的http.Handler
//
//  m1 := mux.NewMethod(nil).
//            MustGet(h1).
//            MustPost(h2)
//  m2 := mux.NewMethod(nil).
//            MustGet(h3).
//            MustGet(h4)
//  host := mux.NewHost(nil)
//  host.Add("abc.example.com", m1)
//  host.Add("?(?P<site>.*).example.com", m2) // 正则
//  http.ListenAndServe("8080", host)
type Host struct {
	mu           sync.Mutex
	errorHandler ErrorHandler

	entries      []*entry
	namedEntries map[string]*entry
}

// 新建Host实例。
// err为错误处理状态函数。
func NewHost(err ErrorHandler) *Host {
	if err == nil {
		err = defaultErrorHandler
	}

	return &Host{
		errorHandler: err,
		entries:      make([]*entry, 0, 2),
		namedEntries: make(map[string]*entry, 2),
	}
}

// 添加相应域名的处理函数。
// 若该域名已经存在，则返回错误信息。
// pattern，为域名信息，若以?开头，则表示这是个正则表达式匹配。
// h 当值为空时，触发panic。
func (host *Host) Add(pattern string, h http.Handler) error {
	if h == nil {
		panic("参数handler不能为空")
	}

	host.mu.Lock()
	defer host.mu.Unlock()

	_, found := host.namedEntries[pattern]
	if found {
		return fmt.Errorf("该表达式[%v]已经存在", pattern)
	}

	entry := newEntry(pattern, h)
	host.namedEntries[pattern] = entry
	host.entries = append(host.entries, entry)

	return nil
}

// implement http.Handler.ServeHTTP()
func (host *Host) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host.mu.Lock()
	defer host.mu.Unlock()

	for _, entry := range host.entries {
		if !entry.match(req.Host) {
			continue
		}

		ctx := GetContext(req)
		ctx.Set("domains", entry.getNamedCapture(req.Host))
		entry.handler.ServeHTTP(w, req)
		return
	}

	host.errorHandler(w, "没有找到与之匹配的主机名", 404)
}
