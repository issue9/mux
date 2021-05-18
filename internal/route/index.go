// SPDX-License-Identifier: MIT

package route

import (
	"sort"
	"strings"
)

var methodIndexes map[string]int

var indexes = map[int]indexEntity{}

func init() {
	methodIndexes = make(map[string]int, len(Methods))
	for i, m := range Methods {
		methodIndexes[m] = 1 << i
	}
}

type indexEntity struct {
	methods []string
	options string
}

func (r *Route) buildMethods() {
	r.methodIndex = 0
	for method := range r.handlers {
		r.methodIndex += methodIndexes[method]
	}

	if _, found := indexes[r.methodIndex]; !found {
		methods := make([]string, 0, len(r.handlers))
		for method := range r.handlers {
			methods = append(methods, method)
		}
		sort.Strings(methods)
		entity := indexEntity{methods: methods, options: strings.Join(methods, ", ")}
		indexes[r.methodIndex] = entity
	}
}
