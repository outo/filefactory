package path

import (
	"os"
)

var (
	impl = getProductionImplementation()
)

func init() {
	ResetImplementation()
}

func getProductionImplementation() Implementation {
	return Implementation{
		OsLstat: os.Lstat,
	}
}

//not recommended to tweak in production
func MockForTest(mocking func(modifyThis *Implementation)) {
	mocking(&impl)
}

func ResetImplementation() {
	impl = getProductionImplementation()
}

type Implementation struct {
	//builtin
	OsLstat func(path string) (info os.FileInfo, err error)
	//custom
}
