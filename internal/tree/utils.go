// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import "errors"

func prefixLen(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	for i := 0; i < l; i++ {
		if s1[i] != s2[i] {
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
		case '{':
			if brace {
				return errors.New("不能嵌套 {")
			}
			brace = true
		case '}':
			if !brace {
				return errors.New("} 必须与 { 成对出现")
			}
			brace = false
		}
	}

	return nil
}
