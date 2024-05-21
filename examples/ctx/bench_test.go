// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/mux/v9/routertest"
)

func BenchmarkRouter(b *testing.B) {
	h := HandlerFunc(func(c *CTX) {
		if _, err := c.W.Write([]byte(c.R.URL.Path)); err != nil {
			panic(err)
		}
	})

	t := routertest.NewTester[Handler](call, HandlerFunc(notFound), methodNotAllowedBuilder, optionsHandlerBuilder)
	t.Bench(b, h)
}
