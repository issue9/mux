// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:generate go run make.go

package handlers

// Map 返回所有的 methodMap 数据，仅供 make.go 产生数据
func Map() map[int16]string {
	ret := make(map[int16]string, len(methodTypeMap))
	for k, v := range methodTypeMap {
		ret[int16(k)] = v
	}
	return ret
}
