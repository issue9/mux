// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/mux/v7/internal/syntax"
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

func TestTree_buildMethods(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, false, syntax.NewInterceptors())

	// delete=1
	tree.buildMethods(1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions])

	// get=1,delete=2
	tree.buildMethods(1, http.MethodDelete, http.MethodGet)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 2, http.MethodGet: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet])

	// get=1,delete=1
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodDelete: 1, http.MethodGet: 1})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodDelete]+methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet])

	// get=1,delete=0
	tree.buildMethods(-1, http.MethodDelete)
	a.Equal(tree.methods, map[string]int{http.MethodGet: 1, http.MethodDelete: 0})
	a.Equal(tree.node.methodIndex, methodIndexMap[http.MethodOptions]+methodIndexMap[http.MethodGet])
}
