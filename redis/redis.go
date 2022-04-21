package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

const keyPrefix = "gh-search-"

// Client represents Redis database client.
type Client struct {
	db *redis.Client
}

func (c *Client) Get(ctx context.Context, key string, out interface{}) (ok bool, err error) {
	val, err := c.db.Get(ctx, keyPrefix+key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if err = json.Unmarshal([]byte(val), out); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) Set(ctx context.Context, key string, val interface{}, expr time.Duration) error {
	// Skip caching when key and value is empty.
	if key == "" || val == nil {
		return nil
	}

	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return c.db.Set(ctx, keyPrefix+key, string(b), expr).Err()
}

func (c *Client) Close() error {
	return c.db.Close()
}

// NewClient returns a new Redis client.
func NewClient(url string) (*Client, error) {
	ctx := context.Background()

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb}, nil
}
