// SPDX-License-Identifier: MIT

package mux

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const defaultIndex = "index.html"

type fileServer struct {
	http.FileSystem
	paramName    string
	index        string
	errorHandler func(w http.ResponseWriter, status int, msg interface{})
}

// FileServer 声明静态文件服务
//
// fsys 访问的文件系统；
// name 路径保存在 context 中的参数名；
// index 当用户访问的是目录时，将自动读取此目录下的 index 文件，如果为空则为 index.html；
// errHandler 对各类出错的处理，如果为空会调用 http.Error 进行相应的处理。
// 如果要自定义，目前 status 可能的值有 403、404 和 500；
// 返回对象同时实现了 http.FileSystem 接口；
//
//  r := NewRouter("")
//  r.Get("/assets/{path}", FileServer(http.Dir("./assets"), "path", "index.html", nil)
func FileServer(fsys http.FileSystem, name, index string, errHandler func(w http.ResponseWriter, status int, msg interface{})) http.Handler {
	if fsys == nil {
		panic("参数 fsys 不能为空")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}

	if index == "" {
		index = defaultIndex
	}

	if errHandler == nil {
		errHandler = func(w http.ResponseWriter, status int, msg interface{}) {
			http.Error(w, fmt.Sprint(msg), status)
		}
	}

	return &fileServer{
		FileSystem:   fsys,
		paramName:    name,
		index:        index,
		errorHandler: errHandler,
	}
}

func (f *fileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, found := GetParams(r).Get(f.paramName)
	if !found {
		panic(fmt.Sprintf("未找到参数 %s 对应的值", f.paramName))
	}

	err := f.serve(name, w, r)
	switch {
	case errors.Is(err, os.ErrPermission):
		f.errorHandler(w, http.StatusForbidden, nil)
	case errors.Is(err, os.ErrNotExist):
		f.errorHandler(w, http.StatusNotFound, nil)
	case err != nil:
		f.errorHandler(w, http.StatusInternalServerError, err)
	}
}

// p 表示文件地址；
// 如果 p 是目录，则会自动读 p 目录下的 f.index 文件，
func (f *fileServer) serve(p string, w http.ResponseWriter, r *http.Request) error {
	if p == "" {
		p = "."
	}

STAT:
	fi, err := f.Open(p)
	if err != nil {
		return err
	}

	stat, err := fi.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		p = path.Join(p, f.index)
		goto STAT
	}

	data, err := ioutil.ReadAll(fi)
	if err != nil {
		return err
	}
	buf := bytes.NewReader(data)

	http.ServeContent(w, r, filepath.Base(p), stat.ModTime(), buf)
	return nil
}
