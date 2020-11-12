// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v3/internal/handlers"
	"github.com/issue9/mux/v3/params"
)

// Matcher 验证一个请求是否符合要求
type Matcher interface {
	Match(*http.Request) bool
}

// Hosts 限定域名的匹配工具
type Hosts struct {
	domains   []string // 域名列表
	wildcards []string // 泛域名列表，只保存 * 之后的部分内容
}

// NewHosts 声明新的 Hosts 实例
func NewHosts(domain ...string) *Hosts {
	h := &Hosts{
		domains:   make([]string, 0, len(domain)),
		wildcards: make([]string, 0, len(domain)),
	}

	h.Add(domain...)

	return h
}

// NewMux 添加子路由组
//
// 该路由只有符合 matcher 的要求才会进入，其它与 Mux 功能相同。
//
// name 表示不该路由组的名称，需要唯一，否则返回 false；
func (mux *Mux) NewMux(name string, matcher Matcher) (*Mux, bool) {
	if mux.routers == nil {
		mux.routers = make([]*Mux, 0, 5)
	}

	if sliceutil.Count(mux.routers, func(i int) bool { return mux.routers[i].name == name }) > 0 {
		return nil, false
	}

	m := New(mux.disableOptions, mux.disableHead, mux.skipCleanPath, mux.notFound, mux.methodNotAllowed)
	m.name = name
	m.matcher = matcher
	mux.routers = append(mux.routers, m)
	return m, true
}

func (mux *Mux) match(r *http.Request) (*handlers.Handlers, params.Params) {
	path := r.URL.Path
	if !mux.skipCleanPath {
		path = cleanPath(path)
	}

	for _, m := range mux.routers {
		if hs, ps := m.match(r); hs != nil {
			return hs, ps
		}
	}

	if mux.matcher == nil || mux.matcher.Match(r) {
		return mux.tree.Handler(path)
	}
	return nil, nil
}

// Match Matcher.Match
func (hs *Hosts) Match(r *http.Request) bool {
	hostname := r.URL.Hostname()
	for _, domain := range hs.domains {
		if domain == hostname {
			return true
		}
	}

	for _, wildcard := range hs.wildcards {
		if strings.HasSuffix(hostname, wildcard) {
			return true
		}
	}

	return false
}

// Add 添加新的域名
//
// domain 可以是泛域名，比如 *.example.com，但不能是 s1.*.example.com。
//
// NOTE: 重复的值不会重复添加。
func (hs *Hosts) Add(domain ...string) {
	for _, d := range domain {
		switch {
		case strings.HasPrefix(d, "*."):
			d = d[1:] // 保留 . 符号
			if sliceutil.Count(hs.wildcards, func(i int) bool { return d == hs.wildcards[i] }) <= 0 {
				hs.wildcards = append(hs.wildcards, d)
			}
		default:
			if sliceutil.Count(hs.domains, func(i int) bool { return d == hs.domains[i] }) <= 0 {
				hs.domains = append(hs.domains, d)
			}
		}
	}
}

// Delete 删除域名
//
// NOTE: 如果不存在，则不作任何改变。
func (hs *Hosts) Delete(domain string) {
	switch {
	case strings.HasPrefix(domain, "*."):
		size := sliceutil.Delete(hs.wildcards, func(i int) bool { return hs.wildcards[i] == domain[1:] })
		hs.wildcards = hs.wildcards[:size]
	default:
		size := sliceutil.Delete(hs.domains, func(i int) bool { return hs.domains[i] == domain })
		hs.domains = hs.domains[:size]
	}
}
