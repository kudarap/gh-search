package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrReqFailed    = errors.New("github: request failed")
	ErrRateLimitHit = errors.New("github: rate limit reached")
)

const (
	APIBaseURL           = "https://api.github.com"
	APIUserEndpoint      = "/user"
	APIRateLimitEndpoint = "/rate_limit"

	DefaultTimeout = 2 * time.Second
)

// RateLimit response header keys.
const (
	HeaderRateLimitLimit     = "x-ratelimit-limit"
	HeaderRateLimitRemaining = "x-ratelimit-remaining"
	HeaderRateLimitReset     = "x-ratelimit-reset"
	HeaderRateLimitUsed      = "x-ratelimit-used"
)

// Client represents Github's client service.
type Client struct {
	baseURL     string
	accessToken string

	// custom httpClient for controlled request and timeouts.
	httpClient *http.Client

	// requestGroup to prevent duplicate in-flight requests
	requestGroup singleflight.Group

	RateLimit RateLimit
}

// baseRequests sends GET requests and uses access token when available to increase rate limits.
func (c *Client) baseRequests(ctx context.Context, path string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.accessToken))
	}

	return c.httpClient.Do(req)
}

// NewClient initializes GitHub client and setup rate limits.
func NewClient(url string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
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
