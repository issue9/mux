// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"strings"
)

const versionString = "version="

// Version 限定版本号的中间件
type Version struct {
	versions []string
	inHeader bool
}

// NewVersion 返回 Version 实例
func NewVersion(inHeader bool, version ...string) *Version {
	for i, v := range version {
		if v == "" {
			panic("参数 v 不能为空值")
		}

		if !inHeader {
			if v[0] != '/' {
				v = "/" + v
			}
			if v[len(v)-1] != '/' {
				v += "/"
			}
			version[i] = v
		}
	}

	return &Version{
		inHeader: inHeader,
		versions: version,
	}
}

// Match Matcher.Match
func (v *Version) Match(r *http.Request) bool {
	if v.inHeader {
		return v.matchInHeader(r)
	}
	return v.matchInURL(r)
}

func (v *Version) matchInHeader(r *http.Request) bool {
	ver := findVersionNumberInHeader(r.Header.Get("Accept"))
	for _, vv := range v.versions {
		if vv == ver {
			return true
		}
	}
	return false
}

func (v *Version) matchInURL(r *http.Request) bool {
	p := r.URL.Path
	for _, ver := range v.versions {
		if strings.HasPrefix(p, ver) {
			r.URL.Path = strings.TrimPrefix(p, ver[:len(ver)-1])
			return true
		}
	}
	return false
}

// 从 accept 中找到版本号，或是没有找到时，返回第二个参数 false。
func findVersionNumberInHeader(accept string) string {
	accepts := strings.Split(accept, ";")
	for _, str := range accepts {
		str = strings.ToLower(strings.TrimSpace(str))
		if index := strings.Index(str, versionString); index >= 0 {
			return str[index+len(versionString):]
		}
	}

	return ""
}
