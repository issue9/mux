// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func BenchmarkServeFile(b *testing.B) {
	fsys := os.DirFS("./")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/assets/mux.go", nil)
		ServeFile(fsys, "mux.go", "", w, r)
	}
}
