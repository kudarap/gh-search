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

func TestClient_User(t *testing.T) {
	testcases := []struct {
		name string
		// deps
		testSrv *httptest.Server
		timeout time.Duration
		// args
		username string
		// returns
		want    *ghsearch.User
		wantErr error
	}{
		{
			"ok",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, rawRespBodyUser)
			}),
			0,
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
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, rawRespBody404)
			}),
			0,
			"kudarap",
			nil,
			ghsearch.ErrUserNotFound,
		},
		{
			"internal error",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				setDefaultTestHeaders(w.Header())
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, rawRespBody500)
			}),
			0,
			"kudarap",
			nil,
			ghsearch.ErrUserSourceFailed,
		},
		{
			"timed out",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second)
			}),
			time.Second / 2,
			"kudarap",
			nil,
			ghsearch.ErrUserSourceTimeout,
		},
	}
	for _, tc := range testcases {
		ctx := context.Background()
		gcl, err := github.NewClient(tc.testSrv.URL, tc.timeout)
		if err != nil {
			t.Errorf("github.NewClient should not error: %s", err)
			t.FailNow()
		}

		gcl.RateLimit = github.RateLimit{
			Limit:     60,
			Remaining: 0,
			Used:      0,
			ResetsAt:  time.Now(),
		}
		got, gotErr := gcl.User(ctx, tc.username)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("\ngot: \n\t%#v \nwant: \n\t%#v", got, tc.want)
		}
		if gotErr != tc.wantErr {
			t.Errorf("err: %#v, want: %#v", gotErr, tc.wantErr)
		}
	}
}

func TestClient_User_SingleFlight(t *testing.T) {

}

func setDefaultTestHeaders(h http.Header) {
	h.Add(github.HeaderRateLimitLimit, "60")
	h.Add(github.HeaderRateLimitRemaining, "59")
	h.Add(github.HeaderRateLimitUsed, "1")
	h.Add(github.HeaderRateLimitReset, strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
}

const rawRespBodyUser = `{
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

const rawRespBody403RateLimit = `{
  "message": "Rate Limit Exceeded",
}`

const rawRespBody404 = `{
  "message": "Not Found",
  "documentation_url": "https://docs.github.com/rest/reference/users#get-a-user"
}`

const rawRespBody500 = `{
  "message": "Internal Error"
}`
