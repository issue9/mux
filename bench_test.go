// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
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

func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b, false)

	paths := []string{
		"",
		"/api//",
		"/api////users/1",
		"//api/users/1",
		"api///users////1",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api//../",
		"/api/..//../",
		"/api../",
		"api../",
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.URL.Path = paths[i%len(paths)]
		_, req := CleanPath(w, r)
		a.True(len(req.URL.Path) > 0)
	}
}
