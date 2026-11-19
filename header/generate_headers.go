// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"encoding/csv"
	"io/fs"
	"net/http"
	"strings"

	"github.com/issue9/errwrap"
	"github.com/issue9/source/codegen"
)

const (
	url        = "https://www.iana.org/assignments/http-fields/field-names.csv"
	pkgName    = "header"
	filename   = "generated_headers.go"
	fileHeader = "// 文件由工具生成，如无必要不要手动修改！\n\n"
)

func main() {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	ianaReader := csv.NewReader(resp.Body)
	ianaRecords, err := ianaReader.ReadAll()
	if err != nil {
		panic(err)
	}

	buf := &errwrap.Buffer{}
	buf.WString(fileHeader).
		WString("package ").WString(pkgName).WString("\n\n")

	buf.WString("const(\n")

	// 第一元素为标题，最后一个元素为 *
	for _, record := range ianaRecords[1 : len(ianaRecords)-1] {
		if record[1] == "deprecated" {
			continue
		}

		val := record[0]
		id := strings.ReplaceAll(val, "-", "")
		buf.Printf("%s=\"%s\"// %s\n", id, val, record[3])
	}
	buf.WString(")\n\n")

	if err = codegen.Dump(filename, buf.Bytes(), fs.ModePerm); err != nil {
		panic(err)
	}
}
