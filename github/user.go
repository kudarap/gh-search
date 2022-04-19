package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kudarap/ghsearch"
)

// UserResponse represents github user data.
type UserResponse ghsearch.User

// User returns Github user details by username.
func (c *Client) User(ctx context.Context, username string) (*ghsearch.User, error) {
	if err := c.RateLimit.check(); err != nil {
		return nil, err
	}

	v, err, _ := c.requestGroup.Do("username", func() (interface{}, error) {
		// TODO invalidate cache
		return c.fetchUser(ctx, username)
	})
	if err != nil {
		return nil, err
	}

	ur := v.(*UserResponse)
	u := ghsearch.User(*ur)
	return &u, nil
}

func (c *Client) fetchUser(ctx context.Context, username string) (*UserResponse, error) {
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

	var ur UserResponse
	if err = decodeBody(resp, &ur); err != nil {
		return nil, err
	}
	return &ur, err
}
