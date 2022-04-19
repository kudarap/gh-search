package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"
)

var (
	ErrReqFailed    = errors.New("github: request failed")
	ErrRateLimitHit = errors.New("github: rate limit reached")
)

const (
	APIBaseURL           = "https://api.github.com"
	APIUserEndpoint      = "/users"
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

// getRequest sends GET request and uses access token when available to increase rate limits.
func (c *Client) getRequest(ctx context.Context, path string) (*http.Response, error) {
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
func NewClient(accessToken string) (*Client, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, errors.New("access token required")
	}

	c := NewCustomClient(APIBaseURL, accessToken, DefaultTimeout)
	if err := c.acquireRateLimit(); err != nil {
		return nil, err
	}
	return c, nil
}

func NewCustomClient(url, accessToken string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	var c Client
	c.baseURL = url
	c.accessToken = accessToken
	c.httpClient = &http.Client{Timeout: timeout}
	return &c
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
