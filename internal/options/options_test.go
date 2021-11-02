// SPDX-License-Identifier: MIT

package options

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestOptions_Sanitize(t *testing.T) {
	a := assert.New(t)

	o := &Options{}
	a.NotError(o.Sanitize())
	a.NotNil(o.CORS).
		NotNil(o.NotFound).
		NotNil(o.MethodNotAllowed)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	o.MethodNotAllowed.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 405).
		Equal(w.Body.String(), http.StatusText(http.StatusMethodNotAllowed)+"\n")
}
