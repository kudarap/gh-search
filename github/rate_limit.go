package github

import (
	"context"
	"errors"
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

// consume determines if we haven't reached the rate limit
// and also checks if it's good to send again base on reset time.
func (l *RateLimit) consume() error {
	if !l.ResetsAt.IsZero() && l.ResetsAt.Before(time.Now()) {
		return nil
	}
	if l.Remaining == 0 {
		return ErrRateLimitHit
	}

	l.Remaining--
	l.Used++
	return nil
}

func (l *RateLimit) updateFrom(h http.Header) {
	hrl := rateLimitFrom(h)
	if l.Remaining > hrl.Remaining {
		return
	}
	*l = hrl
}

// RequestRateLimit returns current core rate limit.
func (c *Client) RequestRateLimit() (*RateLimit, error) {
	resp, err := c.getRequest(context.Background(), APIRateLimitEndpoint)
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

// RateLimit returns RateLimit details from a response.
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
	rl.Used, _ = strconv.Atoi(h.Get(HeaderRateLimitUsed))
	resetsAt, _ := strconv.Atoi(h.Get(HeaderRateLimitReset))
	if resetsAt != 0 {
		ts, _ := strconv.ParseInt(h.Get(HeaderRateLimitReset), 10, 64)
		rl.ResetsAt = time.Unix(ts, 0)
	}
	return rl
}
