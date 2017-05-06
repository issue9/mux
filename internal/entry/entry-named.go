// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

type name struct {
	name    string // 名称
	start   int    // 名称的起始字符
	endByte byte   // 结束后的第一个字符
}

type named struct {
	*items
	names []*name
}

// Entry.Type
func (n *named) Type() int {
	return TypeRegexp
}

// Entry.Match
func (n *named) Match(path string) int {
	pathIndex := 0
	for _, name := range n.names {

	}
}

// Entry.Params
func (n *named) Params(url string) map[string]string {
	// TODO
	return nil
}
