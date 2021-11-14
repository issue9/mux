// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
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
	a := assert.New(t)
	def := NewRouter("", AllowedCORS)
	a.NotNil(def)

	p := def.Prefix("/abc")
	a.Equal(p.prefix, "/abc")
	a.Equal(p.Router(), def)

	p = def.Prefix("")
	a.Equal(p.prefix, "")
}

func TestPrefix_Prefix(t *testing.T) {
	a := assert.New(t)
	def := NewRouter("", AllowedCORS)
	a.NotNil(def)

	p := def.Prefix("/abc")
	pp := p.Prefix("/def")
	a.Equal(pp.prefix, "/abc/def")
	a.Equal(p.Router(), def)

	p = def.Prefix("")
	pp = p.Prefix("/abc")
	a.Equal(pp.prefix, "/abc")
}

func TestPrefix_URL(t *testing.T) {
	a := assert.New(t)
	def := NewRouter("", AllowedCORS, URL(false, "https://example.com"))
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
