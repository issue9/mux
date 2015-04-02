// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/issue9/context"
)

// 用于匹配域名的http.Handler
//
//  m1 := mux.NewMethod().
//            MustGet(h1).
//            MustPost(h2)
//  m2 := mux.NewMethod().
//            MustGet(h3).
//            MustGet(h4)
//  host := mux.NewHost()
//  host.Add("abc.example.com", m1)
//  host.Add("?(?P<site>.*).example.com", m2) // 正则
//  http.ListenAndServe("8080", host)
type Host struct {
	mu           sync.Mutex
	entries      []*entry
	namedEntries map[string]*entry
}

// 新建Host实例。
func NewHost() *Host {
	return &Host{
		entries:      make([]*entry, 0, 1),
		namedEntries: make(map[string]*entry, 1),
	}
}

// 添加相应域名的处理函数。
// 若该域名已经存在，则返回错误信息。
// pattern，为域名信息，若以?开头，则表示这是个正则表达式匹配。
// 当h值为空时，返回错误信息。
func (host *Host) Add(pattern string, h http.Handler) error {
	if h == nil {
		return errors.New("参数h不能为空值")
	}

	host.mu.Lock()
	defer host.mu.Unlock()

	_, found := host.namedEntries[pattern]
	if found {
		return fmt.Errorf("Add:该表达式[%v]已经存在", pattern)
	}

	entry, err := newEntry(pattern, h)
	if err != nil {
		return err
	}
	host.namedEntries[pattern] = entry
	host.entries = append(host.entries, entry)

	return nil
}

// 等同host.Add，但第二个参数为一个函数。
func (host *Host) AddFunc(pattern string, h func(http.ResponseWriter, *http.Request)) error {
	return host.Add(pattern, http.HandlerFunc(h))
}

// implement http.Handler.ServeHTTP()
func (host *Host) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	host.mu.Lock()
	defer host.mu.Unlock()

	for _, entry := range host.entries {
		if ok, mapped := entry.match(req.Host); ok {
			ctx := context.Get(req)
			ctx.Set("domains", mapped)
			entry.handler.ServeHTTP(w, req)
			return
		}
	}

	panic(fmt.Sprintf("没有找到与之匹配的主机名:[%v]", req.Host))
}
