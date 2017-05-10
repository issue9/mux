// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGithubAPI(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}

	srv := New(false, false, nil, nil)
	for _, api := range apis {
		srv.AddFunc(api.path, h, api.method)
	}

	w := httptest.NewRecorder()

	// 按照添加顺序，这应该是比较靠后的
	r, _ := http.NewRequest(http.MethodGet, "/repos/issue9/mux/branches/master/protection/enforce_admins", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(w, r)
	}
}

type api struct {
	method string
	path   string
}

var apis = []*api{
	{http.MethodGet, "/events"},
	{http.MethodGet, "/repos/{owner}/{repo}/events"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/events"},
	{http.MethodGet, "/networks/{owner}/{repo}/events"},
	{http.MethodGet, "/orgs/{org}/events"},
	{http.MethodGet, "/users/{username}/received_events"},
	{http.MethodGet, "/users/{username}/received_events/public"},
	{http.MethodGet, "/users/{username}/events"},
	{http.MethodGet, "/users/{username}/events/public"},
	{http.MethodGet, "/users/{username}/events/orgs/{org}"},
	{http.MethodGet, "/feeds"},
	{http.MethodGet, "/notifications"},
	{http.MethodGet, "/repos/{owner}/{repo}/notifications"},
	{http.MethodPut, "/notifications"},
	{http.MethodGet, "/notifications/threads/{id}"},
	{http.MethodPatch, "/notifications/threads/{id}"},
	{http.MethodGet, "/notifications/threads/{id}/subscription"},
	{http.MethodPut, "/notifications/threads/{id}/subscription"},
	{http.MethodDelete, "/notifications/threads/{id}/subscription"},
	{http.MethodGet, "/repos/{owner}/{repo}/stargazers"},
	{http.MethodGet, "/users/{username}/starred"},
	{http.MethodGet, "/user/starred"},
	{http.MethodGet, "/user/starred/{owner}/{repo}"},
	{http.MethodPut, "/user/starred/{owner}/{repo}"},
	{http.MethodDelete, "/user/starred/{owner}/{repo}"},
	{http.MethodGet, "/repos/{owner}/{repo}/subscribers"},
	{http.MethodGet, "/users/{username}/subscriptions"},
	{http.MethodGet, "/user/subscriptions"},
	{http.MethodGet, "/repos/{owner}/{repo}/subscription"},
	{http.MethodPut, "/repos/{owner}/{repo}/subscription"},
	{http.MethodDelete, "/repos/{owner}/{repo}/subscription"},
	{http.MethodGet, "/user/subscriptions/{owner}/{repo}"},
	{http.MethodPut, "/user/subscriptions/{owner}/{repo}"},
	{http.MethodDelete, "/user/subscriptions/{owner}/{repo}"},

	{http.MethodGet, "/gists/{gist_id}/comments"},
	{http.MethodGet, "/gists/{gist_id}/comments/{id}"},
	{http.MethodPost, "/gists/{gist_id}/comments"},
	{http.MethodPatch, "/gists/{gist_id}/comments/{id}"},
	{http.MethodDelete, "/gists/{gist_id}/comments/{id}"},

	{http.MethodGet, "/repos/{owner}/{repo}/git/blobs/{sha}"},
	{http.MethodPost, "/repos/{owner}/{repo}/git/blobs"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/commits/{sha}"},
	{http.MethodPost, "/repos/{owner}/{repo}/git/commits"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/commits/{sha}"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/refs/{ref}"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/refs/heads/feature"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/refs"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/refs/tags"},
	{http.MethodPost, "/repos/{owner}/{repo}/git/refs"},
	{http.MethodPatch, "/repos/{owner}/{repo}/git/refs/{ref}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/git/refs/{ref}"},

	{http.MethodGet, "/repos/{owner}/{repo}/git/tags/{sha}"},
	{http.MethodPost, "/repos/{owner}/{repo}/git/tags"},
	{http.MethodGet, "/repos/{owner}/{repo}/git/tags/{sha}"},

	{http.MethodGet, "/repos/{owner}/{repo}/git/trees/{sha}"},
	{http.MethodPost, "/repos/{owner}/{repo}/git/trees"},

	{http.MethodGet, "/integration/installations"},
	{http.MethodGet, "/integration/installations/{installation_id}"},
	{http.MethodGet, "/user/installations"},
	{http.MethodPost, "/installations/{installation_id}/access_tokens"},
	{http.MethodGet, "/installation/repositories"},
	{http.MethodGet, "/user/installations/{installation_id}/repositories"},
	{http.MethodPut, "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{http.MethodDelete, "/installations/{installation_id}/repositories/{repo}sitory_id"},
	{http.MethodGet, "/repos/{owner}/{repo}/assignees"},
	{http.MethodGet, "/repos/{owner}/{repo}/assignees/{assignee}"},
	{http.MethodPost, "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{http.MethodDelete, "/repos/{owner}/{repo}/issues/{number}/assignees"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/{number}/comments"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/comments"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/comments/{id}"},
	{http.MethodPost, "/repos/{owner}/{repo}/issues/{number}/comments"},
	{http.MethodPatch, "/repos/{owner}/{repo}/issues/comments/{id}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/issues/comments/{id}"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/{issue_number}/events"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/events"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/events/{id}"},
	{http.MethodGet, "/repos/{owner}/{repo}/labels"},
	{http.MethodGet, "/repos/{owner}/{repo}/labels/{name}"},
	{http.MethodPost, "/repos/{owner}/{repo}/labels"},
	{http.MethodPatch, "/repos/{owner}/{repo}/labels/{name}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/labels/{name}"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/{number}/labels"},
	{http.MethodPost, "/repos/{owner}/{repo}/issues/{number}/labels"},
	{http.MethodDelete, "/repos/{owner}/{repo}/issues/{number}/labels/{name}"},
	{http.MethodPut, "/repos/{owner}/{repo}/issues/{number}/labels"},
	{http.MethodDelete, "/repos/{owner}/{repo}/issues/{number}/labels"},
	{http.MethodGet, "/repos/{owner}/{repo}/milestones/{number}/labels"},

	{http.MethodGet, "/repos/{owner}/{repo}/milestones"},
	{http.MethodGet, "/repos/{owner}/{repo}/milestones/{number}"},
	{http.MethodPost, "/repos/{owner}/{repo}/milestones"},
	{http.MethodPatch, "/repos/{owner}/{repo}/milestones/{number}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/milestones/{number}"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/{issue_number}/timeline"},

	{http.MethodPost, "/orgs/{org}/migrations"},
	{http.MethodGet, "/orgs/{org}/migrations"},
	{http.MethodGet, "/orgs/{org}/migrations/{id}"},
	{http.MethodGet, "/orgs/{org}/migrations/{id}/archive"},
	{http.MethodDelete, "/orgs/{org}/migrations/{id}/archive"},
	{http.MethodDelete, "/orgs/{org}/migrations/{id}/repos/{repo}_name/lock"},

	{http.MethodPut, "/repos/{owner}/{repo}/import"},
	{http.MethodGet, "/repos/{owner}/{repo}/import"},
	{http.MethodPatch, "/repos/{owner}/{repo}/import"},
	{http.MethodGet, "/repos/{owner}/{repo}/import/authors"},
	{http.MethodPatch, "/repos/{owner}/{repo}/import/authors/{author_id}"},
	{http.MethodPatch, "/{owner}/{name}/import/lfs"},
	{http.MethodGet, "/{owner}/{name}/import/large_files"},
	{http.MethodDelete, "/repos/{owner}/{repo}/import"},
	{http.MethodGet, "/emojis"},
	{http.MethodGet, "/gitignore/templates"},
	{http.MethodGet, "/gitignore/templates/C"},

	// license
	{http.MethodGet, "/licenses"},
	{http.MethodGet, "/licenses/{license}"},
	{http.MethodGet, "/repos/{owner}/{repo}"},
	{http.MethodGet, "/repos/{owner}/{repo}/license"},

	{http.MethodPost, "/markdown"},
	{http.MethodPost, "/markdown/raw"},
	{http.MethodGet, "/meta"},
	{http.MethodGet, "/rate_limit"},

	{http.MethodGet, "/orgs/{org}/members"},
	{http.MethodGet, "/orgs/{org}/members/{username}"},
	{http.MethodDelete, "/orgs/{org}/members/{username}"},
	{http.MethodGet, "/orgs/{org}/public_members"},
	{http.MethodGet, "/orgs/{org}/public_members/{username}"},
	{http.MethodPut, "/orgs/{org}/public_members/{username}"},
	{http.MethodDelete, "/orgs/{org}/public_members/{username}"},
	{http.MethodGet, "/orgs/{org}/memberships/{username}"},
	{http.MethodPut, "/orgs/{org}/memberships/{username}"},
	{http.MethodDelete, "/orgs/{org}/memberships/{username}"},
	{http.MethodGet, "/orgs/{org}/invitations"},
	{http.MethodGet, "/user/memberships/orgs"},
	{http.MethodGet, "/user/memberships/orgs/{org}"},
	{http.MethodPatch, "/user/memberships/orgs/{org}"},

	{http.MethodGet, "/orgs/{org}/outside_collaborators"},
	{http.MethodDelete, "/orgs/{org}/outside_collaborators/{username}"},
	{http.MethodPut, "/orgs/{org}/outside_collaborators/{username}"},

	{http.MethodGet, "/orgs/{org}/teams"},
	{http.MethodGet, "/teams/{id}"},
	{http.MethodPost, "/orgs/{org}/teams"},
	{http.MethodPatch, "/teams/{id}"},
	{http.MethodDelete, "/teams/{id}"},
	{http.MethodGet, "/teams/{id}/members/{username}"},
	{http.MethodPut, "/teams/{id}/members/{username}"},
	{http.MethodDelete, "/teams/{id}/members/{username}"},
	{http.MethodGet, "/teams/{id}/memberships/{username}"},
	{http.MethodPut, "/teams/{id}/memberships/{username}"},
	{http.MethodDelete, "/teams/{id}/memberships/{username}"},
	{http.MethodGet, "/teams/{id}/repos"},
	{http.MethodGet, "/teams/{id}/invitations"},
	{http.MethodGet, "/teams/{id}/repos/{owner}/{repo}"},
	{http.MethodPut, "/teams/{id}/repos/{org}/{repo}"},
	{http.MethodDelete, "/teams/{id}/repos/{owner}/{repo}"},
	{http.MethodGet, "/orgs/{org}/hooks"},
	{http.MethodGet, "/orgs/{org}/hooks/{id}"},
	{http.MethodPost, "/orgs/{org}/hooks"},
	{http.MethodPatch, "/orgs/{org}/hooks/{id}"},
	{http.MethodPost, "/orgs/{org}/hooks/{id}/pings"},
	{http.MethodDelete, "/orgs/{org}/hooks/{id}"},
	{http.MethodGet, "/orgs/{org}/blocks"},
	{http.MethodGet, "/orgs/{org}/blocks/{username}"},
	{http.MethodPut, "/orgs/{org}/blocks/{username}"},
	{http.MethodDelete, "/orgs/{org}/blocks/{username}"},

	{http.MethodGet, "/projects/columns/{column_id}/cards"},
	{http.MethodGet, "/projects/columns/cards/{id}"},
	{http.MethodPost, "/projects/columns/{column_id}/cards"},
	{http.MethodPatch, "/projects/columns/cards/{id}"},
	{http.MethodDelete, "/projects/columns/cards/{id}"},
	{http.MethodPost, "/projects/columns/cards/{id}/moves"},

	{http.MethodGet, "/projects/{project_id}/columns"},
	{http.MethodPost, "/projects/{project_id}/columns"},
	{http.MethodDelete, "/projects/columns/{id}"},
	{http.MethodPost, "/projects/columns/{id}/moves"},

	{http.MethodGet, "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/comments"},
	{http.MethodPost, "/repos/{owner}/{repo}/pulls/{number}/reviews"},
	{http.MethodPost, "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/events"},
	{http.MethodPut, "/repos/{owner}/{repo}/pulls/{number}/reviews/{id}/dismissals"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/comments"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{http.MethodPost, "/repos/{owner}/{repo}/pulls/{number}/comments"},
	{http.MethodPatch, "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{http.MethodDelete, "/repos/{owner}/{repo}/pulls/comments/{id}"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{http.MethodPost, "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},
	{http.MethodDelete, "/repos/{owner}/{repo}/pulls/{number}/requested_reviewers"},

	{http.MethodGet, "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{http.MethodPost, "/repos/{owner}/{repo}/comments/{id}/reactions"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{http.MethodPost, "/repos/{owner}/{repo}/issues/{number}/reactions"},
	{http.MethodGet, "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{http.MethodPost, "/repos/{owner}/{repo}/issues/comments/{id}/reactions"},
	{http.MethodGet, "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{http.MethodPost, "/repos/{owner}/{repo}/pulls/comments/{id}/reactions"},
	{http.MethodDelete, "/reactions/{id}"},

	{http.MethodGet, "/repos/{owner}/{repo}/branches"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{http.MethodPut, "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection"},
	{http.MethodPatch, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{http.MethodPut, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{http.MethodPost, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{http.MethodPatch, "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{http.MethodPost, "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{http.MethodPut, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{http.MethodPost, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams"},
	{http.MethodGet, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{http.MethodPut, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{http.MethodPost, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
	{http.MethodDelete, "/repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users"},
}
