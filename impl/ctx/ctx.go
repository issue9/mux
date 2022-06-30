// SPDX-License-Identifier: MIT

package ctx

import (
	"net/http"

	"github.com/issue9/mux/v6/types"
)

type (
	CTX struct {
		R *http.Request
		W http.ResponseWriter
		P types.Params
	}

	Handler interface {
		Handle(*CTX)
	}

	HandlerFunc func(ctx *CTX)
)

func (f HandlerFunc) Handle(c *CTX) { f(c) }

func call(w http.ResponseWriter, r *http.Request, ps types.Params, h Handler) {
	h.Handle(&CTX{R: r, W: w, P: ps})
}

func optionsHandlerBuilder(p types.Node) Handler {
	return HandlerFunc(func(ctx *CTX) {
		ctx.W.Header().Set("allow", p.AllowHeader())
	})
}
