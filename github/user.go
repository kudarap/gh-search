package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kudarap/ghsearch"
)

const requestGroupKeyExpr = time.Second * 2

// User returns Github user details by username.
func (c *Client) User(ctx context.Context, username string) (*ghsearch.User, error) {
	if err := c.RateLimit.check(); err != nil {
		return nil, err
	}

	// avoid duplicate inflight requests.
	v, err, _ := c.requestGroup.Do(username, func() (interface{}, error) {
		// invalidates share result after requestGroupKeyExpr.
		go func() {
			time.Sleep(requestGroupKeyExpr)
			c.requestGroup.Forget(username)
		}()
		return c.fetchUser(ctx, username)
	})
	if err != nil {
		return nil, err
	}

	return v.(*ghsearch.User), nil
}

func (c *Client) fetchUser(ctx context.Context, username string) (*ghsearch.User, error) {
	url := fmt.Sprintf("%s/%s", APIUserEndpoint, username)
	resp, err := c.baseRequests(ctx, url)
	if err != nil {
		if os.IsTimeout(err) {
			return nil, ghsearch.ErrUserSourceTimeout
		}
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ghsearch.ErrUserNotFound
	}
	if responseHasError(resp) {
		return nil, ghsearch.ErrUserSourceFailed
	}

	// TODO: not concurrent safe
	c.RateLimit = rateLimitFrom(resp.Header)

	var u ghsearch.User
	if err = decodeBody(resp, &u); err != nil {
		return nil, err
	}
	return &u, err
}
