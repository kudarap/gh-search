package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kudarap/ghsearch"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	APIBaseURL     = "https://api.github.com"
	defaultTimeout = 2 * time.Second
)

// Ratelimit response headers.
const (
	HeaderRatelimitLimit     = "x-Ratelimit-limit"
	HeaderRatelimitRemaining = "x-Ratelimit-remaining"
	HeaderRatelimitReset     = "x-Ratelimit-reset"
	HeaderRatelimitUsed      = "x-Ratelimit-used"
)

// Client represents GitHub's client service.
type Client struct {
	baseURL    string
	httpClient *http.Client
	// api key
	// group requests
	// caching
	Ratelimit Ratelimit
}

// Ratelimit represents GitHub rate limits.
type Ratelimit struct {
	Limit     int
	Remaining int
	Used      int
	ResetsAt  time.Time
}

func (l Ratelimit) Reached() bool {
	return l.Remaining == 0
}

func NewClient(url string) *Client {
	var c Client
	c.baseURL = url
	c.httpClient = &http.Client{Timeout: defaultTimeout}
	return &c
}

func (c *Client) User(ctx context.Context, username string) (*ghsearch.User, error) {
	// check rate limit
	// request once

	url := fmt.Sprintf("%s/users/%s", c.baseURL, username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, ghsearch.ErrUserSourceTimeout
		}
		return nil, ghsearch.ErrUserSourceFailed
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ghsearch.ErrUserNotFound
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		return nil, ghsearch.ErrUserSourceFailed
	}

	c.updateRatelimit(resp.Header)

	var u ghsearch.User
	if err = responseEncoder(resp, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *Client) updateRatelimit(h http.Header) {
	c.Ratelimit.Limit, _ = strconv.Atoi(h.Get(HeaderRatelimitLimit))
	c.Ratelimit.Remaining, _ = strconv.Atoi(h.Get(HeaderRatelimitRemaining))
	c.Ratelimit.Remaining, _ = strconv.Atoi(h.Get(HeaderRatelimitRemaining))
	ts, _ := strconv.ParseInt(h.Get(HeaderRatelimitReset), 10, 64)
	c.Ratelimit.ResetsAt = time.Unix(ts, 0)
}

func responseEncoder(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}
