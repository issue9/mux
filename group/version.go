// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package group

import (
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/issue9/mux/v7/header"
	"github.com/issue9/mux/v7/types"
)

// PathVersion 匹配路径中的版本号
//
// 会修改 [http.Request.URL.Path] 的值，去掉匹配的版本号路径部分，比如：
//
//	/v1/path.html
//
// 如果匹配 v1 版本，会修改为：
//
//	/path.html
type PathVersion struct {
	paramName string
	versions  []string // 匹配的版本号列表，需要以 / 作分隔，比如 /v3/  /v4/  /v11/
}

// HeaderVersion 匹配报头的版本号
//
// 匹配报头 Accept 中的报头信息。
type HeaderVersion struct {
	paramName string
	acceptKey string
	versions  []string
	errlog    *log.Logger
}

// NewPathVersion 声明 PathVersion 实例
//
// param 将版本号作为参数保存到上下文中是的名称，如果不需要保存参数，可以设置为空值；
// version 版本的值，可以为空，表示匹配任意值；
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

// NewHeaderVersion 声明 HeaderVersion 实例
//
// param 将版本号作为参数保存到上下文中时的名称，如果不需要保存参数，可以设置为空值；
// errlog 错误日志输出通道，如果为空则采用 [log.Default]；
// key 表示在 accept 报头中的表示版本号的参数名，如果为空则采用 version；
// version 版本的值，可能为空，表示匹配任意值；
func NewHeaderVersion(param, key string, errlog *log.Logger, version ...string) *HeaderVersion {
	if key == "" {
		key = "version"
	}

	if errlog == nil {
		errlog = log.Default()
	}

	return &HeaderVersion{
		paramName: param,
		acceptKey: key,
		versions:  version,
		errlog:    errlog,
	}
}

func (v *HeaderVersion) Match(r *http.Request, ctx *types.Context) (ok bool) {
	header := r.Header.Get(header.Accept)
	if header == "" {
		return false
	}

	_, ps, err := mime.ParseMediaType(header)
	if err != nil {
		if v.errlog != nil {
			v.errlog.Println(err)
		}
		return false
	}

	ver := ps[v.acceptKey]
	for _, vv := range v.versions {
		if vv == ver {
			if v.paramName != "" {
				ctx.Set(v.paramName, vv)
			}
			return true
		}
	}
	return false
}

func (v *PathVersion) Match(r *http.Request, ctx *types.Context) (ok bool) {
	p := r.URL.Path
	for _, ver := range v.versions {
		if strings.HasPrefix(p, ver) {
			vv := ver[:len(ver)-1]

			r.URL.Path = strings.TrimPrefix(p, vv)
			if v.paramName != "" {
				ctx.Set(v.paramName, vv)
			}

			return true
		}
	}
	return false
}
