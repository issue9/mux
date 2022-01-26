// SPDX-License-Identifier: MIT

package muxutil

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func TestTrace(t *testing.T) {
	a := assert.New(t, false)

	r := rest.NewRequest(a, http.MethodTrace, "/path").Body([]byte("<body>")).Request()
	w := httptest.NewRecorder()
	a.NotError(Trace(w, r, false))
	body := w.Body.String()
	a.Contains(body, "/path").
		NotContains(body, "body").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)

	w = httptest.NewRecorder()
	a.NotError(Trace(w, r, true))
	body = w.Body.String()
	a.Contains(body, "/path").
		Contains(body, "&lt;body&gt;").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get("content-type"), traceContentType)
}

func TestServeFile(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("../")

	w := httptest.NewRecorder()
	r := rest.Get(a, "/assets/").Request()
	a.NotError(ServeFile(fsys, "", "go.mod", w, r))
	a.Contains(w.Body.String(), "module github.com/issue9/mux")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.NotError(ServeFile(fsys, "params/params.go", "", w, r))
	a.NotEmpty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.ErrorIs(ServeFile(fsys, "params/", "", w, r), fs.ErrNotExist)
	a.Empty(w.Body.String())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/assets/").Request()
	a.ErrorIs(ServeFile(fsys, "not-exists", "", w, r), fs.ErrNotExist)
	a.Empty(w.Body.String())

}
