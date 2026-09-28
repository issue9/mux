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
	optionsIndexes = map[int]optionsEntity{}

	index := methodIndexes[http.MethodGet]
	buildOptionsIndexes(index)
	a.Equal(1, len(optionsIndexes)).
		Equal(optionsIndexes[index].options, "GET").
		Equal(optionsIndexes[index].methods, []string{"GET"})

	index = methodIndexes[http.MethodGet] + methodIndexes[http.MethodPatch]
	buildOptionsIndexes(index)
	a.Equal(2, len(optionsIndexes)).
		Equal(optionsIndexes[index].options, "GET, PATCH").
		Equal(optionsIndexes[index].methods, []string{"GET", "PATCH"})

	// 重置为空
	optionsIndexes = map[int]optionsEntity{}
}

func TestTree_buildMethods(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	// delete=1
	tree.buildMethods(1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1})
	a.Equal(tree.node.optionsIndex, methodIndexes[http.MethodDelete]+methodIndexes[http.MethodOptions])

	// get=1,delete=2
	tree.buildMethods(1, http.MethodDelete, http.MethodGet)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 2, http.MethodGet: 1})
	a.Equal(tree.node.optionsIndex, methodIndexes[http.MethodDelete]+methodIndexes[http.MethodOptions]+methodIndexes[http.MethodGet])

	// get=1,delete=1
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1, http.MethodGet: 1})
	a.Equal(tree.node.optionsIndex, methodIndexes[http.MethodDelete]+methodIndexes[http.MethodOptions]+methodIndexes[http.MethodGet])

	// get=1,delete=0
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodGet: 1, http.MethodDelete: 0})
	a.Equal(tree.node.optionsIndex, methodIndexes[http.MethodOptions]+methodIndexes[http.MethodGet])
}
