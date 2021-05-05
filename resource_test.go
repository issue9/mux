// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func (t *tester) resource(p string) *Resource {
	return t.mux.Resource(p)
}

func TestResource(t *testing.T) {
	a := assert.New(t)
	test := newTester(t, false, true, false)
	h := test.resource("/h/1")
	a.NotNil(h)
	f := test.resource("/f/1")
	a.NotNil(f)

	h.Get(buildHandler(201))
	test.matchTrue(http.MethodGet, "/h/1", 201)
	f.GetFunc(buildFunc(201))
	test.matchTrue(http.MethodGet, "/f/1", 201)

	h.Post(buildHandler(202))
	test.matchTrue(http.MethodPost, "/h/1", 202)
	f.PostFunc(buildFunc(202))
	test.matchTrue(http.MethodPost, "/f/1", 202)

	h.Put(buildHandler(203))
	test.matchTrue(http.MethodPut, "/h/1", 203)
	f.PutFunc(buildFunc(203))
	test.matchTrue(http.MethodPut, "/f/1", 203)

	h.Patch(buildHandler(204))
	test.matchTrue(http.MethodPatch, "/h/1", 204)
	f.PatchFunc(buildFunc(204))
	test.matchTrue(http.MethodPatch, "/f/1", 204)

	h.Delete(buildHandler(205))
	test.matchTrue(http.MethodDelete, "/h/1", 205)
	f.DeleteFunc(buildFunc(205))
	test.matchTrue(http.MethodDelete, "/f/1", 205)

	// Any
	h = test.resource("/h/any")
	h.Any(buildHandler(206))
	test.matchTrue(http.MethodGet, "/h/any", 206)
	test.matchTrue(http.MethodPost, "/h/any", 206)
	test.matchTrue(http.MethodPut, "/h/any", 206)
	test.matchTrue(http.MethodPatch, "/h/any", 206)
	test.matchTrue(http.MethodDelete, "/h/any", 206)
	test.matchTrue(http.MethodTrace, "/h/any", 206)

	f = test.resource("/f/any")
	f.AnyFunc(buildFunc(206))
	test.matchTrue(http.MethodGet, "/f/any", 206)
	test.matchTrue(http.MethodPost, "/f/any", 206)
	test.matchTrue(http.MethodPut, "/f/any", 206)
	test.matchTrue(http.MethodPatch, "/f/any", 206)
	test.matchTrue(http.MethodDelete, "/f/any", 206)
	test.matchTrue(http.MethodTrace, "/f/any", 206)

	// remove
	f.Remove(http.MethodGet, http.MethodHead)
	test.matchTrue(http.MethodGet, "/f/any", 405)
	test.matchTrue(http.MethodDelete, "/f/any", 206)

	f.Clean()
	test.matchTrue(http.MethodGet, "/f/any", 404)
	test.matchTrue(http.MethodDelete, "/f/any", 404)

	// options
	f = test.resource("/f/options")
	f.Options("ABC")
	test.optionsTrue("/f/options", 200, "ABC")
}

func TestRouter_Resource(t *testing.T) {
	a := assert.New(t)
	mux := New(false, true, false, nil, nil, "", nil)
	a.NotNil(mux)

	r1 := mux.Resource("/abc/1")
	a.NotNil(r1)
	a.Equal(r1.Router(), mux.Router)
	a.Equal(r1.pattern, "/abc/1")

	r2 := mux.Resource("/abc/1")
	a.NotNil(r2)
	a.False(r1 == r2) // 不是同一个 *Resource

	r2.Delete(buildHandler(201))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/abc/1", nil)
	mux.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)
}

func TestPrefix_Resource(t *testing.T) {
	a := assert.New(t)

	router := Default()
	a.NotNil(router)
	p := router.Prefix("/p1")

	r1 := p.Resource("/abc/1")
	a.NotNil(r1)

	r1.Delete(buildHandler(201))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/p1/abc/1", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201)
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t)
	mux := New(false, true, false, nil, nil, "", nil)
	a.NotNil(mux)

	// 非正则
	res := mux.Resource("/api/v1")
	a.NotNil(res)
	url, err := res.URL(map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/v1")

	// 正常的单个参数
	res = mux.Resource("/api/{id:\\d+}/{path}")
	a.NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "/api/1/p1")

	// 不对正则参数做类型校验，可以生成不符合正则要求的路径。
	// 方便特殊情况下使用。
	url, err = res.URL(map[string]string{"id": "xxx", "path": "p1"})
	a.NotError(err).Equal(url, "/api/xxx/p1")

	// 多个参数
	res = mux.Resource("/api/{action}/{id:\\d+}")
	a.NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
	// 缺少参数
	url, err = res.URL(map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	url, err = res.URL(map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
}
