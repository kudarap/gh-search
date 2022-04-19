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

	"github.com/kudarap/ghsearch/github"
)

func TestClient_RequestRateLimit(t *testing.T) {
	testcases := []struct {
		name string
		// deps
		testSrv *httptest.Server
		// returns
		want    *github.RateLimit
		wantErr error
	}{
		{
			"rate limit acquired",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, rawRespBodyRateLimit)
			}),
			&github.RateLimit{
				Limit:     60,
				Remaining: 60,
				Used:      0,
				ResetsAt:  time.Unix(1650240000, 0),
			},
			nil,
		},
		{
			"request rate limit failed",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, rawRespBody404)
			}),
			nil,
			github.ErrReqFailed,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gh := github.NewCustomClient(tc.testSrv.URL, "", time.Second)
			got, gotErr := gh.RequestRateLimit()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\ngot: \n\t%#v \nwant: \n\t%#v", got, tc.want)
			}
			if gotErr != tc.wantErr {
				t.Errorf("err: %#v, want: %#v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestClient_User_RateLimitCheck(t *testing.T) {
	now := time.Now()
	testcases := []struct {
		name string
		// deps
		testSrv *httptest.Server
		// client state
		current github.RateLimit
		want    github.RateLimit
		// returns
		wantErr error
	}{
		{
			"rate limit ok",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add(github.HeaderRateLimitLimit, "60")
				w.Header().Add(github.HeaderRateLimitRemaining, "59")
				w.Header().Add(github.HeaderRateLimitUsed, "1")
				w.Header().Add(github.HeaderRateLimitReset, strconv.FormatInt(now.Unix(), 10))
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, rawRespBodyUser)
			}),
			github.RateLimit{
				Limit:     60,
				Remaining: 60,
				Used:      0,
			},
			github.RateLimit{
				Limit:     60,
				Remaining: 59,
				Used:      1,
				ResetsAt:  time.Unix(now.Unix(), 0),
			},
			nil,
		},
		{
			"rate limit reached",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, rawRespBody403RateLimit)
			}),
			github.RateLimit{
				Limit:     0,
				Remaining: 0,
				Used:      0,
			},
			github.RateLimit{},
			github.ErrRateLimitHit,
		},
		{
			"rate limit resets",
			newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add(github.HeaderRateLimitLimit, "60")
				w.Header().Add(github.HeaderRateLimitRemaining, "59")
				w.Header().Add(github.HeaderRateLimitUsed, "1")
				w.Header().Add(github.HeaderRateLimitReset, strconv.FormatInt(now.Add(time.Minute).Unix(), 10))
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, rawRespBodyUser)
			}),
			github.RateLimit{
				Limit:     0,
				Remaining: 0,
				Used:      0,
				ResetsAt:  now.Add(-time.Minute),
			},
			github.RateLimit{
				Limit:     60,
				Remaining: 59,
				Used:      1,
				ResetsAt:  time.Unix(now.Add(time.Minute).Unix(), 0),
			},
			nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			client := github.NewCustomClient(tc.testSrv.URL, "", 0)
			client.RateLimit = tc.current

			ctx := context.Background()
			_, gotErr := client.User(ctx, "kudarap")
			if !reflect.DeepEqual(client.RateLimit, tc.want) {
				t.Errorf("\ngot: \n\t%#v \nwant: \n\t%#v", client.RateLimit, tc.want)
			}
			if gotErr != tc.wantErr {
				t.Errorf("err: %#v, want: %#v", gotErr, tc.wantErr)
			}
		})
	}
}

const rawRespBodyRateLimit = `{
  "resources": {
    "core": {
      "limit": 60,
      "remaining": 60,
      "reset": 1650240000,
      "used": 0,
      "resource": "core"
    },
    "graphql": {
      "limit": 0,
      "remaining": 0,
      "reset": 1650284278,
      "used": 0,
      "resource": "graphql"
    },
    "integration_manifest": {
      "limit": 5000,
      "remaining": 5000,
      "reset": 1650284278,
      "used": 0,
      "resource": "integration_manifest"
    },
    "search": {
      "limit": 10,
      "remaining": 10,
      "reset": 1650280738,
      "used": 0,
      "resource": "search"
    }
  },
  "rate": {
    "limit": 60,
    "remaining": 53,
    "reset": 1650283741,
    "used": 7,
    "resource": "core"
  }
}
`
