// SPDX-License-Identifier: MIT

package route

import "net/http"

type headResponse struct {
	http.ResponseWriter
}

func (resp *headResponse) Write([]byte) (int, error) { return 0, nil }
