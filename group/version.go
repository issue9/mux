// SPDX-License-Identifier: MIT

package group

import (
	"net/http"
	"strings"
)

const versionString = "version="

// PathVersion 匹配路径中的版本号
//
// 会修改 *http.Request.URL.Path 的值，去掉匹配的版本号路径部分，比如：
//  /v1/path.html
// 如果匹配 v1 版本，会修改为：
//  /path.html
type PathVersion struct {
	// 需要匹配的版本号列表，需要以 / 作分隔，比如 /v3/  /v4/  /v11/
	Versions []string
}

// HeaderVersion 匹配报头的版本号
//
// 匹配报头 Accept 中的报头信息。
type HeaderVersion struct {
	Versions []string
}

// NewPathVersion 声明 PathVersion 实例
func NewPathVersion(version ...string) *PathVersion {
	for i, v := range version {
		if v == "" {
			panic("参数 v 不能为空值")
		}

		if v[0] != '/' {
			v = "/" + v
		}
		if v[len(v)-1] != '/' {
			v += "/"
		}
		version[i] = v
	}

	return &PathVersion{Versions: version}
}

func (v *HeaderVersion) Match(r *http.Request) (*http.Request, bool) {
	ver := findVersionNumberInHeader(r.Header.Get("Accept"))
	for _, vv := range v.Versions {
		if vv == ver {
			return r, true
		}
	}
	return nil, false
}

func (v *PathVersion) Match(r *http.Request) (*http.Request, bool) {
	p := r.URL.Path
	for _, ver := range v.Versions {
		if strings.HasPrefix(p, ver) {
			r.URL.Path = strings.TrimPrefix(p, ver[:len(ver)-1])
			return r, true
		}
	}
	return nil, false
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
