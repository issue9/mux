// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v3/group"
	"github.com/issue9/mux/v3/internal/handlers"
	"github.com/issue9/mux/v3/params"
)

// New 添加子路由组
//
// 该路由只有符合 group.Matcher 的要求才会进入，其它与 Mux 功能相同。
//
// name 表示该路由组的名称，需要唯一，否则返回 false；
func (mux *Mux) New(name string, matcher group.Matcher) (*Mux, bool) {
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
	for _, m := range mux.routers {
		if hs, ps := m.match(r); hs != nil {
			return hs, ps
		}
	}

	if mux.matcher == nil || mux.matcher.Match(r) {
		return mux.tree.Handler(r.URL.Path)
	}
	return nil, nil
}
