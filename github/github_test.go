package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/kudarap/ghsearch"
	"github.com/kudarap/ghsearch/github"
)

func TestUser(t *testing.T) {
	testcases := []struct {
		name string
		// deps
		testSrv *httptest.Server
		// args
		username string
		// returns
		want    *ghsearch.User
		wantErr error
	}{
		{
			"ok",
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, rawRespBody200)
			})),
			"kudarap",
			&ghsearch.User{
				Name:        "",
				Login:       "kudarap",
				Company:     "Openovate Labs",
				Followers:   5,
				PublicRepos: 38,
			},
			nil,
		},
		{
			"not found",
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, rawRespBody404)
			})),
			"kudarap",
			nil,
			ghsearch.ErrUserNotFound,
		},
		{
			"internal error",
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, rawRespBody500)
			})),
			"kudarap",
			nil,
			ghsearch.ErrUserSourceFailed,
		},
		{
			"timed out",
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second * 3)
			})),
			"kudarap",
			nil,
			ghsearch.ErrUserSourceTimeout,
		},
		//{
		//"rate limit reached",
		//httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//	resetsAt := strconv.FormatInt(time.Now().Unix(), 10)
		//	w.Header().Add(github.HeaderRatelimitLimit, "60")
		//	w.Header().Add(github.HeaderRatelimitRemaining, "0")
		//	w.Header().Add(github.HeaderRatelimitReset, resetsAt)
		//	w.Header().Add(github.HeaderRatelimitUsed, "60")
		//	w.WriteHeader(http.StatusInternalServerError)
		//	fmt.Fprintf(w, rawRespBody403Ratelimit)
		//})),
		//"kudarap",
		//nil,
		//ghsearch.ErrUserSourceFailed,
		//},
	}
	for _, tc := range testcases {
		ctx := context.Background()
		gcl := github.NewClient(tc.testSrv.URL)
		got, gotErr := gcl.User(ctx, tc.username)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("\ngot: \n\t%#v \nwant: \n\t%#v", got, tc.want)
		}
		if gotErr != tc.wantErr {
			t.Errorf("err: %#v, want: %#v", gotErr, tc.wantErr)
		}
	}
}

func setDefaultTestHeaders(h http.Header) {
	resetsAt := strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)
	h.Add(github.HeaderRatelimitReset, resetsAt)
	h.Add(github.HeaderRatelimitLimit, "60")
	h.Add(github.HeaderRatelimitRemaining, "59")
	h.Add(github.HeaderRatelimitUsed, "1")
}

const rawRespBody200 = `{
  "login": "kudarap",
  "id": 3943674,
  "node_id": "MDQ6VXNlcjM5NDM2NzQ=",
  "avatar_url": "https://avatars.githubusercontent.com/u/3943674?v=4",
  "gravatar_id": "",
  "url": "https://api.github.com/users/kudarap",
  "html_url": "https://github.com/kudarap",
  "followers_url": "https://api.github.com/users/kudarap/followers",
  "following_url": "https://api.github.com/users/kudarap/following{/other_user}",
  "gists_url": "https://api.github.com/users/kudarap/gists{/gist_id}",
  "starred_url": "https://api.github.com/users/kudarap/starred{/owner}{/repo}",
  "subscriptions_url": "https://api.github.com/users/kudarap/subscriptions",
  "organizations_url": "https://api.github.com/users/kudarap/orgs",
  "repos_url": "https://api.github.com/users/kudarap/repos",
  "events_url": "https://api.github.com/users/kudarap/events{/privacy}",
  "received_events_url": "https://api.github.com/users/kudarap/received_events",
  "type": "User",
  "site_admin": false,
  "name": null,
  "company": "Openovate Labs",
  "blog": "http://chiligarlic.com",
  "location": "Kaer Morhen",
  "email": null,
  "hireable": true,
  "bio": null,
  "twitter_username": null,
  "public_repos": 38,
  "public_gists": 11,
  "followers": 5,
  "following": 1,
  "created_at": "2013-03-22T17:53:23Z",
  "updated_at": "2022-04-01T14:14:04Z"
}`

const rawRespBody403Ratelimit = `{
  "message": "Rate Limit Exceeded",
}`

const rawRespBody404 = `{
  "message": "Not Found",
  "documentation_url": "https://docs.github.com/rest/reference/users#get-a-user"
}`

const rawRespBody500 = `{
  "message": "Internal Error"
}`
