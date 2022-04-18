package github_test

import (
	"net/http"
	"net/http/httptest"
)

func newTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}
