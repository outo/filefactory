package filefactory

import (
	"time"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/attr"
	"github.com/outo/filefactory/verify"
)

const MsgConstructorDidNotDoItsJob = "whoops, constructor is meant to create a wrapper in any case"

//Primarily, I wanted to have means to declaratively setup and verify real file structure and it would
// be very inconvenient and harder to read if I had to provide matching times to declarations
// where times don't really matter.
// The notion is that files created with a given instance of a FileFactory would carry the same defaults (incl. times),
// unless alternative attribute values were passed into file declaration.
//Create another instance of FileFactory and these files will have different default timestamp.
//So it is important to remember that file definitions and file assertions have to go through
// FileFactory.FilesToCreate and FileFactory.FilesToExpect within the same factory instance.
type FileFactory struct {
	hardcodedFileFactoryDefaults,
	extraFileFactoryDefaults []interface{}
}

type DefinitionConstructor func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File

func New(extraFileFactoryDefaults ...interface{}) (ff FileFactory) {
	//pre-pending some default values, which will be overwritten in case the variadic type already has them
	hardcodedFileFactoryDefaults := []interface{}{
		attr.CurrentUid(),
		attr.PrimaryGid(),
		attr.ModifiedTime(time.Now()),
		attr.AccessedTime(time.Now().Add(90 * time.Minute).Add(15 * time.Second)), //added some more time as it is easier to spot than nanoseconds difference
	}

	return FileFactory{
		hardcodedFileFactoryDefaults: hardcodedFileFactoryDefaults,
		extraFileFactoryDefaults:     extraFileFactoryDefaults,
	}
}

func (ff FileFactory) FilesToCreate(constructors ...DefinitionConstructor) (files []file.File) {

	for _, constructor := range constructors {
		f := constructor(ff.hardcodedFileFactoryDefaults, ff.extraFileFactoryDefaults)
		if f == nil {
			panic(MsgConstructorDidNotDoItsJob)
		}
		files = append(files, f)
	}
	return files
}

//an alias method for FileFactory.FilesToCreate
func (ff FileFactory) FilesToExpect(constructors ...DefinitionConstructor) (file []file.File) {
	return ff.FilesToCreate(constructors...)
}

//Ideally, at most one invocation per single test as per the explanation in this function.
//Otherwise, care is advised as modified timestamp may be overwritten by the system.
func CreateFiles(root string, files ...file.File) (err error) {
	for _, f := range files {
		err = f.Create(root)
		if err != nil {
			return
		}
	}

	//Split into two loops, so that attributes (especially timestamps) are aligned only once all files are created.
	//You could have a directory created and attributes aligned, and then you may need to create a file within this directory.
	//Doing that will update (on NIXes) modified and change timestamps on the directory itself which means the
	// modified timestamp will be updated with current time.
	for _, f := range files {
		err = f.AlignAttributes(true, true, true, root)
		if err != nil {
			return
		}
	}
	return
}

func VerifyFiles(root string, expectedFiles ...file.File) (err error) {
	aggregatedVerificationErrors := verify.Errors{}
	for _, f := range expectedFiles {
		err := f.Verify(root)
		err = aggregatedVerificationErrors.Merge(err)
		if err != nil {
			return err
		}
	}
	return aggregatedVerificationErrors.MapToNilIfNone()
}

