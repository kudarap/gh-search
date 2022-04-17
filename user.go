package ghsearch

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
	// User returns user details.
	User(username string) (*User, error)
}

type userService struct {
}

func (u *userService) Users(usernames []string) ([]*User, error) {
	//TODO implement me
	panic("implement me")
}
