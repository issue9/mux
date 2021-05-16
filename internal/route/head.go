// SPDX-License-Identifier: MIT

package route

import "net/http"

type headResponse struct {
	http.ResponseWriter
}

func (resp *headResponse) Write([]byte) (int, error) { return 0, nil }

func (r *Route) headServeHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(&headResponse{ResponseWriter: w}, r)
	})
}
