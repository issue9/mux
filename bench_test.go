// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func init() {
	for _, api := range apis {
		ps := make(map[string]string, 5)
		p := api.pattern
		for {
			start := strings.IndexByte(p, '{')
			end := strings.IndexByte(p, '}')
			if start == -1 {
				break
			}
			if end <= start {
				panic("无效的路由项" + api.pattern)
			}
			k := p[start+1 : end]
			ps[k] = k

			if end+1 == len(p) {
				break
			}
			p = p[end+1:]
		}

		path := strings.Replace(api.pattern, "}", "", -1)
		api.test = strings.Replace(path, "{", "", -1)
		api.ps = ps
	}

	var srv *Router

	calcMemStats(func() {
		srv = NewRouter("", Lock)
		for _, api := range apis {
			srv.HandleFunc(api.pattern, benchHandler, api.method)
		}
	})
}

func benchHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(r.URL.Path)); err != nil {
		panic(err)
	}
}

func calcMemStats(load func()) {
	stats := &runtime.MemStats{}

	runtime.GC()
	runtime.ReadMemStats(stats)
	before := stats.HeapAlloc

	load()

	runtime.ReadMemStats(stats)
	after := stats.HeapAlloc
	fmt.Printf("%d Bytes\n", after-before)
}

func BenchmarkURL(b *testing.B) {
	a := assert.New(b)

	router := NewRouter("", Lock)
	for _, api := range apis {
		router.HandleFunc(api.pattern, benchHandler, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		url, err := URL(api.pattern, api.ps)
		a.NotError(err).Equal(url, api.test)
	}
}

func BenchmarkRouter_URL(b *testing.B) {
	a := assert.New(b)
	domain := "https://github.com"

	router := NewRouter("", Lock, URLDomain(domain))
	for _, api := range apis {
		router.HandleFunc(api.pattern, benchHandler, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		url, err := router.URL(true, api.pattern, api.ps)
		a.NotError(err).Equal(url, domain+api.test)
	}
}

// 动态添加和删除路由
func BenchmarkRouter_AddAndServeHTTP(b *testing.B) {
	router := NewRouter("", Lock)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r := httptest.NewRequest(api.method, api.test, nil)

		router.HandleFunc(api.pattern, benchHandler, api.method)
		router.ServeHTTP(w, r)
		router.Remove(api.pattern, api.method)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

// 对一组添加完成的路由进行路由操作
func BenchmarkRouter_ServeHTTP(b *testing.B) {
	router := NewRouter("")
	for _, api := range apis {
		router.HandleFunc(api.pattern, benchHandler, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r := httptest.NewRequest(api.method, api.test, nil)
		router.ServeHTTP(w, r)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

type api struct {
	method  string
	pattern string // {xx} 风格的路由项
	test    string // 测试地址
	ps      map[string]string
}

// 数据来源 github.com 的接口定义
var apis = []*api{
	{method: http.MethodGet, pattern: "/events"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/events"},
	{method: http.MethodGet, pattern: "/networks/{owner}/{repo}/events"},
	{method: http.MethodGet, pattern: "/orgs/{org}/events"},
	{method: http.MethodGet, pattern: "/users/{username}/received_events"},
	{method: http.MethodGet, pattern: "/users/{username}/received_events/public"},
	{method: http.MethodGet, pattern: "/users/{username}/events"},
	{method: http.MethodGet, pattern: "/users/{username}/events/public"},
	{method: http.MethodGet, pattern: "/users/{username}/events/orgs/{org}"},
	{method: http.MethodGet, pattern: "/feeds"},
	{method: http.MethodGet, pattern: "/notifications"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/notifications"},
	{method: http.MethodPut, pattern: "/notifications"},
	{method: http.MethodGet, pattern: "/notifications/threads/{id}"},
	{method: http.MethodPatch, pattern: "/notifications/threads/{id}"},
	{method: http.MethodGet, pattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodPut, pattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodDelete, pattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/stargazers"},
	{method: http.MethodGet, pattern: "/users/{username}/starred"},
	{method: http.MethodGet, pattern: "/user/starred"},
	{method: http.MethodGet, pattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodPut, pattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodDelete, pattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/subscribers"},
	{method: http.MethodGet, pattern: "/users/{username}/subscriptions"},
	{method: http.MethodGet, pattern: "/user/subscriptions"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodGet, pattern: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodPut, pattern: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodDelete, pattern: "/user/subscriptions/{owner}/{repo}"},

	{method: http.MethodGet, pattern: "/gists/{gist_id}/comments"},
	{method: http.MethodGet, pattern: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodPost, pattern: "/gists/{gist_id}/comments"},
	{method: http.MethodPatch, pattern: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodDelete, pattern: "/gists/{gist_id}/comments/{id}"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/blobs/{sha}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/git/blobs"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/commits/{sha}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/git/commits"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/refs/heads/feature"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/refs/tags"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/git/refs/{ref}"},

	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/git/tags"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/tags/{sha}"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/git/trees/{sha}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/git/trees"},

	{method: http.MethodGet, pattern: "/integration/installations"},
	{method: http.MethodGet, pattern: "/integration/installations/{installation_id}"},
	{method: http.MethodGet, pattern: "/user/installations"},
	{method: http.MethodPost, pattern: "/installations/{installation_id}/access_tokens"},
	{method: http.MethodGet, pattern: "/installation/repositories"},
	{method: http.MethodGet, pattern: "/user/installations/{installation_id}/repositories"},
	{method: http.MethodPut, pattern: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodDelete, pattern: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/assignees"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/assignees/{assignee}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/comments"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/{issue_number}/events"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/events"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/events/{id}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/issues/{number}/labels/{name}"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/milestones/{number}/labels"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/{issue_number}/timeline"},

	{method: http.MethodPost, pattern: "/orgs/{org}/migrations"},
	{method: http.MethodGet, pattern: "/orgs/{org}/migrations"},
	{method: http.MethodGet, pattern: "/orgs/{org}/migrations/{id}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/migrations/{id}/repos/{repo}_name/lock"},

	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/import/authors"},

	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/import/authors/{author_id}"},
	{method: http.MethodPatch, pattern: "/{owner}/{name}/import/lfs"},
	{method: http.MethodGet, pattern: "/{owner}/{name}/import/large_files"},

	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, pattern: "/emojis"},
	{method: http.MethodGet, pattern: "/gitignore/templates"},
	{method: http.MethodGet, pattern: "/gitignore/templates/C"},

	// license
	{method: http.MethodGet, pattern: "/licenses"},
	{method: http.MethodGet, pattern: "/licenses/{license}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/license"},

	{method: http.MethodPost, pattern: "/markdown"},
	{method: http.MethodPost, pattern: "/markdown/raw"},
	{method: http.MethodGet, pattern: "/meta"},
	{method: http.MethodGet, pattern: "/rate_limit"},

	{method: http.MethodGet, pattern: "/orgs/{org}/members"},
	{method: http.MethodGet, pattern: "/orgs/{org}/members/{username}"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/members/{username}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/public_members"},
	{method: http.MethodGet, pattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodPut, pattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodPut, pattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/invitations"},
	{method: http.MethodGet, pattern: "/user/memberships/orgs"},
	{method: http.MethodGet, pattern: "/user/memberships/orgs/{org}"},
	{method: http.MethodPatch, pattern: "/user/memberships/orgs/{org}"},

	{method: http.MethodGet, pattern: "/orgs/{org}/outside_collaborators"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/outside_collaborators/{username}"},
	{method: http.MethodPut, pattern: "/orgs/{org}/outside_collaborators/{username}"},

	{method: http.MethodGet, pattern: "/orgs/{org}/teams"},
	{method: http.MethodGet, pattern: "/teams/{id}"},
	{method: http.MethodPost, pattern: "/orgs/{org}/teams"},
	{method: http.MethodPatch, pattern: "/teams/{id}"},
	{method: http.MethodDelete, pattern: "/teams/{id}"},
	{method: http.MethodGet, pattern: "/teams/{id}/members/{username}"},
	{method: http.MethodPut, pattern: "/teams/{id}/members/{username}"},
	{method: http.MethodDelete, pattern: "/teams/{id}/members/{username}"},
	{method: http.MethodGet, pattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodPut, pattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodDelete, pattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodGet, pattern: "/teams/{id}/repos"},
	{method: http.MethodGet, pattern: "/teams/{id}/invitations"},
	//{method: http.MethodGet, pattern: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodPut, pattern: "/teams/{id}/repos/{org}/{repo}"},
	//{method: http.MethodDelete, pattern: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/hooks"},
	{method: http.MethodGet, pattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, pattern: "/orgs/{org}/hooks"},
	{method: http.MethodPatch, pattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, pattern: "/orgs/{org}/hooks/{id}/pings"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodGet, pattern: "/orgs/{org}/blocks"},
	{method: http.MethodGet, pattern: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodPut, pattern: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodDelete, pattern: "/orgs/{org}/blocks/{username}"},

	{method: http.MethodGet, pattern: "/projects/columns/{column_id}/cards"},
	{method: http.MethodGet, pattern: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, pattern: "/projects/columns/{column_id}/cards"},
	{method: http.MethodPatch, pattern: "/projects/columns/cards/{id}"},
	{method: http.MethodDelete, pattern: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, pattern: "/projects/columns/cards/{id}/moves"},

	{method: http.MethodGet, pattern: "/projects/{project_id}/columns"},
	{method: http.MethodPost, pattern: "/projects/{project_id}/columns"},
	{method: http.MethodDelete, pattern: "/projects/columns/{id}"},
	{method: http.MethodPost, pattern: "/projects/columns/{id}/moves"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/comments"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/events"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/dismissals"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/comments"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodDelete, pattern: "/reactions/{id}"},

	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodPatch, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodGet, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPut, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPost, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodDelete, pattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
}
