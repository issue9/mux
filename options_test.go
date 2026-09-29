// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v5"
	"github.com/issue9/assert/v5/rest"
	"github.com/issue9/mux/v10/internal/cors"
)

func TestOption(t *testing.T) {
	a := assert.New(t, false)

	r := newRouter(a, "def")
	a.NotNil(r)

	r = newRouter(a, "def2", WithCORS([]string{"https://example.com"}, nil, nil, 3600, false))
	a.NotNil(r).
		Equal(r.cors.Origins, []string{"https://example.com"}).
		Nil(r.cors.AllowHeaders).
		Equal(r.cors.MaxAge, 3600)

	r = newRouter(a, "def3", WithCORS([]string{"https://example.com"}, nil, nil, 0, true))
	a.NotNil(r)

	a.Panic(func() {
		r = newRouter(a, "def4", WithCORS([]string{"*"}, nil, nil, 0, true))
	})
}

func TestRecovery(t *testing.T) {
	a := assert.New(t, false)

	p := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("panic test") })

	// 未指定 Recovery

	router := newRouter(a, "def")
	a.NotNil(router).Nil(router.recoverFunc)
	router.Get("/path", p)
	a.Panic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
	})

	// WriterRecovery

	out := new(bytes.Buffer)
	router = newRouter(a, "def2", WithWriteRecovery(404, out))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)

	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Wait(time.Microsecond*500).
			Contains(out.String(), "panic test", out.String()).
			Contains(out.String(), "options_test.go:43", out.String()).
			Equal(w.Code, 404)
	})

	// LogRecovery

	out = new(bytes.Buffer)
	l := slog.New(slog.NewTextHandler(out, nil))
	router = newRouter(a, "def3", WithLogRecovery(405, l))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(405, w.Code)
		a.Contains(out.String(), "panic test\\n")         // 保证第一行是 panic 输出的信息
		a.Contains(out.String(), "TestRecovery.func1\\n") // 保证第二行是 panic 函数名
		a.Contains(out.String(), "options_test.go:43\\n") // 保证第三行是 panic 的行号
	})

	// StatusRecovery

	router = newRouter(a, "def4", WithStatusRecovery(406))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.NotPanic(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 406)
	})

	// 忽略 http.ErrAbortHandler

	p = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(http.ErrAbortHandler) })

	router = newRouter(a, "def5", WithStatusRecovery(406))
	a.NotNil(router).NotNil(router.recoverFunc)
	router.Get("/path", p)
	a.PanicValue(func() {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		router.ServeHTTP(w, r)
		a.Equal(w.Code, 200)
	}, http.ErrAbortHandler)
}

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t, false)

	o, err := buildOption()
	a.NotError(err).
		NotNil(o).
		NotNil(o.cors)

	// URLDomain

	o, err = buildOption(func(o *options) { o.pathPrefix = "https://example.com" })
	a.NotError(err).NotNil(o).Equal(o.pathPrefix, "https://example.com")

	o, err = buildOption(func(o *options) { o.pathPrefix = "https://example.com/" })
	a.NotError(err).NotNil(o).Equal(o.pathPrefix, "https://example.com")

	o, err = buildOption(func(o *options) { o.cors = &cors.CORS{AllowCredentials: true, Origins: []string{"*"}} })
	a.Error(err).Nil(o)
}
