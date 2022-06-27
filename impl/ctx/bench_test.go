// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/mux/v6/routertest"
)

func BenchmarkRouter(b *testing.B) {
	h := HandlerFunc(func(c *CTX) {
		if _, err := c.W.Write([]byte(c.R.URL.Path)); err != nil {
			panic(err)
		}
	})

	t := routertest.NewTester(call, optionsHandlerBuilder)
	t.Bench(b, h)
}
