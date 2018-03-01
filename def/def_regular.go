package def

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"bytes"
	"encoding/base64"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory"
	"github.com/outo/filefactory/attr"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
	"math"
)

type Regular struct {
	file.Meta
	Size int64
	Seed int64
}

func Reg(relPath string, extraFileSpecificAttributes ...interface{}) filefactory.DefinitionConstructor {
	return func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File {
		regular := Regular{}

		fileSpecificDefaults := []interface{}{
			attr.ModePerm(0666),
			attr.Size(20),
		}

		combined :=
			append(hardcodedFileFactoryDefaults,
				append(fileSpecificDefaults,
					append(extraFileFactoryDefaults,
					extraFileSpecificAttributes...
				)...
			)...
		)

		regular.Populate(relPath, combined...)

		for _, attribute := range combined {
			switch catt := attribute.(type) {
			case attr.Size:
				regular.Size = int64(catt)
			case attr.Seed:
				regular.Seed = int64(catt)
			}
		}

		//otherwise it is not a regular file
		regular.Mode &= ^os.ModeType

		return &regular
	}
}

func (f Regular) String() string {
	return fmt.Sprintf("%s %d", f.Meta.String(), f.Size)
}

func (f Regular) Create(root string) (err error) {
	path := filepath.Join(root, f.Path)
	dir := filepath.Dir(path)
	err = impl.OsMkdirAll(dir, 0777)
	if err != nil {
		return
	}

	return impl.IoutilWriteFile(path, ProvidePseudoRandomBytes(f.Size, f.Seed), f.Mode)
}

func ProvidePseudoRandomBytes(size, seed int64) (bs []byte) {
	rnd := rand.New(rand.NewSource(seed))
	bs = make([]byte, size)
	rnd.Read(bs)
	return
}

func (f Regular) Verify(root string) (err error) {
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

	absolutePath := filepath.Join(root, f.Path)
	fi, err := impl.OsLstat(absolutePath)
	if err != nil {
		return
	}

	if f.Should(verify.Size(true)) {
		if fi.Size() != f.Size {
			verr.Add(diff.Size, absolutePath, errors.New(fmt.Sprintf("expected %d, actual %d", f.Size, fi.Size())))
		}
	}

	if f.Should(verify.Contents(true)) {
		actualBytes, err := impl.IoutilReadFile(absolutePath)
		if err != nil {
			return err
		}
		expectedBytes := ProvidePseudoRandomBytes(f.Size, f.Seed)
		if !bytes.Equal(actualBytes, expectedBytes) {
			expectedBytesSampleLength := int(math.Min(50, float64(len(expectedBytes))))
			actualBytesSampleLength := int(math.Min(50, float64(len(actualBytes))))
			verr.Add(diff.Contents, absolutePath, errors.New(fmt.Sprintf("base64(bytes[:<=50]) for expected %s, actual %s",
				base64.StdEncoding.EncodeToString(expectedBytes[:expectedBytesSampleLength]),
				base64.StdEncoding.EncodeToString(actualBytes[:actualBytesSampleLength]),
			)))
		}
	}

	return verr.MapToNilIfNone()
}
