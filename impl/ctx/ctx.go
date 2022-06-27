// SPDX-License-Identifier: MIT

package ctx

import (
	"net/http"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/params"
)

type (
	CTX struct {
		R *http.Request
		W http.ResponseWriter
		P mux.Params
	}

	Handler interface {
		Handle(*CTX)
	}

	HandlerFunc func(ctx *CTX)
)

func (f HandlerFunc) Handle(c *CTX) { f(c) }

func call(w http.ResponseWriter, r *http.Request, ps mux.Params, h Handler) {
	h.Handle(&CTX{R: r, W: w, P: ps})
}

func optionsHandlerBuilder(p params.Node) Handler {
	return HandlerFunc(func(ctx *CTX) {
		ctx.W.Header().Set("allow", p.AllowHeader())
	})
}
