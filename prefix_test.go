// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func (t *tester) prefix(p string) *Prefix {
	return t.router.Prefix(p)
}

func TestPrefix(t *testing.T) {
	test := newTester(t)
	p := test.prefix("/p")

	p.Get("/h/1", rest.BuildHandler(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/p/h/1", 201)
	p.GetFunc("/f/1", rest.BuildHandlerFunc(test.a, 201, "", nil))
	test.matchCode(http.MethodGet, "/p/f/1", 201)

	p.Post("/h/1", rest.BuildHandler(test.a, 202, "", nil))
	test.matchCode(http.MethodPost, "/p/h/1", 202)
	p.PostFunc("/f/1", rest.BuildHandlerFunc(test.a, 202, "", nil))
	test.matchCode(http.MethodPost, "/p/f/1", 202)

	p.Put("/h/1", rest.BuildHandler(test.a, 203, "", nil))
	test.matchCode(http.MethodPut, "/p/h/1", 203)
	p.PutFunc("/f/1", rest.BuildHandlerFunc(test.a, 203, "", nil))
	test.matchCode(http.MethodPut, "/p/f/1", 203)

	p.Patch("/h/1", rest.BuildHandler(test.a, 204, "", nil))
	test.matchCode(http.MethodPatch, "/p/h/1", 204)
	p.PatchFunc("/f/1", rest.BuildHandlerFunc(test.a, 204, "", nil))
	test.matchCode(http.MethodPatch, "/p/f/1", 204)

	p.Delete("/h/1", rest.BuildHandler(test.a, 205, "", nil))
	test.matchCode(http.MethodDelete, "/p/h/1", 205)
	p.DeleteFunc("/f/1", rest.BuildHandlerFunc(test.a, 205, "", nil))
	test.matchCode(http.MethodDelete, "/p/f/1", 205)

	// Any
	p.Any("/h/any", rest.BuildHandler(test.a, 206, "", nil))
	test.matchCode(http.MethodGet, "/p/h/any", 206)
	test.matchCode(http.MethodPost, "/p/h/any", 206)
	test.matchCode(http.MethodPut, "/p/h/any", 206)
	test.matchCode(http.MethodPatch, "/p/h/any", 206)
	test.matchCode(http.MethodDelete, "/p/h/any", 206)
	test.matchCode(http.MethodTrace, "/p/h/any", 206)

	p.AnyFunc("/f/any", rest.BuildHandlerFunc(test.a, 206, "", nil))
	test.matchCode(http.MethodGet, "/p/f/any", 206)
	test.matchCode(http.MethodPost, "/p/f/any", 206)
	test.matchCode(http.MethodPut, "/p/f/any", 206)
	test.matchCode(http.MethodPatch, "/p/f/any", 206)
	test.matchCode(http.MethodDelete, "/p/f/any", 206)
	test.matchCode(http.MethodTrace, "/p/f/any", 206)

	// remove
	p.Remove("/f/any", http.MethodDelete, http.MethodGet)
	test.matchCode(http.MethodGet, "/p/f/any", 405)   // 已经删除
	test.matchCode(http.MethodTrace, "/p/f/any", 206) // 未删除

	// clean
	p.Clean()
	test.matchCode(http.MethodTrace, "/p/f/any", 404)
	test.matchOptions("/p/h/1", 404, "")
	test.matchCode(http.MethodDelete, "/p/f/1", 404)
}

func TestRouter_Prefix(t *testing.T) {
	a := assert.New(t, false)
	def := NewRouter("", AllowedCORS)
	a.NotNil(def)

	p := def.Prefix("/abc")
	a.Equal(p.prefix, "/abc")
	a.Equal(p.Router(), def)

	p = def.Prefix("")
	a.Equal(p.prefix, "")
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t, false)
	def := NewRouter("", AllowedCORS)
	a.NotNil(def)

	p := def.Prefix("/abc", buildMiddleware(a, "p1"), buildMiddleware(a, "p2"))
	pp := p.Prefix("/def", buildMiddleware(a, "pp1"), buildMiddleware(a, "pp2"))
	a.Equal(pp.prefix, "/abc/def")
	a.Equal(p.Router(), def)
	pp.Delete("", rest.BuildHandler(a, 201, "-201-", nil))

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodDelete, "/abc/def", nil)
	a.NotError(err).NotNil(r)
	def.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 201).
		Equal(w.Body.String(), "-201-p1p2pp1pp2")

	p = def.Prefix("")
	pp = p.Prefix("/abc")
	a.Equal(pp.prefix, "/abc")
}

func TestPrefix_URL(t *testing.T) {
	a := assert.New(t, false)
	def := NewRouter("", AllowedCORS, URLDomain("https://example.com"))
	a.NotNil(def)

	// 非正则
	p := def.Prefix("/api")
	a.NotNil(p)
	url, err := p.URL(false, "/v1", map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "https://example.com/api/v1")

	p = def.Prefix("//api")
	a.NotNil(p)
	url, err = p.URL(false, "/v1", map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "https://example.com//api/v1")

	// 正常的单个参数
	p = def.Prefix("/api")
	a.NotNil(p)
	url, err = p.URL(false, "/{id:\\d+}/{path}", map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "https://example.com/api/1/p1")

	url, err = p.URL(false, "/{id:\\d+}///{path}", map[string]string{"id": "1", "path": "p1"})
	a.NotError(err).Equal(url, "https://example.com/api/1///p1")

	// 多个参数
	p = def.Prefix("/api")
	a.NotNil(p)
	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "https://example.com/api/blog/1")

	// 缺少参数
	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")

	url, err = p.URL(false, "/{action}/{id:\\d+}", map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "https://example.com/api/blog/1")
}
