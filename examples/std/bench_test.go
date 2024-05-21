// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package std

import (
	"net/http"
	"testing"

	"github.com/issue9/mux/v9/routertest"
)

func BenchmarkRouter(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.URL.Path)); err != nil {
			panic(err)
		}
	}

	t := routertest.NewTester(call, http.NotFoundHandler(), methodNotAllowedBuilder, optionsHandlerBuilder)
	t.Bench(b, http.HandlerFunc(h))
}
