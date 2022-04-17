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

const MaxUsernameInput = 10

// User represents a user details.
type User struct {
	Name        string
	Login       string
	Company     string
	Followers   int
	PublicRepos int
}

// UserService provides access to user service.
type UserService interface {
	// Users returns a list of user base on usernames.
	Users(ctx context.Context, usernames []string) ([]*User, error)
}

// UserSource provides operation for retrieving user.
type UserSource interface {
	// User returns a user details.
	User(ctx context.Context, username string) (*User, error)
}

// DefaultUserService represents a default implementation of user service.
type DefaultUserService struct {
	source UserSource
}

func (us *DefaultUserService) Users(ctx context.Context, usernames []string) ([]*User, error) {
	length := len(usernames)
	if length == 0 {
		return nil, nil
	}
	if length > MaxUsernameInput {
		return nil, ErrTooManyInput
	}

	users := make([]*User, length)
	errG, ctx := errgroup.WithContext(ctx)
	for i, uname := range usernames {
		if strings.TrimSpace(uname) == "" {
			continue
		}

		i, uname := i, uname // https://golang.org/doc/faq#closures_and_goroutines
		errG.Go(func() error {
			user, err := us.source.User(ctx, uname)
			if err != nil && !errors.Is(err, ErrUserNotFound) {
				return err
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
