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
	// TODO: check rate limit
	// TODO: request once

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
	if responseHasError(resp) {
		return nil, ghsearch.ErrUserSourceFailed
	}

	// TODO: not concurrent safe
	c.RateLimit = rateLimitFrom(resp.Header)

	var ur UserResponse
	if err = decodeBody(resp, &ur); err != nil {
		return nil, err
	}
	u := ghsearch.User(ur)
	return &u, nil
}
