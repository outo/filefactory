package def

import (
	"io/ioutil"
	"os"
	"github.com/outo/filefactory/file"
)

var impl Implementation

func init() {
	ResetImplementation()
}

func GetProductionImplementation() Implementation {
	i := Implementation{
		//built-in
		OsLstat:         os.Lstat,
		OsReadlink:      os.Readlink,
		OsMkdirAll:      os.MkdirAll,
		OsSymlink:       os.Symlink,
		IoutilWriteFile: ioutil.WriteFile,
		IoutilReadFile:  ioutil.ReadFile,
		//custom
		MetaVerify: func(meta file.Meta, root string) error {
			return meta.Verify(root)
		},
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
	//builtin
	OsLstat         func(name string) (os.FileInfo, error)
	OsReadlink      func(name string) (string, error)
	OsMkdirAll      func(path string, perm os.FileMode) error
	OsSymlink       func(oldname string, newname string) error
	IoutilWriteFile func(filename string, data []byte, perm os.FileMode) error
	IoutilReadFile  func(filename string) ([]byte, error)
	//custom
	MetaVerify func(meta file.Meta, root string) error
}
