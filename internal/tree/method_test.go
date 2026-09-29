// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v5"

	"github.com/issue9/mux/v10/internal/syntax"
)

func TestBuildOptionsIndexes(t *testing.T) {
	a := assert.New(t, false)
	optionsIndexesMux.Lock()
	optionsIndexes = map[int]*optionsEntity{}
	optionsIndexesMux.Unlock()

	index := methodIndexes[http.MethodGet]
	buildOptionsIndexes(index)
	a.Equal(1, len(optionsIndexes)).
		Equal(optionsIndexes[index].allow, "GET").
		Equal(optionsIndexes[index].methods, []string{"GET"})

	index = methodIndexes[http.MethodGet] + methodIndexes[http.MethodPatch]
	buildOptionsIndexes(index)
	a.Equal(2, len(optionsIndexes)).
		Equal(optionsIndexes[index].allow, "GET, PATCH").
		Equal(optionsIndexes[index].methods, []string{"GET", "PATCH"})

	// 重置为空
	optionsIndexesMux.Lock()
	optionsIndexes = map[int]*optionsEntity{}
	optionsIndexesMux.Unlock()
}

func TestTree_buildMethods(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	// delete=1
	tree.buildMethods(1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1})
	a.Equal(tree.node.options.Load().allow, "DELETE, OPTIONS")

	// get=1,delete=2
	tree.buildMethods(1, http.MethodDelete, http.MethodGet)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 2, http.MethodGet: 1})
	a.Equal(tree.node.options.Load().allow, "DELETE, GET, OPTIONS")

	// get=1,delete=1
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1, http.MethodGet: 1})
	a.Equal(tree.node.options.Load().allow, "DELETE, GET, OPTIONS")

	// get=1,delete=0
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodGet: 1, http.MethodDelete: 0})
	a.Equal(tree.node.options.Load().allow, "GET, OPTIONS")
}
