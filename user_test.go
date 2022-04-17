package ghsearch_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kudarap/ghsearch"
)

var errUserSourceCall = errors.New("error getting user")

func TestUserService_Users(t *testing.T) {
	testcases := []struct {
		name string
		// deps
		source ghsearch.UserSource
		// args
		usernames []string
		// returns
		want    []*ghsearch.User
		wantErr error
	}{
		{
			"single ok",
			&mockedUserSource{
				users: map[string]*ghsearch.User{
					"kudarap": {Name: "james"},
				},
			},
			[]string{"kudarap"},
			[]*ghsearch.User{
				{Name: "james"},
			},
			nil,
		},
		{
			"multi ok",
			&mockedUserSource{
				processDuration: time.Second,
				users: map[string]*ghsearch.User{
					"kudarap": {Name: "james"},
					"spec":    {Name: "spectre"},
					"dazz":    {Name: "dazzle"},
				},
			},
			[]string{"kudarap", "spec", "dazz"},
			[]*ghsearch.User{
				{Name: "james"},
				{Name: "spectre"},
				{Name: "dazzle"},
			},
			nil,
		},
		{
			"empty",
			&mockedUserSource{},
			nil,
			nil,
			nil,
		},
		{
			"white space input",
			&mockedUserSource{
				users: map[string]*ghsearch.User{
					"kudarap": {Name: "james"},
				},
			},
			[]string{"kudarap", " ", ""},
			[]*ghsearch.User{
				{Name: "james"},
				nil,
				nil,
			},
			nil,
		},
		{
			"more than 10",
			&mockedUserSource{},
			[]string{
				"kudarap",
				"spec",
				"dazz",
				"riki",
				"lina",
				"ogre",
				"kotl",
				"axe",
				"lion",
				"invo",
				"techies",
			},
			nil,
			ghsearch.ErrTooManyInput,
		},
		{
			"source some not found",
			&mockedUserSource{
				users: map[string]*ghsearch.User{
					"jugg": {Name: "juggernaut"},
					"dazz": {Name: "dazzle"},
				},
			},
			[]string{"jugg", "spec", "dazz"},
			[]*ghsearch.User{
				{Name: "juggernaut"},
				nil,
				{Name: "dazzle"},
			},
			nil,
		},
		{
			"source some has error",
			&mockedUserSource{
				users: map[string]*ghsearch.User{
					"jugg": {Name: "juggernaut"},
				},
				err: errUserSourceCall,
			},
			[]string{"jugg", "dazz"},
			nil,
			errUserSourceCall,
		},
		// source timed out
		// source failing
		// repeating values
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := ghsearch.NewUserService(tc.source)
			got, gotErr := svc.Users(ctx, tc.usernames)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\ngot: \n\t%#v \nwant: \n\t%#v", got, tc.want)
			}
			if gotErr != tc.wantErr {
				t.Errorf("err: %#v, want: %#v", gotErr, tc.wantErr)
			}
		})
	}
}

type mockedUserSource struct {
	users           map[string]*ghsearch.User
	err             error
	processDuration time.Duration
}

func (mus *mockedUserSource) User(ctx context.Context, username string) (*ghsearch.User, error) {
	time.Sleep(mus.processDuration)

	u, found := mus.users[username]
	if !found {
		err := ghsearch.ErrUserNotFound
		if mus.err != nil {
			err = mus.err
		}
		return nil, err
	}
	return u, nil
}
