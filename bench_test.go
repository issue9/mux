// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// go1.8.1 BenchmarkGithubAPI-4   	  200000	      7001 ns/op	    6444 B/op	      22 allocs/op
func BenchmarkGithubAPI(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.Path))
	}

	mux := New(false, false, nil, nil)
	for _, api := range apis {
		mux.AddFunc(api.path, h, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r := httptest.NewRequest(api.method, api.test, nil)
		mux.ServeHTTP(w, r)

		if w.Body.String() != r.URL.Path {
			b.Errorf("BenchmarkGithubAPI: %v:%v", w.Body.String(), r.URL.Path)
		}
	}
}

func init() {
	for _, api := range apis {
		path := strings.Replace(api.path, "{", "", -1)
		api.test = strings.Replace(path, "}", "", -1)
	}
}

type api struct {
	method string
	path   string
	test   string // 测试地址
}

var apis = []*api{
	{method: http.MethodGet, path: "/events"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/events"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/events"},
	{method: http.MethodGet, path: "/networks/{owner}/{repo}/events"},
	{method: http.MethodGet, path: "/orgs/{org}/events"},
	{method: http.MethodGet, path: "/users/{username}/received_events"},
	{method: http.MethodGet, path: "/users/{username}/received_events/public"},
	{method: http.MethodGet, path: "/users/{username}/events"},
	{method: http.MethodGet, path: "/users/{username}/events/public"},
	{method: http.MethodGet, path: "/users/{username}/events/orgs/{org}"},
	{method: http.MethodGet, path: "/feeds"},
	{method: http.MethodGet, path: "/notifications"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/notifications"},
	{method: http.MethodPut, path: "/notifications"},
	{method: http.MethodGet, path: "/notifications/threads/{id}"},
	{method: http.MethodPatch, path: "/notifications/threads/{id}"},
	{method: http.MethodGet, path: "/notifications/threads/{id}/subscription"},
	{method: http.MethodPut, path: "/notifications/threads/{id}/subscription"},
	{method: http.MethodDelete, path: "/notifications/threads/{id}/subscription"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/stargazers"},
	{method: http.MethodGet, path: "/users/{username}/starred"},
	{method: http.MethodGet, path: "/user/starred"},
	{method: http.MethodGet, path: "/user/starred/{owner}/{repo}"},
	{method: http.MethodPut, path: "/user/starred/{owner}/{repo}"},
	{method: http.MethodDelete, path: "/user/starred/{owner}/{repo}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/subscribers"},
	{method: http.MethodGet, path: "/users/{username}/subscriptions"},
	{method: http.MethodGet, path: "/user/subscriptions"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/subscription"},
	{method: http.MethodGet, path: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodPut, path: "/user/subscriptions/{owner}/{repo}"},
	{method: http.MethodDelete, path: "/user/subscriptions/{owner}/{repo}"},

	{method: http.MethodGet, path: "/gists/{gist_id}/comments"},
	{method: http.MethodGet, path: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodPost, path: "/gists/{gist_id}/comments"},
	{method: http.MethodPatch, path: "/gists/{gist_id}/comments/{id}"},
	{method: http.MethodDelete, path: "/gists/{gist_id}/comments/{id}"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/blobs/{sha}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/git/blobs"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/commits/{sha}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/git/commits"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/commits/{sha}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/refs/heads/feature"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/refs/tags"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/git/refs"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/git/refs/{ref}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/git/refs/{ref}"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/tags/{sha}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/git/tags"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/tags/{sha}"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/git/trees/{sha}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/git/trees"},

	{method: http.MethodGet, path: "/integration/installations"},
	{method: http.MethodGet, path: "/integration/installations/{installation_id}"},
	{method: http.MethodGet, path: "/user/installations"},
	{method: http.MethodPost, path: "/installations/{installation_id}/access_tokens"},
	{method: http.MethodGet, path: "/installation/repositories"},
	{method: http.MethodGet, path: "/user/installations/{installation_id}/repositories"},
	{method: http.MethodPut, path: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodDelete, path: "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/assignees"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/assignees/{assignee}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/comments"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/issues/{number}/comments"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/issues/comments/{id}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/{issue_number}/events"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/events"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/events/{id}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/labels"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/labels/{name}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/issues/{number}/labels/{name}"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/issues/{number}/labels"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/milestones/{number}/labels"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/milestones"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/milestones/{number}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/{issue_number}/timeline"},

	{method: http.MethodPost, path: "/orgs/{org}/migrations"},
	{method: http.MethodGet, path: "/orgs/{org}/migrations"},
	{method: http.MethodGet, path: "/orgs/{org}/migrations/{id}"},
	{method: http.MethodGet, path: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, path: "/orgs/{org}/migrations/{id}/archive"},
	{method: http.MethodDelete, path: "/orgs/{org}/migrations/{id}/repos/{repo}_name/lock"},

	{method: http.MethodPut, path: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/import"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/import/authors"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/import/authors/{author_id}"},
	{method: http.MethodPatch, path: "/{owner}/{name}/import/lfs"},
	{method: http.MethodGet, path: "/{owner}/{name}/import/large_files"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/import"},
	{method: http.MethodGet, path: "/emojis"},
	{method: http.MethodGet, path: "/gitignore/templates"},
	{method: http.MethodGet, path: "/gitignore/templates/C"},

	// license
	{method: http.MethodGet, path: "/licenses"},
	{method: http.MethodGet, path: "/licenses/{license}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/license"},

	{method: http.MethodPost, path: "/markdown"},
	{method: http.MethodPost, path: "/markdown/raw"},
	{method: http.MethodGet, path: "/meta"},
	{method: http.MethodGet, path: "/rate_limit"},

	{method: http.MethodGet, path: "/orgs/{org}/members"},
	{method: http.MethodGet, path: "/orgs/{org}/members/{username}"},
	{method: http.MethodDelete, path: "/orgs/{org}/members/{username}"},
	{method: http.MethodGet, path: "/orgs/{org}/public_members"},
	{method: http.MethodGet, path: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodPut, path: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodDelete, path: "/orgs/{org}/public_members/{username}"},
	{method: http.MethodGet, path: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodPut, path: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodDelete, path: "/orgs/{org}/memberships/{username}"},
	{method: http.MethodGet, path: "/orgs/{org}/invitations"},
	{method: http.MethodGet, path: "/user/memberships/orgs"},
	{method: http.MethodGet, path: "/user/memberships/orgs/{org}"},
	{method: http.MethodPatch, path: "/user/memberships/orgs/{org}"},

	{method: http.MethodGet, path: "/orgs/{org}/outside_collaborators"},
	{method: http.MethodDelete, path: "/orgs/{org}/outside_collaborators/{username}"},
	{method: http.MethodPut, path: "/orgs/{org}/outside_collaborators/{username}"},

	{method: http.MethodGet, path: "/orgs/{org}/teams"},
	{method: http.MethodGet, path: "/teams/{id}"},
	{method: http.MethodPost, path: "/orgs/{org}/teams"},
	{method: http.MethodPatch, path: "/teams/{id}"},
	{method: http.MethodDelete, path: "/teams/{id}"},
	{method: http.MethodGet, path: "/teams/{id}/members/{username}"},
	{method: http.MethodPut, path: "/teams/{id}/members/{username}"},
	{method: http.MethodDelete, path: "/teams/{id}/members/{username}"},
	{method: http.MethodGet, path: "/teams/{id}/memberships/{username}"},
	{method: http.MethodPut, path: "/teams/{id}/memberships/{username}"},
	{method: http.MethodDelete, path: "/teams/{id}/memberships/{username}"},
	{method: http.MethodGet, path: "/teams/{id}/repos"},
	{method: http.MethodGet, path: "/teams/{id}/invitations"},
	//{method: http.MethodGet, path: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodPut, path: "/teams/{id}/repos/{org}/{repo}"},
	//{method: http.MethodDelete, path: "/teams/{id}/repos/{owner}/{repo}"},
	{method: http.MethodGet, path: "/orgs/{org}/hooks"},
	{method: http.MethodGet, path: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, path: "/orgs/{org}/hooks"},
	{method: http.MethodPatch, path: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodPost, path: "/orgs/{org}/hooks/{id}/pings"},
	{method: http.MethodDelete, path: "/orgs/{org}/hooks/{id}"},
	{method: http.MethodGet, path: "/orgs/{org}/blocks"},
	{method: http.MethodGet, path: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodPut, path: "/orgs/{org}/blocks/{username}"},
	{method: http.MethodDelete, path: "/orgs/{org}/blocks/{username}"},

	{method: http.MethodGet, path: "/projects/columns/{column_id}/cards"},
	{method: http.MethodGet, path: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, path: "/projects/columns/{column_id}/cards"},
	{method: http.MethodPatch, path: "/projects/columns/cards/{id}"},
	{method: http.MethodDelete, path: "/projects/columns/cards/{id}"},
	{method: http.MethodPost, path: "/projects/columns/cards/{id}/moves"},

	{method: http.MethodGet, path: "/projects/{project_id}/columns"},
	{method: http.MethodPost, path: "/projects/{project_id}/columns"},
	{method: http.MethodDelete, path: "/projects/columns/{id}"},
	{method: http.MethodPost, path: "/projects/columns/{id}/moves"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/comments"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/events"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/dismissals"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/comments"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{method: http.MethodDelete, path: "/reactions/{id}"},

	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodPatch, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{method: http.MethodGet, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPut, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodPost, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{method: http.MethodDelete, path: "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
}
