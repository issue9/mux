// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func (t *tester) resource(p string) *Resource {
	return t.router.Resource(p)
}

func TestResource(t *testing.T) {
	a := assert.New(t, false)
	test := newTester(t)
	h := test.resource("/h/1")
	a.NotNil(h)
	f := test.resource("/f/1")
	a.NotNil(f)

	h.Get(rest.BuildHandler(a, 201, "", nil))
	test.matchCode(http.MethodGet, "/h/1", 201)
	f.GetFunc(rest.BuildHandlerFunc(a, 201, "", nil))
	test.matchCode(http.MethodGet, "/f/1", 201)

	h.Post(rest.BuildHandler(a, 202, "", nil))
	test.matchCode(http.MethodPost, "/h/1", 202)
	f.PostFunc(rest.BuildHandlerFunc(a, 202, "", nil))
	test.matchCode(http.MethodPost, "/f/1", 202)

	h.Put(rest.BuildHandler(a, 203, "", nil))
	test.matchCode(http.MethodPut, "/h/1", 203)
	f.PutFunc(rest.BuildHandlerFunc(a, 203, "", nil))
	test.matchCode(http.MethodPut, "/f/1", 203)

	h.Patch(rest.BuildHandler(a, 204, "", nil))
	test.matchCode(http.MethodPatch, "/h/1", 204)
	f.PatchFunc(rest.BuildHandlerFunc(a, 204, "", nil))
	test.matchCode(http.MethodPatch, "/f/1", 204)

	h.Delete(rest.BuildHandler(a, 205, "", nil))
	test.matchCode(http.MethodDelete, "/h/1", 205)
	f.DeleteFunc(rest.BuildHandlerFunc(a, 205, "", nil))
	test.matchCode(http.MethodDelete, "/f/1", 205)

	// Any
	h = test.resource("/h/any")
	h.Any(rest.BuildHandler(a, 206, "", nil))
	test.matchCode(http.MethodGet, "/h/any", 206)
	test.matchCode(http.MethodPost, "/h/any", 206)
	test.matchCode(http.MethodPut, "/h/any", 206)
	test.matchCode(http.MethodPatch, "/h/any", 206)
	test.matchCode(http.MethodDelete, "/h/any", 206)
	test.matchCode(http.MethodTrace, "/h/any", 206)

	f = test.resource("/f/any")
	f.AnyFunc(rest.BuildHandlerFunc(a, 206, "", nil))
	test.matchCode(http.MethodGet, "/f/any", 206)
	test.matchCode(http.MethodPost, "/f/any", 206)
	test.matchCode(http.MethodPut, "/f/any", 206)
	test.matchCode(http.MethodPatch, "/f/any", 206)
	test.matchCode(http.MethodDelete, "/f/any", 206)
	test.matchCode(http.MethodTrace, "/f/any", 206)

	// remove
	f.Remove(http.MethodGet, http.MethodHead)
	test.matchCode(http.MethodGet, "/f/any", 405)
	test.matchCode(http.MethodDelete, "/f/any", 206)

	f.Clean()
	test.matchCode(http.MethodGet, "/f/any", 404)
	test.matchCode(http.MethodDelete, "/f/any", 404)
}

func TestRouter_Resource(t *testing.T) {
	a := assert.New(t, false)
	def := NewRouter("")
	a.NotNil(def)

	r1 := def.Resource("/abc/1")
	a.NotNil(r1)
	a.Equal(r1.Router(), def)
	a.Equal(r1.pattern, "/abc/1")

	r2 := def.Resource("/abc/1")
	a.NotNil(r2)
	a.False(r1 == r2) // 不是同一个 *Resource

	r2.Delete(rest.BuildHandler(a, 201, "", nil))
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodDelete, "/abc/1", nil)
	a.NotError(err).NotNil(r)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)
}

func TestPrefix_Resource(t *testing.T) {
	a := assert.New(t, false)

	def := NewRouter("")
	a.NotNil(def)

	p := def.Prefix("/p1")

	r1 := p.Resource("/abc/1")
	a.NotNil(r1)

	r1.Delete(rest.BuildHandler(a, 201, "", nil))
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodDelete, "/p1/abc/1", nil)
	a.NotError(err).NotNil(r)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t, false)
	def := NewRouter("", AllowedCORS)
	a.NotNil(def)

	// 非正则
	res := def.Resource("/api/v1")
	a.NotNil(res)
	url, err := res.URL(false, map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/v1")

	// 没有参数
	url, err = res.URL(false, nil)
	a.NotError(err).Equal(url, "/api/v1")

	res = def.Resource("/api//v1")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api//v1")

	// 正常的单个参数
	res = def.Resource("/api/{id:\\d+}/{path}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1/p1")

	// 类型不正确
	url, err = res.URL(false, map[string]string{"id": "xxx", "path": "p1"})
	a.NotError(err).Equal(url, "/api/xxx/p1")
	url, err = res.URL(true, map[string]string{"id": "xxx", "path": "p1"})
	a.Error(err).Empty(url)

	res = def.Resource("/api/{id:\\d+}//{path}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1//p1")

	// 多个参数
	res = def.Resource("/api/{action}/{id:\\d+}")
	a.NotNil(res)
	url, err = res.URL(false, map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")

	// 缺少参数
	url, err = res.URL(false, map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	url, err = res.URL(false, map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
}
