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

	"github.com/issue9/mux/v4/group"
)

var issue9Mux *Mux

func init() {
	for _, api := range apis {
		path := strings.Replace(api.bracePattern, "}", "", -1)
		api.test = strings.Replace(path, "{", "", -1)
	}

	calcMemStats(func() {
		h := func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte(r.URL.Path)); err != nil {
				panic(err)
			}
		}

		issue9Mux = New(true, false, nil, nil)
		def, err := issue9Mux.NewRouter("def", group.MatcherFunc(group.Any), Allowed())
		if err != nil {
			panic(err)
		}

		for _, api := range apis {
			if err := def.HandleFunc(api.bracePattern, h, api.method); err != nil {
				fmt.Println("calcMemStats:", err)
			}
		}
	})
}

func calcMemStats(load func()) {
	stats := &runtime.MemStats{}

	runtime.GC()
	runtime.ReadMemStats(stats)
	before := stats.HeapAlloc

	load()

	runtime.GC()
	runtime.ReadMemStats(stats)
	after := stats.HeapAlloc
	fmt.Printf("%d Bytes\n", after-before)
}

func benchGithubAPI(b *testing.B, srv http.Handler) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r := httptest.NewRequest(api.method, api.test, nil)
		srv.ServeHTTP(w, r)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

func BenchmarkGithubAPI_mux(b *testing.B) {
	benchGithubAPI(b, issue9Mux)
}

type api struct {
	method       string
	bracePattern string // {xx} 风格的路由项
	test         string // 测试地址
}

