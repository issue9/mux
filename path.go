// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"regexp"
)

// 以正则表达式匹配http.Request.URL.Path的Handler
type Path struct {
	m        Matcher
	pathExpr *regexp.Regexp
}

// NewPath新建一个Path实例。
// pattern用于匹配http.Request.URL.Path的正则表达式，可以用命名表达式。
func NewPath(matcher Matcher, pattern string) *Path {
	return &Path{
		m:        matcher,
		pathExpr: regexp.MustCompile(pattern),
	}
}

func (p *Path) ServeHTTP2(w http.ResponseWriter, r *http.Request) bool {
	if !p.pathExpr.MatchString(r.URL.Path) {
		return false
	}

	// 捕获命名项，并保存到context中
	ctx := GetContext(r)
	ctx.Set("params", parseCaptures(p.pathExpr, r.URL.Path))
	return p.m.ServeHTTP2(w, r)
}

func (p *Path) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ServeHTTP2(w, r)
}
