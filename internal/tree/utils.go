// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

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
