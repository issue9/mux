// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v5/internal/syntax"
)

func TestBuildMethodIndexes(t *testing.T) {
	a := assert.New(t, false)

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

	// 重置为空
	methodIndexes = map[int]methodIndexEntity{}
}

func TestNode_serveHTTP(t *testing.T) {
	a := assert.New(t, false)
	tree := New(false, syntax.NewInterceptors())

	a.NotError(tree.Add("/path", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("h1", "h1")
		w.WriteHeader(http.StatusAccepted)
		_, err := w.Write([]byte("get"))
		a.NotError(err)
	}), http.MethodGet))

	node, ps := tree.Route("/path")
	a.Empty(ps.Params).NotNil(node)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodHead, "/path", nil)
	node.Handler(http.MethodHead).ServeHTTP(w, r)
	a.Equal(w.Header().Get("h1"), "h1").
		Empty(w.Body.String()).
		Equal(w.Header().Get("Content-Length"), "3")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/path", nil)
	node.handlers[http.MethodOptions].ServeHTTP(w, r)
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
	a.Equal(tree.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions])
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions])

	// get=1,delete=2
	tree.buildMethods(1, http.MethodDelete, http.MethodGet)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 2, http.MethodGet: 1})
	a.Equal(tree.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions])

	// get=1,delete=1
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1, http.MethodGet: 1})
	a.Equal(tree.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions])

	// get=1,delete=0
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodGet: 1, http.MethodDelete: 0})
	a.Equal(tree.methodIndex, methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet]+methodIndexMap[http.MethodHead])
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions])
}
