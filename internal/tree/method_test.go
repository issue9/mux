// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/params"
)

func TestBuildMethodIndexes(t *testing.T) {
	a := assert.New(t, false)
	methodIndexes = map[int]methodIndexEntity{}

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

	// 重置为空
	methodIndexes = map[int]methodIndexEntity{}
}

func TestNode_serveHTTP(t *testing.T) {
	a := assert.New(t, false)
	tree := New(false, syntax.NewInterceptors())

	// path1，主动调用 WriteHeader

	a.NotError(tree.Add("/path1", func(w http.ResponseWriter, r *http.Request, _ params.Params) {
		w.Header().Set("h1", "h1")
		w.WriteHeader(http.StatusAccepted)

		_, err := w.Write([]byte("get"))
		a.NotError(err)
		_, err = w.Write([]byte("get"))
		a.NotError(err)
	}, http.MethodGet))

	node, ps := tree.Route("/path1")
	a.Zero(ps.Count()).NotNil(node)

	r := rest.NewRequest(a, http.MethodHead, "/path1").Request()
	w := httptest.NewRecorder()
	node.Handler(http.MethodHead)(w, r, nil)
	a.Equal(w.Header().Get("h1"), "h1").
		Empty(w.Body.String()).
		Equal(w.Header().Get("Content-Length"), "6")

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path1").Request()
	node.handlers[http.MethodOptions](w, r, nil)
	a.Empty(w.Header().Get("h1")).
		Equal(w.Header().Get("Allow"), "GET, HEAD, OPTIONS").
		Empty(w.Body.String())

	// path2，不主动调用 WriteHeader

	a.NotError(tree.Add("/path2", func(w http.ResponseWriter, r *http.Request, _ params.Params) {
		w.Header().Set("h1", "h2")

		_, err := w.Write([]byte("get"))
		a.NotError(err)
		_, err = w.Write([]byte("get"))
		a.NotError(err)
	}, http.MethodGet))

	node, ps = tree.Route("/path2")
	a.Zero(ps.Count()).NotNil(node)

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodHead, "/path2").Request()
	node.Handler(http.MethodHead)(w, r, nil)
	a.Equal(w.Header().Get("h1"), "h2").
		Empty(w.Body.String()).
		Equal(w.Header().Get("Content-Length"), "6")

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path2").Request()
	node.handlers[http.MethodOptions](w, r, nil)
	a.Empty(w.Header().Get("h1")).
		Equal(w.Header().Get("Allow"), "GET, HEAD, OPTIONS").
		Empty(w.Body.String())
}

func TestTree_buildMethods(t *testing.T) {
	a := assert.New(t, false)
	tree := New(false, syntax.NewInterceptors())
	a.NotNil(tree)

	// delete=1
	tree.buildMethods(1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions])

	// get=1,delete=2
	tree.buildMethods(1, http.MethodDelete, http.MethodGet)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 2, http.MethodGet: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])

	// get=1,delete=1
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1, http.MethodGet: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])

	// get=1,delete=0
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodGet: 1, http.MethodDelete: 0})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])
}
