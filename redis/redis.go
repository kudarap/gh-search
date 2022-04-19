package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
)

const keyPrefix = "gh-search-"

var json = jsoniter.ConfigFastest

// Client represents Redis database client.
type Client struct {
	db *redis.Client
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

// Close closes database client connection.
func (c *Client) Close() error {
	return c.db.Close()
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
