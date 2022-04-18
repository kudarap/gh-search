package github

import (
	"encoding/json"
	"errors"
	"golang.org/x/sync/singleflight"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrReqFailed = errors.New("github: request failed")
)

const (
	APIBaseURL     = "https://api.github.com"
	DefaultTimeout = 2 * time.Second
)

// RateLimit response headers.
const (
	HeaderRateLimitLimit     = "x-ratelimit-limit"
	HeaderRateLimitRemaining = "x-ratelimit-remaining"
	HeaderRateLimitReset     = "x-ratelimit-reset"
	HeaderRateLimitUsed      = "x-ratelimit-used"
)

// Client represents Github's client service.
type Client struct {
	baseURL string

	// custom httpClient for controlled request and timeouts.
	httpClient *http.Client

	// requestGroup to prevent duplicate in-flight requests
	requestGroup singleflight.Group

	RateLimit RateLimit
}

// NewClient initializes GitHub client and setup rate limits.
func NewClient(url string, timeout time.Duration) (*Client, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	var c Client
	c.baseURL = url
	c.httpClient = &http.Client{Timeout: timeout}
	return &c, nil
}

func decodeBody(r *http.Response, out interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, out); err != nil {
		return err
	}
	return nil
}

func responseHasError(r *http.Response) bool {
	return r.StatusCode >= 400 && r.StatusCode <= 500
}
