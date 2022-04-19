package redis

import (
	"context"
	"github.com/kudarap/ghsearch"
	"time"
)

const userCacheExpr = time.Minute * 2

// UserSourceCache represents user source cache with redis.
type UserSourceCache struct {
	cache   *Client
	userSrc ghsearch.UserSource
}

// NewUserSource creates new instance of user source with cache.
func NewUserSource(c *Client, us ghsearch.UserSource) *UserSourceCache {
	return &UserSourceCache{c, us}
}

func (c *UserSourceCache) User(ctx context.Context, username string) (*ghsearch.User, error) {
	// Check for cached user value.
	cached := &ghsearch.User{}
	hit, err := c.cache.Get(ctx, username, cached)
	if err != nil {
		return nil, err
	}
	if hit && cached != nil {
		return cached, nil
	}

	// Get a new user user value.
	user, err := c.userSrc.User(ctx, username)
	if err != nil {
		return nil, err
	}
	if err = c.cache.Set(ctx, username, user, userCacheExpr); err != nil {
		return nil, err
	}
	return user, nil
}
