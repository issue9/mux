// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"regexp"
)

// 用于匹配http.Request.Host的Handler
//  m1 := mux.NewMethod().
//            Get(h1).
//            Post(h2)
//  m2 := mux.NewMethod().
//            Get(h3).
//            Get(h4)
//  h1 := mux.NewHost(m1, "api.example.com")
//  h2 := mux.NewHost(m2, "www.example.com")
//  http.ListenAndServe("8080", NewMatches(h1, h2))
type Host struct {
	h        Matcher
	hostExpr *regexp.Regexp
}

// host参数为匹配的域名，可以是正则表达式。
func NewHost(handler Matcher, host string) *Host {
	return &Host{
		h:        handler,
		hostExpr: regexp.MustCompile(host),
	}
}

func (h *Host) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	if h.hostExpr.MatchString(r.Host) {
		// 分析域名。
		ctx := GetContext(r)
		ctx.Set("domains", parseCaptures(h.hostExpr, r.Host))
		return h.h.ServeHTTP2(w, r)
	}

	return false
}

func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP2(w, r)
}
