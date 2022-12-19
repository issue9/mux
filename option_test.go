// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"

	"github.com/issue9/mux/v7/internal/tree"
	"github.com/issue9/mux/v7/types"
)

func newRouter(name string, o ...Option) *RouterOf[http.Handler] {
	callFunc := func(w http.ResponseWriter, r *http.Request, p types.Route, h http.Handler) {
		h.ServeHTTP(w, r)
	}
	m := tree.BuildTestNodeHandlerFunc(http.StatusMethodNotAllowed)
	opt := tree.BuildTestNodeHandlerFunc(http.StatusOK)
	return NewRouterOf(name, callFunc, http.NotFoundHandler(), m, opt, o...)
}

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := newRouter("")
	a.NotNil(r)

	r = newRouter("", CORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.cors.Origins, []string{"https://example.com"}).
		Nil(r.cors.AllowHeaders).
		Equal(r.cors.MaxAge, 3600)

	r = newRouter("", CORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = newRouter("", CORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("panic test") })

	router := newRouter("")
	a.NotNil(router).Nil(router.recoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery

	out := new(bytes.Buffer)
	router = newRouter("", WriterRecovery(404, out))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Contains(out.String(), "panic test").
			Equal(w.Code, 404)
	})

	// LogRecovery

	out = new(bytes.Buffer)
	l := log.New(out, "log:", 0)
	router = newRouter("", LogRecovery(405, l))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code)
		lines := strings.Split(out.String(), "\n")
		a.Contains(lines[0], "panic test")                                 // 保证第一行是 panic 输出的信息
		a.Contains(lines[1], "TestRecovery.func1")                         // 保证第二行是 panic 函数名
		a.True(strings.HasSuffix(lines[2], "option_test.go:52"), lines[2]) // 保证第三行是 panic 的行号
	})

	// StatusRecovery

	router = newRouter("", StatusRecovery(406))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 406)
	})
}
