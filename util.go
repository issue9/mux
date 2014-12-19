// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"regexp"
)

// 分析命名捕获，并以map[string]string方式返回所有的命名捕获。
// 若不存在命名组，则返回一个空的map而不是nil
func parseCaptures(expr *regexp.Regexp, str string) map[string]string {
	ret := make(map[string]string)

	subexps := expr.SubexpNames()
	args := expr.FindStringSubmatch(str)
	for index, name := range subexps {
		if name == "" {
			continue
		}

		ret[name] = args[index]
	}

	return ret
}
