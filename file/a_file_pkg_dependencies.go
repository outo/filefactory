package file

import (
	"os"
	"time"
	"github.com/outo/filefactory/dependencies/path"
)

var impl Implementation

func init() {
	ResetImplementation()
}

func getProductionImplementation() Implementation {
	return Implementation{
		//built-in
		OsChmod:   os.Chmod,
		OsChtimes: os.Chtimes,
		OsLchown:  os.Lchown,
		OsLstat:   os.Lstat,
		//custom
		WrapNewFromPath: NewFromPath,
		PathExists:      path.Exists,
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
	OsChmod   func(name string, mode os.FileMode) error
	OsChtimes func(name string, atime time.Time, mtime time.Time) error
	OsLchown  func(name string, uid int, gid int) error
	OsLstat   func(name string) (os.FileInfo, error)
	//custom
	WrapNewFromPath func(path string) (meta Meta, err error)
	PathExists      func(path string) (exists bool, err error)
}
