// SPDX-License-Identifier: MIT

package mux

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7/internal/tree"
)

func BenchmarkServeFile(b *testing.B) {
	fsys := os.DirFS("./")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/assets/mux.go", nil)
		err := ServeFile(fsys, "mux.go", "", w, r)
		if err != nil {
			b.Errorf("测试出错，返回了以下错误 %s", err)
		}
	}
}

func TestMethods(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(Methods(), tree.Methods)
}

func TestCheckSyntax(t *testing.T) {
	a := assert.New(t, false)

	a.NotError(CheckSyntax("/{path"))
	a.NotError(CheckSyntax("/path}"))
	a.Error(CheckSyntax(""))
}

func TestURL(t *testing.T) {
	a := assert.New(t, false)

	url, err := URL("/posts/{id:}", map[string]string{"id": "100"})
	a.NotError(err).Equal(url, "/posts/100")

	url, err = URL("/posts/{id:}", nil)
	a.NotError(err).Equal(url, "/posts/{id:}")

	url, err = URL("/posts/{id:}", map[string]string{"other-": "id"})
	a.Error(err).Empty(url)

	url, err = URL("/posts/{id:\\\\d+}/author/{page}/", map[string]string{"id": "100", "page": "200"})
	a.NotError(err).Equal(url, "/posts/100/author/200/")
}

func TestTrace(t *testing.T) {
	a := assert.New(t, false)

	r := rest.NewRequest(a, http.MethodTrace, "/path").Body([]byte("<body>")).Request()
	w := httptest.NewRecorder()
	a.NotError(Trace(w, r, false))
	body := w.Body.String()
	a.Contains(body, "/path").
		NotContains(body, "body").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)

	w = httptest.NewRecorder()
	a.NotError(Trace(w, r, true))
	body = w.Body.String()
	a.Contains(body, "/path").
		Contains(body, "&lt;body&gt;").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)
}

func TestServeFile(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("./")

	w := httptest.NewRecorder()
	r := rest.Get(a, "/assets/").Request()
	a.NotError(ServeFile(fsys, "", "go.mod", w, r))
	a.Contains(w.Body.String(), "module github.com/issue9/mux")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.NotError(ServeFile(fsys, "types/types.go", "", w, r))
	a.NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.ErrorIs(ServeFile(fsys, "types/", "", w, r), fs.ErrNotExist)
	a.Empty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.ErrorIs(ServeFile(fsys, "not-exists", "", w, r), fs.ErrNotExist)
	a.Empty(w.Body.String())
}

func TestDebug(t *testing.T) {
	a := assert.New(t, false)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/vars", w, r))
	a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/pprof/cmdline", w, r))
	a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	//w = httptest.NewRecorder()
	//r = rest.Get(a, "/path").Query("seconds", "10").Request()
	//a.NotError(Debug("/pprof/profile", w, r))
	//a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/pprof/symbol", w, r))
	a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/pprof/trace", w, r))
	a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	// pprof.Index
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/pprof/heap", w, r))
	a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/", w, r))
	a.Equal(w.Code, http.StatusOK)
	a.Contains(w.Body.Bytes(), debugHtml)

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Query("seconds", "10").Request()
	a.NotError(Debug("/not-exits", w, r))
	a.Equal(w.Code, http.StatusNotFound)
}
