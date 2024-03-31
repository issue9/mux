// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package trace

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v7/header"
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
		Equal(w.Header().Get(header.ContentType), header.MessageHTTP)

	w = httptest.NewRecorder()
	a.NotError(Trace(w, r, true))
	body = w.Body.String()
	a.Contains(body, "/path").
		Contains(body, "&lt;body&gt;").
		True(strings.HasPrefix(body, http.MethodTrace)).
		Equal(w.Header().Get(header.ContentType), header.MessageHTTP)
}
