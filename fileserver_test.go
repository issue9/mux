// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

func TestFileServer(t *testing.T) {
	a := assert.New(t, false)

	r := NewRouter("fs")
	a.NotNil(r)
	fs := FileServer(os.DirFS("./"), "path", "go.mod", nil)
	a.NotNil(fs)
	r.Get("/assets/{path}", fs)

	s := rest.NewServer(a, r, nil)

	// index == go.mod
	s.Get("/assets/").Do(nil).Status(200).
		Status(http.StatusOK).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.Contains(string(body), "module github.com/issue9/mux/")
		})

	s.NewRequest(http.MethodHead, "/assets/").
		Do(nil).Status(200).
		Status(http.StatusOK).
		BodyEmpty()

	s.Get("/assets/params/params.go").Do(nil).
		Status(http.StatusOK).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.Contains(string(body), "package params")
		})

	s.Get("/assets/not-exists").Do(nil).Status(http.StatusNotFound)

	a.Panic(func() {
		FileServer(nil, "name", "", nil)
	})

	a.Panic(func() {
		FileServer(os.DirFS("./"), "", "", nil)
	})

	fs = FileServer(os.DirFS("./"), "path", "", nil)
	fsys, ok := fs.(*fileServer)
	a.True(ok).
		Equal(fsys.index, defaultIndex).
		NotNil(fsys.errorHandler).
		Equal(fsys.paramName, "path")
}
