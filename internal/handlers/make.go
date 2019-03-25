// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"

	"github.com/issue9/mux/v2/internal/handlers"
)

const (
	fileheader         = "// 该文件由 make.go 产生，不需要手动修改！\n\n"
	filename           = "./methods.go"
	packagename        = "handlers"
	methodTypeName     = "methodMap"
	optionsStringsName = "optionsStrings"
)

func main() {
	buf := new(bytes.Buffer)

	ws := func(format string, v ...interface{}) {
		_, err := fmt.Fprintf(buf, format, v...)
		if err != nil {
			panic(err)
		}
	}

	var maps = map[int]string{}
	var size int

	ws(fileheader)
	ws("package %s\n\n", packagename)

	ws("var %s=map[string]int{\n", methodTypeName)
	for index, method := range handlers.Methods {
		ii := 1 << uint(index)
		ws("\"%s\":%d,\n", method, ii)
		maps[ii] = method
		size += ii
	}
	ws("}\n\n")

	ws("var %s = []string{\n", optionsStringsName)

	methods := make([]string, 0, len(handlers.Methods))
	for i := 0; i <= size; i++ {
		for k, v := range maps {
			if i&k == k {
				methods = append(methods, v)
			}
		}

		sort.Strings(methods) // 统一的排序，方便测试使用
		ws("\"%s\",\n", strings.Join(methods, ", "))
		methods = methods[:0]
	} // end for

	ws("}")

	data, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	file.Write(data)
	defer file.Close()
}
