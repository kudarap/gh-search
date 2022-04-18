package github_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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
			gh, err := github.NewClient(tc.testSrv.URL, time.Second)
			if err != nil {
				t.Errorf("github.NewClient should not error: %s", err)
				t.FailNow()
			}

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
