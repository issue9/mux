// SPDX-License-Identifier: MIT

package group

import (
	"net/http"

	"github.com/issue9/mux/v5"
)

type Group = GroupOf[http.Handler]

func New(ms []mux.MiddlewareFunc, o ...mux.Option) *Group {
	return NewOf[http.Handler](mux.DefaultBuildHandlerFunc, ms, o...)
}
