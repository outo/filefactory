package user

import (
	"os/user"
)

var (
	impl = GetProductionImplementation()
)

func init() {
	ResetImplementation()
}

func GetProductionImplementation() Implementation {
	return Implementation{
		//built-in
		UserCurrent: user.Current,
		UserGroupIds: func(u *user.User) ([]string, error) {
			return u.GroupIds()
		},
		//custom
	}
}

//not recommended to tweak in production
func MockForTest(mocking func(modifyThis *Implementation)) {
	mocking(&impl)
}

func ResetImplementation() {
	impl = GetProductionImplementation()
}

type Implementation struct {
	//builtin
	UserCurrent  func() (*user.User, error)
	UserGroupIds func(u *user.User) ([]string, error)
	//custom
}
