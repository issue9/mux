// SPDX-License-Identifier: MIT

package options

type URL struct {
	Strict bool   // 是否需要确认该路由项真实存在
	Domain string // 生成地址的域名部分
}

func (u *URL) sanitize() {
	l := len(u.Domain)
	if l != 0 && u.Domain[l-1] == '/' {
		u.Domain = u.Domain[:l-1]
	}
}
