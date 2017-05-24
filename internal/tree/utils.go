// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import "errors"

const (
	start = '{'
	end   = '}'
)

// 判断两个字符串之间共同的开始内容的长度，
// 不会从{} 中间被分开。
func prefixLen(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	startIndex := -1
	for i := 0; i < l; i++ {
		if s1[i] == ':' { // 如果是正则，直接从 { 开始不再分隔
			return startIndex
		}

		if s1[i] == start {
			startIndex = i
		}

		if s1[i] == end {
			startIndex = -1
		}

		if s1[i] != s2[i] {
			if startIndex > -1 { // 不从命名参数中间分隔
				return startIndex
			}
			return i
		}
	}

	return l
}

// 检测路由语法是否正确
func checkSyntax(pattern string) error {
	brace := false
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		switch b {
		case start:
			if brace {
				return errors.New("不能嵌套 {")
			}
			brace = true
		case end:
			if !brace {
				return errors.New("} 必须与 { 成对出现")
			}
			brace = false
		}
	}

	return nil
}
