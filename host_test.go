// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestHost(t *testing.T) {
	defHandler := func(w http.ResponseWriter, r *http.Request) bool {
		return true
	}

	defFunc := MatcherFunc(defHandler)

	fn := func(host string, hh *Host, wont bool) {
		r, err := http.NewRequest("GET", "", nil)
		assert.NotError(t, err)

		r.Host = host
		assert.Equal(t, r.Host, host)
		assert.Equal(t, hh.ServeHTTP2(nil, r), wont, "域名[%v]无法正确匹配", host)
	}

	h := NewHost(defFunc, "www.example.com")
	fn("www.example.com", h, true)
	fn("www.abc.com", h, false)

	h = NewHost(defFunc, "\\w+.example.com")
	fn("www.example.com", h, true)
	fn("api.example.com", h, true)
	fn("www.abc.com", h, false)
}

func TestHostDomains(t *testing.T) {
	a := assert.New(t)

	var domains map[string]string
	var ok bool
	defHandler := func(w http.ResponseWriter, r *http.Request) bool {
		tmp, found := GetContext(r).Get("domains")
		a.True(found)

		domains, ok = tmp.(map[string]string)
		a.True(ok)

		return true
	}

	defFunc := MatcherFunc(defHandler)

	// host: http.Request.Host的值
	fn := func(host string, hh *Host, wont map[string]string) {
		r, err := http.NewRequest("GET", "", nil)
		a.NotError(err)

		r.Host = host
		a.Equal(r.Host, host)

		a.True(hh.ServeHTTP2(nil, r), "域名[%v]无法正确匹配", host)

		a.Equal(domains, wont)
	}

	// 没有命名参数
	domains = map[string]string{}
	h := NewHost(defFunc, "\\w+.example.com")
	fn("www.example.com", h, map[string]string{})

	// 单个命名参数
	domains = map[string]string{}
	h = NewHost(defFunc, "(?P<city>\\w+).example.com")
	fn("bj.example.com", h, map[string]string{"city": "bj"})

	// 单个全名参数
	domains = map[string]string{}
	h = NewHost(defFunc, "(?P<city>[a-z]*)\\.(?P<prov>[a-z]*).example.com")
	fn("hz.zj.example.com", h, map[string]string{"city": "hz", "prov": "zj"})
}
