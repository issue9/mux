// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package mux 适用第三方框架实现可定制的路由
//
// # 语法
//
// 路由参数采用大括号包含，内部包含名称和规则两部分：`{name:rule}`，
// 其中的 name 表示参数的名称，rule 表示对参数的约束规则。
//
// name 可以包含 `-` 前缀，表示在实际执行过程中，不捕获该名称的对应的值，
// 可以在一定程序上提升性能。
//
// rule 表示对参数的约束，一般为正则或是空，为空表示匹配任意值，
// 拦截器一栏中有关 rule 的高级用法。以下是一些常见的示例。
//
//	/posts/{id}.html                  // 匹配 /posts/1.html
//	/posts-{id}-{page}.html           // 匹配 /posts-1-10.html
//	/posts/{path:\\w+}.html           // 匹配 /posts/2020/11/11/title.html
//	/tags/{tag:\\w+}/{path}           // 匹配 /tags/abc/title.html
package mux

import (
	"bytes"
	"expvar"
	"html"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"path"
	"path/filepath"
	"strings"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/internal/tree"
)

const traceContentType = "message/http"

var emptyInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
func CheckSyntax(pattern string) error {
	_, err := emptyInterceptors.Split(pattern)
	return err
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
//
// NOTE: 仅仅是将 params 填入到 pattern 中， 不会判断参数格式是否正确。
func URL(pattern string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return pattern, nil
	}

	buf := errwrap.StringBuilder{}
	buf.Grow(len(pattern))
	if err := emptyInterceptors.URL(&buf, pattern, params); err != nil {
		return "", err
	}
	return buf.String(), buf.Err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}

// Trace 简单的 Trace 请求方法实现
//
// NOTE: 并不是百分百原样返回，具体可参考 net/http/httputil.DumpRequest 的说明。
// 如果内容包含特殊的 HTML 字符会被 html.EscapeString 转码。
func Trace(w http.ResponseWriter, r *http.Request, body bool) error {
	text, err := httputil.DumpRequest(r, body)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", traceContentType)
		_, err = w.Write([]byte(html.EscapeString(string(text))))
	}

	return err
}

// ServeFile 提供对静态文件的服务
//
// p 表示需要读取的文件名；
// index 表示 p 为目录时，默认读取的文件，为空表示 index.html；
func ServeFile(fsys fs.FS, p, index string, w http.ResponseWriter, r *http.Request) error {
	if index == "" {
		index = "index.html"
	}

	if p == "" || p[len(p)-1] == '/' {
		p += index
	}

STAT:
	f, err := fsys.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		p = path.Join(p, index)
		goto STAT
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		data := make([]byte, stat.Size())
		size, err := f.Read(data)
		if err != nil {
			return err
		}
		rs = bytes.NewReader(data[:size])
	}

	http.ServeContent(w, r, filepath.Base(p), stat.ModTime(), rs)
	return nil
}

// Debug 输出调试信息
//
// p 是指路由中的参数名，比如以下示例中，p 的值为 debug：
//
//	r.Get("/test{debug}", func(w http.ResponseWriter, r *http.Request) {
//	    p := mux.GetParams(r).String("debug")
//	    Debug(p, w, r)
//	}
//
// p 所代表的路径包含了前缀的 /。
func Debug(p string, w http.ResponseWriter, r *http.Request) error {
	switch {
	case p == "/vars":
		expvar.Handler().ServeHTTP(w, r)
	case p == "/pprof/cmdline":
		pprof.Cmdline(w, r)
	case p == "/pprof/profile":
		pprof.Profile(w, r)
	case p == "/pprof/symbol":
		pprof.Symbol(w, r)
	case p == "/pprof/trace":
		pprof.Trace(w, r)
	case p == "/pprof/goroutine":
		pprof.Handler("goroutine").ServeHTTP(w, r)
	case p == "/pprof/threadcreate":
		pprof.Handler("threadcreate").ServeHTTP(w, r)
	case p == "/pprof/mutex":
		pprof.Handler("mutex").ServeHTTP(w, r)
	case p == "/pprof/heap":
		pprof.Handler("heap").ServeHTTP(w, r)
	case p == "/pprof/block":
		pprof.Handler("block").ServeHTTP(w, r)
	case p == "/pprof/allocs":
		pprof.Handler("allocs").ServeHTTP(w, r)
	case strings.HasPrefix(p, "/pprof/"):
		// pprof.Index 写死了 /debug/pprof，所以直接替换这个变量
		r.URL.Path = "/debug/pprof/" + strings.TrimPrefix(p, "/pprof/")
		pprof.Index(w, r)
	case p == "/":
		w.Header().Set("Content-Type", "text/html")
		_, err := w.Write(debugHtml)
		return err
	default:
		http.NotFound(w, r)
	}
	return nil
}

var debugHtml = []byte(`
<!DOCTYPE HTML>
<html>
	<head>
		<title>Debug</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		<a href="vars">vars</a><br />
		<a href="pprof/cmdline">pprof/cmdline</a><br />
		<a href="pprof/profile">pprof/profile</a><br />
		<a href="pprof/symbol">pprof/symbol</a><br />
		<a href="pprof/trace">pprof/trace</a><br />
		<a href="pprof/goroutine">pprof/goroutine</a><br />
		<a href="pprof/threadcreate">pprof/threadcreate</a><br />
		<a href="pprof/mutex">pprof/mutex</a><br />
		<a href="pprof/heap">pprof/heap</a><br />
		<a href="pprof/block">pprof/block</a><br />
		<a href="pprof/allocs">pprof/allocs</a><br />
		<a href="pprof/">pprof/</a>
	</body>
</html>
`)
