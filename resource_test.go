// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func (t *tester) resource(p string) *Resource {
	r := t.mux.Resource(p)
	t.a.NotNil(r)
	return r
}

func TestResource(t *testing.T) {
	a := assert.New(t)
	test := newTester(a, false, true, false)
	h := test.resource("/h/1")
	f := test.resource("/f/1")

	h.Get(buildHandler(1))
	test.matchTrue(http.MethodGet, "/h/1", 1)
	f.GetFunc(buildFunc(1))
	test.matchTrue(http.MethodGet, "/f/1", 1)

	h.Post(buildHandler(2))
	test.matchTrue(http.MethodPost, "/h/1", 2)
	f.PostFunc(buildFunc(2))
	test.matchTrue(http.MethodPost, "/f/1", 2)

	h.Put(buildHandler(3))
	test.matchTrue(http.MethodPut, "/h/1", 3)
	f.PutFunc(buildFunc(3))
	test.matchTrue(http.MethodPut, "/f/1", 3)

	h.Patch(buildHandler(4))
	test.matchTrue(http.MethodPatch, "/h/1", 4)
	f.PatchFunc(buildFunc(4))
	test.matchTrue(http.MethodPatch, "/f/1", 4)

	h.Delete(buildHandler(5))
	test.matchTrue(http.MethodDelete, "/h/1", 5)
	f.DeleteFunc(buildFunc(5))
	test.matchTrue(http.MethodDelete, "/f/1", 5)

	// Any
	h = test.resource("/h/any")
	h.Any(buildHandler(6))
	test.matchTrue(http.MethodGet, "/h/any", 6)
	test.matchTrue(http.MethodPost, "/h/any", 6)
	test.matchTrue(http.MethodPut, "/h/any", 6)
	test.matchTrue(http.MethodPatch, "/h/any", 6)
	test.matchTrue(http.MethodDelete, "/h/any", 6)
	test.matchTrue(http.MethodTrace, "/h/any", 6)

	f = test.resource("/f/any")
	f.AnyFunc(buildFunc(6))
	test.matchTrue(http.MethodGet, "/f/any", 6)
	test.matchTrue(http.MethodPost, "/f/any", 6)
	test.matchTrue(http.MethodPut, "/f/any", 6)
	test.matchTrue(http.MethodPatch, "/f/any", 6)
	test.matchTrue(http.MethodDelete, "/f/any", 6)
	test.matchTrue(http.MethodTrace, "/f/any", 6)
}

func TestMux_Resource(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, true, false, nil, nil)
	a.NotNil(srvmux)

	r1 := srvmux.Resource("/abc/1")
	a.NotNil(r1)
	a.Equal(r1.Mux(), srvmux)
	a.Equal(r1.pattern, "/abc/1")

	r2 := srvmux.Resource("/abc/1")
	a.NotNil(r2)
	a.False(r1 == r2) // 不是同一个 *Resource
}

func TestResource_Name_URL(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, true, false, nil, nil)
	a.NotNil(srvmux)

	// 非正则
	res := srvmux.Resource("/api/v1")
	a.NotNil(res)
	url, err := res.URL(map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/v1")

	// 正常的单个参数
	res = srvmux.Resource("/api/{id:\\d+}/{path}")
	a.NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1/p1")

	// 多个参数
	res = srvmux.Resource("/api/{action}/{id:\\d+}")
	a.NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
	// 缺少参数
	url, err = res.URL(map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	a.NotError(res.Name("action"))
	url, err = res.Mux().URL("action", map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
}
