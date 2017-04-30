// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"strings"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

// Prefix 封装了 Mux，使所有添加的路由项的匹配模式都带上指定的路径前缀。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
//
// 当指定多个相同的 Prefix 时，调用其中的一个 Clean 也将会清除其它的：
//  p1 := srv.Prefix("prefix")
//  p2 := srv.Prefix("prefix")
//  p2.Clean() 将同时清除 p1 的内容，因为有相同的前缀
type Prefix struct {
	mux    *Mux
	prefix string
}

// Options 手动指定 OPTIONS 请求方法的值。
//
// 若无特殊需求，不用调用些方法，系统会自动计算符合当前路由的请求方法列表。
func (p *Prefix) Options(pattern string, allow string) *Prefix {
	p.mux.Options(p.prefix+pattern, allow)
	return p
}

// Add 相当于 Mux.Add(prefix+pattern, h, "POST"...) 的简易写法
func (p *Prefix) Add(pattern string, h http.Handler, methods ...string) error {
	return p.mux.Add(p.prefix+pattern, h, methods...)
}

func (p *Prefix) add(pattern string, h http.Handler, methods ...string) *Prefix {
	if err := p.mux.Add(p.prefix+pattern, h, methods...); err != nil {
		panic(err)
	}

	return p
}

// Get 相当于 Mux.Get(prefix+pattern, h) 的简易写法
func (p *Prefix) Get(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, http.MethodGet)
}

// Post 相当于 Mux.Post(prefix+pattern, h) 的简易写法
func (p *Prefix) Post(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, http.MethodPost)
}

// Delete 相当于 Mux.Delete(prefix+pattern, h)的简易写法
func (p *Prefix) Delete(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, http.MethodDelete)
}

// Put 相当于 Mux.Put(prefix+pattern, h) 的简易写法
func (p *Prefix) Put(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, http.MethodPut)
}

// Patch 相当于 Mux.Patch(prefix+pattern, h) 的简易写法
func (p *Prefix) Patch(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, http.MethodPatch)
}

// Any 相当于 Mux.Any(prefix+pattern, h) 的简易写法
func (p *Prefix) Any(pattern string, h http.Handler) *Prefix {
	return p.add(pattern, h, method.Default...)
}

// AddFunc 功能同 Mux.AddFunc(prefix+pattern, fun, ...)
func (p *Prefix) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) error {
	return p.mux.AddFunc(p.prefix+pattern, fun, methods...)
}

func (p *Prefix) addFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *Prefix {
	if err := p.mux.AddFunc(p.prefix+pattern, fun, methods...); err != nil {
		panic(err)
	}
	return p
}

// GetFunc 相当于 Mux.GetFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, http.MethodGet)
}

// PutFunc 相当于 Mux.PutFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, http.MethodPut)
}

// PostFunc 相当 于Mux.PostFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, http.MethodPost)
}

// DeleteFunc 相当于 Mux.DeleteFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, http.MethodDelete)
}

// PatchFunc 相当于 Mux.PatchFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, http.MethodPatch)
}

// AnyFunc 相当于 Mux.AnyFunc(prefix+pattern, func) 的简易写法
func (p *Prefix) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *Prefix {
	return p.addFunc(pattern, fun, method.Default...)
}

// Remove 删除指定匹配模式的路由项
func (p *Prefix) Remove(pattern string, methods ...string) *Prefix {
	p.mux.Remove(p.prefix+pattern, methods...)
	return p
}

// Clean 清除所有以 Prefix.prefix 开头的 Entry。
//
// NOTE: 若 mux 中也有同样开头的 Entry，也照样会被删除。
func (p *Prefix) Clean() *Prefix {
	p.mux.mu.Lock()
	defer p.mux.mu.Unlock()

	for item := p.mux.entries.Front(); item != nil; {
		curr := item
		item = item.Next() // 提前记录下个元素，因为 item 有可能被删除

		ety := curr.Value.(entry.Entry)
		pattern := ety.Pattern()
		if strings.HasPrefix(pattern, p.prefix) {
			if empty := ety.Remove(method.Supported...); empty {
				p.mux.entries.Remove(curr)
			}
		}
	} // end for

	return p
}

// Prefix 创建一个路由组，该组中添加的路由项，都会带上前缀 prefix
// prefix 前缀字符串，所有从 Prefix 中声明的路由都将包含此前缀。
//  p := g.Prefix("/api")
//  p.Get("/users")  // 相当于 g.Get("/api/users")
//  p.Get("/user/1") // 相当于 g.Get("/api/user/1")
func (p *Prefix) Prefix(prefix string) *Prefix {
	return &Prefix{
		mux:    p.mux,
		prefix: p.prefix + prefix,
	}
}

// Prefix 创建一个路由组，该组中添加的路由项，都会带上前缀 prefix
// prefix 前缀字符串，所有从 Prefix 中声明的路由都将包含此前缀。
//  p := srv.Prefix("/api")
//  p.Get("/users")  // 相当于 srv.Get("/api/users")
//  p.Get("/user/1") // 相当于 srv.Get("/api/user/1")
func (mux *Mux) Prefix(prefix string) *Prefix {
	return &Prefix{
		mux:    mux,
		prefix: prefix,
	}
}

// Mux 返回与当前关联的 *Mux 实例
func (p *Prefix) Mux() *Mux {
	return p.mux
}
