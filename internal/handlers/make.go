// SPDX-License-Identifier: MIT

// +build ignore

package main

import (
	"bytes"
	"go/format"
	"os"
	"sort"
	"strings"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v4/internal/handlers"
)

const (
	fileHeader         = "// 该文件由 make.go 产生，不需要手动修改！\n\n"
	filename           = "./methods.go"
	packageName        = "handlers"
	methodTypeName     = "methodMap"
	optionsStringsName = "optionsStrings"
)

func main() {
	buf := &errwrap.Buffer{Buffer: bytes.Buffer{}}

	var maps = map[int]string{}
	var size int

	buf.WString(fileHeader)
	buf.Printf("package %s\n\n", packageName)

	buf.Printf("var %s=map[string]int{\n", methodTypeName)
	for index, method := range handlers.Methods {
		ii := 1 << uint(index)
		buf.Printf("\"%s\":%d,\n", method, ii)
		maps[ii] = method
		size += ii
	}
	buf.WString("}\n\n")

	buf.Printf("var %s = []string{\n", optionsStringsName)

	methods := make([]string, 0, len(handlers.Methods))
	for i := 0; i <= size; i++ {
		for k, v := range maps {
			if i&k == k {
				methods = append(methods, v)
			}
		}

		sort.Strings(methods) // 统一的排序，方便测试使用
		buf.Printf("\"%s\",\n", strings.Join(methods, ", "))
		methods = methods[:0]
	} // end for

	buf.WByte('}')

	data, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	file.Write(data)

	if err = file.Close(); err != nil {
		panic(err)
	}
}
