// SPDX-License-Identifier: MIT

package group

import (
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/issue9/mux/v5/internal/syntax"
)

const versionKey = "version"

// PathVersion 匹配路径中的版本号
//
// 会修改 http.Request.URL.Path 的值，去掉匹配的版本号路径部分，比如：
//  /v1/path.html
// 如果匹配 v1 版本，会修改为：
//  /path.html
type PathVersion struct {
	paramName string

	// 需要匹配的版本号列表，需要以 / 作分隔，比如 /v3/  /v4/  /v11/
	versions []string
}

// HeaderVersion 匹配报头的版本号
//
// 匹配报头 Accept 中的报头信息。
type HeaderVersion struct {
	// 将版本号作为参数保存到上下文中是的名称
	//
	// 如果不需要，可以设置为空值。
	Key string

	// 支持的版本号列表
	//
	// 比如 accept=application/json;version=1.0，version= 之后的内容。
	Versions []string

	// 错误日志输出通道
	//
	// 如果为空，则不输出任何内容。
	ErrLog *log.Logger
}

// NewPathVersion 声明 PathVersion 实例
//
// param 将版本号作为参数保存到上下文中是的名称，如果不需要，可以设置为空值。
func NewPathVersion(param string, version ...string) *PathVersion {
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

	return &PathVersion{paramName: param, versions: version}
}

func (v *HeaderVersion) Match(r *http.Request) (*http.Request, bool) {
	_, ps, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		if v.ErrLog != nil {
			v.ErrLog.Println(err)
		}
		return nil, false
	}

	ver := ps[versionKey]
	for _, vv := range v.Versions {
		if vv == ver {
			if v.Key != "" {
				r = syntax.WithValue(r, &syntax.Params{Params: []syntax.Param{{K: v.Key, V: vv}}})
			}
			return r, true
		}
	}
	return nil, false
}

func (v *PathVersion) Match(r *http.Request) (*http.Request, bool) {
	p := r.URL.Path
	for _, ver := range v.versions {
		if strings.HasPrefix(p, ver) {
			vv := ver[:len(ver)-1]

			if v.paramName == "" {
				r = r.Clone(r.Context()) // r.URL.Path 已改变
			} else {
				r = syntax.WithValue(r, &syntax.Params{Params: []syntax.Param{{K: v.paramName, V: vv}}})
			}

			r.URL.Path = strings.TrimPrefix(p, vv)
			return r, true
		}
	}
	return nil, false
}
