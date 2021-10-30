// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestBuildMethodIndexes(t *testing.T) {
	a := assert.New(t)

	a.Empty(methodIndexes)

	index := methodIndexMap[http.MethodGet]
	buildMethodIndexes(index)
	a.Equal(1, len(methodIndexes)).
		Equal(methodIndexes[index].options, "GET").
		Equal(methodIndexes[index].methods, []string{"GET"})

	index = methodIndexMap[http.MethodGet] + methodIndexMap[http.MethodPatch]
	buildMethodIndexes(index)
	a.Equal(2, len(methodIndexes)).
		Equal(methodIndexes[index].options, "GET, PATCH").
		Equal(methodIndexes[index].methods, []string{"GET", "PATCH"})
}

func TestNode_serveHTTP(t *testing.T) {
	a := assert.New(t)
	tree := New(false)

	a.NotError(tree.Add("/path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("h1", "h1")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("get"))
	}), http.MethodGet))

	node, ps := tree.Route("/path")
	a.Empty(ps).NotNil(node)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodHead, "/path", nil)
	node.handlers[http.MethodHead].ServeHTTP(w, r)
	a.Equal(w.Header().Get("h1"), "h1").
		Empty(w.Body.String())

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	node.handlers[http.MethodOptions].ServeHTTP(w, r)
	a.Empty(w.Header().Get("h1")).
		Equal(w.Header().Get("Allow"), "GET, HEAD, OPTIONS").
		Empty(w.Body.String())
}
