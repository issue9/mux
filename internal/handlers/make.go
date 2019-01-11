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
	fileheader  = "// 该文件由 make.go 产生，不需要手动修改！\n\n"
	filename    = "./options_table.go"
	packagename = "handlers"
)

func main() {
	items := handlers.Map()
	keys := make([]int16, 0, len(items))
	var size int16
	for k := range items {
		keys = append(keys, k)
		size += k
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	buf := new(bytes.Buffer)

	buf.WriteString(fileheader)
	fmt.Fprintf(buf, "package %s\n\n", packagename)
	fmt.Fprintln(buf, "var optionsStrings = []string{")

	methods := make([]string, 0, len(items))
	for i := int16(0); i <= size; i++ {
		for _, k := range keys {
			if i&k == k {
				methods = append(methods, items[k])
			}
		}

		sort.Strings(methods) // 统一的排序，方便测试使用
		fmt.Fprintf(buf, "\"%s\",\n", strings.Join(methods, ", "))
		methods = methods[:0]
	} // end for

	buf.WriteByte('}')

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
