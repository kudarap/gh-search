package github

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RateLimit represents github rate limit data.
type RateLimit struct {
	Limit     int
	Remaining int
	Used      int
	ResetsAt  time.Time
}

func (l RateLimit) reached() bool {
	return l.Remaining == 0
}

// RequestRateLimit returns current core rate limit.
func (c *Client) RequestRateLimit() (*RateLimit, error) {
	url := fmt.Sprintf("%s/rate_limit", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if responseHasError(resp) {
		return nil, ErrReqFailed
	}

	var r RateLimitResponse
	if err = decodeBody(resp, &r); err != nil {
		return nil, err
	}
	rl := r.RateLimit()
	return rl, nil
}

func (c *Client) acquireRateLimit() error {
	// Retrieves current rate limits before sending requests.
	rl, err := c.RequestRateLimit()
	if err != nil {
		return err
	}
	if rl == nil {
		return errors.New("empty rate limit response")
	}

	c.RateLimit = *rl
	return nil
}

// RateLimitResponse represents Github's rate limit resource.
type RateLimitResponse struct {
	Resources struct {
		Core struct {
			Limit     int
			Remaining int
			Used      int
			Reset     int64
		}
	}
}

func (r RateLimitResponse) RateLimit() *RateLimit {
	c := r.Resources.Core
	return &RateLimit{
		Limit:     c.Limit,
		Remaining: c.Remaining,
		Used:      c.Used,
		ResetsAt:  time.Unix(c.Reset, 0),
	}
}

func rateLimitFrom(h http.Header) RateLimit {
	var rl RateLimit
	rl.Limit, _ = strconv.Atoi(h.Get(HeaderRateLimitLimit))
	rl.Remaining, _ = strconv.Atoi(h.Get(HeaderRateLimitRemaining))
	rl.Remaining, _ = strconv.Atoi(h.Get(HeaderRateLimitRemaining))
	ts, _ := strconv.ParseInt(h.Get(HeaderRateLimitReset), 10, 64)
	rl.ResetsAt = time.Unix(ts, 0)
	return rl
}