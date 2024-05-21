// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/internal/syntax"
	"github.com/issue9/mux/v8/internal/tree"
	"github.com/issue9/mux/v8/types"
)

type (
	// Matcher 验证一个请求是否符合要求
	//
	// Matcher 用于路由项的前置判断，用于对路由项进行归类，
	// 符合同一个 Matcher 的路由项，再各自进行路由。比如按域名进行分组路由。
	Matcher interface {
		// Match 验证请求是否符合当前对象的要求
		//
		// 如果返回 false，那么不应当对参所指向的内容作修改，否则可能影响后续的判断。
		Match(*http.Request, *types.Context) bool
	}

	MatcherFunc func(*http.Request, *types.Context) bool

	// Hosts 限定域名的匹配工具
	Hosts struct {
		i    *syntax.Interceptors
		tree *tree.Tree[any]
	}

	pathVersion struct {
		paramName string
		versions  []string // 匹配的版本号列表，需要以 / 作分隔，比如 /v3/  /v4/  /v11/
	}

	headerVersion struct {
		paramName string
		acceptKey string
		versions  []string
		errlog    func(error)
	}
)

func (f MatcherFunc) Match(r *http.Request, p *types.Context) bool { return f(r, p) }

func anyRouter(*http.Request, *types.Context) bool { return true }

// AndMatcher 按顺序符合每一个要求
//
// 前一个对象返回的实例将作为下一个对象的输入参数。
func AndMatcher(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request, ctx *types.Context) bool {
		for _, mm := range m {
			if !mm.Match(r, ctx) {
				return false
			}
		}
		return true
	})
}

// OrMatcher 仅需符合一个要求
func OrMatcher(m ...Matcher) Matcher {
	return MatcherFunc(func(r *http.Request, ctx *types.Context) bool {
		for _, mm := range m {
			if ok := mm.Match(r, ctx); ok {
				return true
			}
		}
		return false
	})
}

// AndMatcherFunc 需同时符合每一个要求
func AndMatcherFunc(f ...func(*http.Request, *types.Context) bool) Matcher {
	return AndMatcher(f2i(f...)...)
}

// OrMatcherFunc 仅需符合一个要求
func OrMatcherFunc(f ...func(*http.Request, *types.Context) bool) Matcher {
	return OrMatcher(f2i(f...)...)
}

func f2i(f ...func(*http.Request, *types.Context) bool) []Matcher {
	ms := make([]Matcher, 0, len(f))
	for _, ff := range f {
		ms = append(ms, MatcherFunc(ff))
	}
	return ms
}

// NewHosts 声明新的 [Hosts] 实例
func NewHosts(lock bool, domain ...string) *Hosts {
	i := syntax.NewInterceptors()
	f := func(types.Node) any { return nil }
	t := tree.New("", lock, i, nil, false, f, f)
	h := &Hosts{tree: t, i: i}
	h.Add(domain...)
	return h
}

// RegisterInterceptor 注册拦截器
//
// NOTE: 拦截器只有在注册之后添加的域名才有效果。
func (hs *Hosts) RegisterInterceptor(f InterceptorFunc, name ...string) { hs.i.Add(f, name...) }

func (hs *Hosts) Match(r *http.Request, ctx *types.Context) bool {
	h := r.Host // r.URL.Hostname() 可能为空，r.Host 一直有值！
	if i := strings.LastIndexByte(h, ':'); i != -1 && validOptionalPort(h[i:]) {
		h = h[:i]
	}
	if strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]") { // ipv6
		h = h[1 : len(h)-1]
	}

	ctx.Path = strings.ToLower(h)
	_, _, exists := hs.tree.Handler(ctx, http.MethodGet)
	return exists
}

// 源自 https://github.com/golang/go/blob/d8762b2f4532cc2e5ec539670b88bbc469a13938/src/net/url/url.go#L769
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// Add 添加新的域名
//
// 域名的格式和路由的语法格式是一样的，比如：
//
//	api.example.com
//	{sub:[a-z]+}.example.com
//
// 如果存在命名参数，也可以通过也可通过 [types.Params] 接口获取。
// 当语法错误时，会触发 panic，可通过 [CheckSyntax] 检测语法的正确性。
func (hs *Hosts) Add(domain ...string) {
	for _, d := range domain {
		err := hs.tree.Add(strings.ToLower(d), hs.emptyHandlerFunc, nil, http.MethodGet)
		if err != nil {
			panic(err)
		}
	}
}

func (hs *Hosts) Delete(domain string) { hs.tree.Remove(domain) }

func (hs *Hosts) emptyHandlerFunc() {}

// NewPathVersion 声明匹配路径中版本号的 [Matcher] 实例
//
// param 将版本号作为参数保存到上下文中是的名称，如果不需要保存参数，可以设置为空值；
// version 版本的值，可以为空，表示匹配任意值；
//
// NOTE: 会修改 [http.Request.URL.Path] 的值，去掉匹配的版本号路径部分，比如：
//
//	/v1/path.html
//
// 如果匹配 v1 版本，会修改为：
//
//	/path.html
func NewPathVersion(param string, version ...string) Matcher {
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

	return &pathVersion{paramName: param, versions: version}
}

// NewHeaderVersion 声明匹配报头 Accept 中版本号的 [Matcher] 实现
//
// param 将版本号作为参数保存到上下文中时的名称，如果不需要保存参数，可以设置为空值；
// errlog 错误日志输出通道，如果为空则采用 [log.Default]；
// key 表示在 accept 报头中的表示版本号的参数名，如果为空则采用 version；
// version 版本的值，可能为空，表示匹配任意值；
func NewHeaderVersion(param, key string, errlog func(error), version ...string) Matcher {
	if key == "" {
		key = "version"
	}

	if errlog == nil {
		errlog = func(err error) { log.Println(err) }
	}

	return &headerVersion{
		paramName: param,
		acceptKey: key,
		versions:  version,
		errlog:    errlog,
	}
}

func (v *headerVersion) Match(r *http.Request, ctx *types.Context) bool {
	header := r.Header.Get(header.Accept)
	if header == "" {
		return false
	}

	_, ps, err := mime.ParseMediaType(header)
	if err != nil {
		v.errlog(err)
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

func (v *pathVersion) Match(r *http.Request, ctx *types.Context) bool {
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