var apis = []*api{
	{method: http.MethodGet, bracePattern: "/events"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/events"},
	{method: http.MethodGet, bracePattern: "/networks/{owner}/{repo}/events"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/events"},
	{method: http.MethodGet, bracePattern: "/users/{username}/received_events"},
	{method: http.MethodGet, bracePattern: "/users/{username}/received_events/public"},
	{method: http.MethodGet, bracePattern: "/users/{username}/events"},
	{method: http.MethodGet, bracePattern: "/users/{username}/events/public"},
	{method: http.MethodGet, bracePattern: "/users/{username}/events/orgs/{org}"},
	{method: http.MethodGet, bracePattern: "/feeds"},
	{method: http.MethodGet, bracePattern: "/notifications"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/notifications"},
	{method: http.MethodPut, bracePattern: "/notifications"},
	{method: http.MethodGet, bracePattern: "/notifications/threads/{id}"},
	{method: http.MethodPatch, bracePattern: "/notifications/threads/{id}"},
	{method: http.MethodGet, bracePattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodPut, bracePattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodDelete, bracePattern: "/notifications/threads/{id}/subscription"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/stargazers"},
	{method: http.MethodGet, bracePattern: "/users/{username}/starred"},
	{method: http.MethodGet, bracePattern: "/user/starred"},
	{method: http.MethodGet, bracePattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodPut, bracePattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodDelete, bracePattern: "/user/starred/{owner}/{repo}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/subscribers"},
	{method: http.MethodGet, bracePattern: "/users/{username}/subscriptions"},
	{method: http.MethodGet, bracePattern: "/user/subscriptions"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodGet, bracePattern: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodPut, bracePattern: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodDelete, bracePattern: "/user/subscriptions/{owner}/{repo}"},

	{method: http.MethodGet, bracePattern: "/gists/{gist_id}/comments"},
	{method: http.MethodGet, bracePattern: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodPost, bracePattern: "/gists/{gist_id}/comments"},
	{method: http.MethodPatch, bracePattern: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodDelete, bracePattern: "/gists/{gist_id}/comments/{id}"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/blobs/{sha}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/git/blobs"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/commits/{sha}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/git/commits"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/refs/heads/feature"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/refs/tags"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/git/refs/{ref}"},

	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/git/tags"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/tags/{sha}"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/git/trees/{sha}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/git/trees"},

	{method: http.MethodGet, bracePattern: "/integration/installations"},
	{method: http.MethodGet, bracePattern: "/integration/installations/{installation_id}"},
	{method: http.MethodGet, bracePattern: "/user/installations"},
	{method: http.MethodPost, bracePattern: "/installations/{installation_id}/access_tokens"},
	{method: http.MethodGet, bracePattern: "/installation/repositories"},
	{method: http.MethodGet, bracePattern: "/user/installations/{installation_id}/repositories"},
	{method: http.MethodPut, bracePattern: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodDelete, bracePattern: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/assignees"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/assignees/{assignee}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/comments"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/{issue_number}/events"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/events"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/events/{id}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/issues/{number}/labels/{name}"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/milestones/{number}/labels"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/{issue_number}/timeline"},

	{method: http.MethodPost, bracePattern: "/orgs/{org}/migrations"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/migrations"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/migrations/{id}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/migrations/{id}/repos/{repo}_name/lock"},

	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/import/authors"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/import/authors/{author_id}"},
	{method: http.MethodPatch, bracePattern: "/{owner}/{name}/import/lfs"},
	{method: http.MethodGet, bracePattern: "/{owner}/{name}/import/large_files"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, bracePattern: "/emojis"},
	{method: http.MethodGet, bracePattern: "/gitignore/templates"},
	{method: http.MethodGet, bracePattern: "/gitignore/templates/C"},

	// license
	{method: http.MethodGet, bracePattern: "/licenses"},
	{method: http.MethodGet, bracePattern: "/licenses/{license}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/license"},

	{method: http.MethodPost, bracePattern: "/markdown"},
	{method: http.MethodPost, bracePattern: "/markdown/raw"},
	{method: http.MethodGet, bracePattern: "/meta"},
	{method: http.MethodGet, bracePattern: "/rate_limit"},

	{method: http.MethodGet, bracePattern: "/orgs/{org}/members"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/members/{username}"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/members/{username}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/public_members"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodPut, bracePattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodPut, bracePattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/invitations"},
	{method: http.MethodGet, bracePattern: "/user/memberships/orgs"},
	{method: http.MethodGet, bracePattern: "/user/memberships/orgs/{org}"},
	{method: http.MethodPatch, bracePattern: "/user/memberships/orgs/{org}"},

	{method: http.MethodGet, bracePattern: "/orgs/{org}/outside_collaborators"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/outside_collaborators/{username}"},
	{method: http.MethodPut, bracePattern: "/orgs/{org}/outside_collaborators/{username}"},

	{method: http.MethodGet, bracePattern: "/orgs/{org}/teams"},
	{method: http.MethodGet, bracePattern: "/teams/{id}"},
	{method: http.MethodPost, bracePattern: "/orgs/{org}/teams"},
	{method: http.MethodPatch, bracePattern: "/teams/{id}"},
	{method: http.MethodDelete, bracePattern: "/teams/{id}"},
	{method: http.MethodGet, bracePattern: "/teams/{id}/members/{username}"},
	{method: http.MethodPut, bracePattern: "/teams/{id}/members/{username}"},
	{method: http.MethodDelete, bracePattern: "/teams/{id}/members/{username}"},
	{method: http.MethodGet, bracePattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodPut, bracePattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodDelete, bracePattern: "/teams/{id}/memberships/{username}"},
	{method: http.MethodGet, bracePattern: "/teams/{id}/repos"},
	{method: http.MethodGet, bracePattern: "/teams/{id}/invitations"},
	//{method: http.MethodGet, bracePattern: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodPut, bracePattern: "/teams/{id}/repos/{org}/{repo}"},
	//{method: http.MethodDelete, bracePattern: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/hooks"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, bracePattern: "/orgs/{org}/hooks"},
	{method: http.MethodPatch, bracePattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, bracePattern: "/orgs/{org}/hooks/{id}/pings"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/blocks"},
	{method: http.MethodGet, bracePattern: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodPut, bracePattern: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodDelete, bracePattern: "/orgs/{org}/blocks/{username}"},

	{method: http.MethodGet, bracePattern: "/projects/columns/{column_id}/cards"},
	{method: http.MethodGet, bracePattern: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, bracePattern: "/projects/columns/{column_id}/cards"},
	{method: http.MethodPatch, bracePattern: "/projects/columns/cards/{id}"},
	{method: http.MethodDelete, bracePattern: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, bracePattern: "/projects/columns/cards/{id}/moves"},

	{method: http.MethodGet, bracePattern: "/projects/{project_id}/columns"},
	{method: http.MethodPost, bracePattern: "/projects/{project_id}/columns"},
	{method: http.MethodDelete, bracePattern: "/projects/columns/{id}"},
	{method: http.MethodPost, bracePattern: "/projects/columns/{id}/moves"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/comments"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/events"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/dismissals"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/comments"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodDelete, bracePattern: "/reactions/{id}"},

	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodPatch, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodGet, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPut, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPost, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodDelete, bracePattern: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
}

func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b)

	paths := []string{
		"",
		"/api//",
		"/api////users/1",
		"//api/users/1",
		"api///users////1",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api//../",
		"/api/..//../",
		"/api../",
		"api../",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ret := cleanPath(paths[i%len(paths)])
		a.True(len(ret) > 0)
	}
}
