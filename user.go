package ghsearch

import (
	"errors"
	"strings"
)

// ErrTooManyInput indicates the input reached maximum allowed usernames.
var ErrTooManyInput = errors.New("too many username input")

const maxUsernameInput = 10

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
	Users(usernames []string) ([]*User, error)
}

// UserSource provides operation for retrieving user.
type UserSource interface {
	// User returns a user details.
	User(username string) (*User, error)
}

// DefaultUserService represents a default implementation of user service.
type DefaultUserService struct {
	source UserSource
}

// NewUserService return default user service.
func NewUserService(source UserSource) *DefaultUserService {
	return &DefaultUserService{source}
}

func (us *DefaultUserService) Users(usernames []string) ([]*User, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	if len(usernames) > maxUsernameInput {
		return nil, ErrTooManyInput
	}

	var users []*User
	for _, u := range usernames {
		if strings.TrimSpace(u) == "" {
			continue
		}

		user, err := us.source.User(u)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}
