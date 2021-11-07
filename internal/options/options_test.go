// SPDX-License-Identifier: MIT

package options

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestOptions_sanitize(t *testing.T) {
	a := assert.New(t)

	o := &Options{}
	a.NotError(o.sanitize())
	a.NotNil(o.CORS).
		NotNil(o.NotFound).
		NotNil(o.MethodNotAllowed)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	o.MethodNotAllowed.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, 405).
		Equal(w.Body.String(), http.StatusText(http.StatusMethodNotAllowed)+"\n")
}

func TestBuild(t *testing.T) {
	a := assert.New(t)

	o, err := Build()
	a.NotError(err).
		NotNil(o).
		False(o.CaseInsensitive).
		NotNil(o.CORS).
		NotNil(o.NotFound).
		NotNil(o.Interceptors)

	o, err = Build(func(o *Options) { o.CaseInsensitive = true })
	a.NotError(err).
		NotNil(o).
		True(o.CaseInsensitive)

	o, err = Build(func(o *Options) {
		o.CORS = &CORS{
			Origins:          []string{"*"},
			AllowCredentials: true,
		}
	})
	a.ErrorString(err, "不能同时成立").Nil(o)
}
