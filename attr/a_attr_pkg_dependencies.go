package attr

import (
	"github.com/outo/filefactory/dependencies/user"
)

var impl Implementation

func init() {
	ResetImplementation()
}

func GetProductionImplementation() Implementation {
	i := Implementation{
		UserCurrentIds: user.CurrentIds,
	}
	return i
}

//not recommended to tweak in production
func MockForTest(mocking func(modifyThis *Implementation)) {
	mocking(&impl)
}

func ResetImplementation() {
	impl = GetProductionImplementation()
}

type Implementation struct {
	UserCurrentIds func() (uid uint32, primaryGid uint32, otherGid uint32, err error)
}
