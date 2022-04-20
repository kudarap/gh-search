package ghsearch

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/sync/errgroup"
)

var (
	// ErrTooManyInput indicates the input reached maximum allowed usernames.
	ErrTooManyInput = errors.New("too many username input")

	// ErrUserNotFound indicates user details can't be found from the source.
	ErrUserNotFound = errors.New("user not found")
)

// MaxUsersInputLength represents allowed maximum number of username input.
const MaxUsersInputLength = 10

// User represents a user details.
type User struct {
	Name        string `json:"name"`
	Login       string `json:"login"`
	Company     string `json:"company"`
	Followers   int    `json:"followers"`
	PublicRepos int    `json:"public_repos"`
}

// UserService provides access to user service.
type UserService interface {
	// Users returns a list of user base on usernames.
	Users(ctx context.Context, usernames []string) ([]*User, error)
}

// UserSource provides operation for retrieving user.
type UserSource interface {
	// User returns a user details from an external source.
	// Empty string username input will return immediately
	// with a nil User and error.
	User(ctx context.Context, username string) (*User, error)
}

// DefaultUserService represents a default implementation of user service.
type DefaultUserService struct {
	source UserSource
}

func (us *DefaultUserService) Users(ctx context.Context, usernames []string) ([]*User, error) {
	usernames = cleanUsernames(usernames)
	length := len(usernames)
	if length == 0 {
		return nil, nil
	}
	if length > MaxUsersInputLength {
		return nil, ErrTooManyInput
	}

	users := make([]*User, length)

	// Getting user details concurrently, since we already know the length of the input
	// its safe to process this way. when we have unknown number of input it might be
	// better to use a worker or semaphore.
	errG, ctx := errgroup.WithContext(ctx)
	for i, uname := range usernames {
		i, uname := i, uname // https://golang.org/doc/faq#closures_and_goroutines
		errG.Go(func() error {
			user, err := us.source.User(ctx, uname)
			if err != nil && !errors.Is(err, ErrUserNotFound) {
				return NewSourceError(err)
			}
			users[i] = user
			return nil
		})
	}
	if err := errG.Wait(); err != nil {
		return nil, err
	}

	return users, nil
}

// NewUserService return default user service.
func NewUserService(source UserSource) *DefaultUserService {
	return &DefaultUserService{source}
}

func cleanUsernames(usernames []string) []string {
	uu := make([]string, len(usernames))
	for i, u := range usernames {
		uu[i] = strings.TrimSpace(u)
	}
	return uu
}
