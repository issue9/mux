// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/interceptor"
	"github.com/issue9/mux/v5/internal/tree"
)

func TestMux_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	test := newTester(t, true, false)

	a.NotError(test.router.Handle("/posts/{path}.html", buildHandler(201)))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 201)

	a.NotError(test.router.Handle("/posts/{path:.+}.html", buildHandler(202)))
	test.matchTrue(http.MethodGet, "/posts/2017/1.html", 202)

	a.NotError(test.router.Handle("/posts/{id:digit}123", buildHandler(203)))
	test.matchTrue(http.MethodGet, "/posts/123123", 203)
}

// 测试匹配顺序是否正确
func TestRouter_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)

	test := newTester(t, true, false)
	a.NotError(test.router.GetFunc("/posts/{id}", buildHandlerFunc(203)))        // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildHandlerFunc(202)))   // f2
	a.NotError(test.router.GetFunc("/posts/1", buildHandlerFunc(201)))           // f1
	a.NotError(test.router.GetFunc("/posts/{id:[0-9]+}", buildHandlerFunc(199))) // f0 两个正则，后添加的永远匹配不到
	test.matchTrue(http.MethodGet, "/posts/1", 201)                              // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 202)                              // f1 正则路由
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                            // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                               // f3

	// interceptor
	test = newTester(t, true, false)
	a.NotError(interceptor.Register(interceptor.MatchDigit, "[0-9]+"))
	a.NotError(test.router.GetFunc("/posts/{id}", buildHandlerFunc(203)))        // f3
	a.NotError(test.router.GetFunc("/posts/{id:\\d+}", buildHandlerFunc(202)))   // f2 永远匹配不到
	a.NotError(test.router.GetFunc("/posts/1", buildHandlerFunc(201)))           // f1
	a.NotError(test.router.GetFunc("/posts/{id:[0-9]+}", buildHandlerFunc(210))) // f0 interceptor 权限比正则要高
	test.matchTrue(http.MethodGet, "/posts/1", 201)                              // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2", 210)                              // f1 interceptor
	test.matchTrue(http.MethodGet, "/posts/abc", 203)                            // f3 命名路由
	test.matchTrue(http.MethodGet, "/posts/", 203)                               // f3
	interceptor.Deregister("[0-9]+")

	test = newTester(t, true, false)
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\d+}", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/p1/{p1}/p2/{p2:\\w+}", buildHandlerFunc(202))) // f2
	test.matchTrue(http.MethodGet, "/p1/1/p2/1", 201)                               // f1
	test.matchTrue(http.MethodGet, "/p1/2/p2/s", 202)                               // f2

	test = newTester(t, true, false)
	a.NotError(test.router.GetFunc("/posts/{id}/{page}", buildHandlerFunc(202))) // f2
	a.NotError(test.router.GetFunc("/posts/{id}/1", buildHandlerFunc(201)))      // f1
	test.matchTrue(http.MethodGet, "/posts/1/1", 201)                            // f1 普通路由项完全匹配
	test.matchTrue(http.MethodGet, "/posts/2/5", 202)                            // f2 命名完全匹配

	test = newTester(t, true, false)
	a.NotError(test.router.GetFunc("/tags/{id}.html", buildHandlerFunc(201))) // f1
	a.NotError(test.router.GetFunc("/tags.html", buildHandlerFunc(202)))      // f2
	a.NotError(test.router.GetFunc("/{path}", buildHandlerFunc(203)))         // f3
	test.matchTrue(http.MethodGet, "/tags", 203)                              // f3 // 正好与 f1 的第一个节点匹配
	test.matchTrue(http.MethodGet, "/tags/1.html", 201)                       // f1
	test.matchTrue(http.MethodGet, "/tags.html", 202)                         // f2
}

func TestMethods(t *testing.T) {
	a := assert.New(t)
	a.Equal(Methods(), tree.Methods)
}

func TestIsWell(t *testing.T) {
	a := assert.New(t)

	a.NotError(IsWell("/{path"))
	a.NotError(IsWell("/path}"))
	a.Error(IsWell(""))
}
