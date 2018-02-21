package def

import (
	"os"
	"path/filepath"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory"
	"github.com/outo/filefactory/attr"
)

type Directory struct {
	file.Meta
}

func Dir(relPath string, extraFileSpecificAttributes ...interface{}) filefactory.DefinitionConstructor {
	return func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File {
		directory := Directory{}

		fileSpecificDefaults := []interface{}{
			attr.ModePerm(0777),
		}

		combined :=
			append(hardcodedFileFactoryDefaults,
				append(fileSpecificDefaults,
					append(extraFileFactoryDefaults,
						extraFileSpecificAttributes...
					)...
				)...
			)

		directory.Populate(relPath, combined...)

		//has to be if it is a directory
		directory.Mode &= ^os.ModeType
		directory.Mode |= os.ModeDir

		return &directory
	}
}

func (f Directory) Create(root string) (err error) {
	path := filepath.Join(root, f.Path)

	dir := filepath.Dir(path)
	err = impl.OsMkdirAll(dir, 0777)
	if err != nil {
		return
	}

	return impl.OsMkdirAll(path, f.Mode)
}

func (f Directory) Verify(root string) (err error) {
	return impl.MetaVerify(f.Meta, root)
}
