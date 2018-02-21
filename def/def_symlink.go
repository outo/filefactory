package def

import (
	"errors"
	"fmt"
	"path/filepath"
	"os"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
)

type Symlink struct {
	file.Meta
	LinkTarget string
}

func Sym(relPath string, linkTarget string, extraFileSpecificAttributes ...interface{}) filefactory.DefinitionConstructor {
	return func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File {
		symlink := Symlink{
			LinkTarget: linkTarget,
		}

		fileSpecificDefaults := []interface{}{
			verify.ModePerm(false),
			verify.ModifiedTime(false),
			verify.AccessedTime(false),
		}

		combined :=
			append(hardcodedFileFactoryDefaults,
				append(fileSpecificDefaults,
					append(extraFileFactoryDefaults,
						extraFileSpecificAttributes...
					)...
				)...
			)

		symlink.Populate(relPath, combined...)

		//a must for a symlink
		symlink.Mode &= ^os.ModeType
		symlink.Mode |= os.ModeSymlink

		return &symlink
	}
}

func (f Symlink) String() string {
	return fmt.Sprintf("%s -> %s", f.Meta.String(), f.LinkTarget)
}

func (f Symlink) Create(root string) (err error) {
	path := filepath.Join(root, f.Path)

	dir := filepath.Dir(path)
	err = impl.OsMkdirAll(dir, 0777)
	if err != nil {
		return
	}

	return impl.OsSymlink(f.LinkTarget, path)
}

func (f Symlink) Verify(root string) (err error) {
	verr := &verify.Errors{}

	err = impl.MetaVerify(f.Meta, root)
	if err = verr.Merge(err); err != nil {
		return
	}

	if verr.IsFileNotPresentOrNotAccessible() {
		return verr
	} else if verr.IsFileTypeUnexpected() {
		return verr
	}

	path := filepath.Join(root, f.Path)
	linkTarget, err := impl.OsReadlink(path)
	if err != nil {
		return
	}

	if f.Should(verify.SymlinkTarget(true)) {
		if linkTarget != f.LinkTarget {
			verr.Add(diff.LinkTarget, path, errors.New(fmt.Sprintf("expected %s, actual %s", f.LinkTarget, linkTarget)))
		}
	}

	return verr.MapToNilIfNone()
}
