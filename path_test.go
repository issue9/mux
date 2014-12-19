// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestPath(t *testing.T) {
	defHandler := func(w http.ResponseWriter, r *http.Request) bool {
		return true
	}

	defFunc := MatcherFunc(defHandler)

	fn := func(pattern string, p *Path, wont bool) {
		r, err := http.NewRequest("GET", pattern, nil)
		assert.NotError(t, err)
		assert.Equal(t, p.ServeHTTP2(nil, r), wont)
	}

	p := NewPath(defFunc, "/api")
	fn("/api", p, true)
	fn("/api/v1", p, true)

	p = NewPath(defFunc, "/api/v(\\d+)")
	fn("/api", p, false)
	fn("/api/v1", p, true)
	fn("/api/v1/post/1", p, true)
}

func TestPathParams(t *testing.T) {
	a := assert.New(t)

	var params map[string]string
	var ok bool
	defHandler := func(w http.ResponseWriter, r *http.Request) bool {
		tmp, found := GetContext(r).Get("params")
		a.True(found)

		params, ok = tmp.(map[string]string)
		a.True(ok)

		return true
	}

	defFunc := MatcherFunc(defHandler)

	fn := func(pattern string, p *Path, wont map[string]string) {
		r, err := http.NewRequest("GET", pattern, nil)
		a.NotError(err)
		a.True(p.ServeHTTP2(nil, r))
		a.Equal(params, wont)
	}

	// 无命名参数
	params = map[string]string{}
	p := NewPath(defFunc, "/api/v(\\d+)")
	fn("/api/v1", p, map[string]string{})
	fn("/api/v1/post/1", p, map[string]string{})

	// 单个命名参数
	params = map[string]string{}
	p = NewPath(defFunc, "/api/v(?P<version>\\d+)")
	fn("/api/v1", p, map[string]string{"version": "1"})

	// 多个命名参数
	params = map[string]string{}
	p = NewPath(defFunc, "/api/v(?P<version>\\d+)/(?P<action>\\w*)")
	fn("/api/v1/login", p, map[string]string{"version": "1", "action": "login"})
	fn("/api/v1/", p, map[string]string{"version": "1", "action": ""})
}
