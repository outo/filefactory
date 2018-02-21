package filefactory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"errors"
	"path/filepath"
	"time"
	"os"
	"github.com/outo/filefactory"
	"github.com/outo/filefactory/def"
	"github.com/outo/filefactory/attr"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
)

var _ = Describe("examples", func() {

	Specify("simple case, but very explicit example with a lot of boilerplate within the test case, making it hard to read what happens", func() {

		//first create an instance of FileFactory using constructor function
		ff := filefactory.New()

		//define a regular file, directory and a symlink, use relative path
		fileDefinitions := ff.FilesToCreate(
			def.Reg("relative/path/of/regular-file"),
			def.Dir("relative/path/to/directory"),
			def.Sym("relative/path/to/symlink", "symlink/target/path"),
		)

		//we need some place to create the files (definitions are using relative paths)
		// below we are creating a temporary directory to accommodate them
		tempDirRoot, err := ioutil.TempDir("", "temporary-directory-just-for-this-test")
		if err != nil {
			panic(err)
		}
		//so we don't leave trash behind
		defer os.RemoveAll(tempDirRoot)

		//create files from the above definitions under root
		filefactory.CreateFiles(tempDirRoot, fileDefinitions...)

		// here some processing would occur which shouldn't change these files as a side effect

		//now let's verify files exists and are as declared
		err = filefactory.VerifyFiles(tempDirRoot, fileDefinitions...)
		if err != nil {
			panic(err)
		}

		// they do, because there was no error returned
	})

	Describe("given temporary root directory and fresh FileFactory instance created for each test, so that boilerplate is not getting in the way", func() {
		var (
			fileFactory filefactory.FileFactory
			tempRootDir string
		)

		abs := func(relPath string) (absPath string) {
			return filepath.Join(tempRootDir, relPath)
		}

		BeforeEach(func() {
			fileFactory = filefactory.New()

			var err error
			tempRootDir, err = ioutil.TempDir("", "fs-copy-test-")
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			//os.RemoveAll(tempRootDir)
		})

		Specify("create and verify multiple items with specific attributes", func() {
			fileDeclarations := fileFactory.FilesToCreate(
				def.Reg("relative/path/to/regular-file",
					attr.ModePerm(0765),
					attr.OtherGid(),
					attr.ModifiedTime(time.Now()),
					attr.Size(500),
					attr.Seed(7573453)),
				def.Reg("relative/path/to/another-regular-file",
					attr.ModePerm(0400),
					attr.AccessedTime(time.Now()),
					attr.Size(200),
					attr.Seed(27)),
				def.Dir("relative/path/to/directory",
					attr.ModePerm(0700),
					attr.AccessedTime(time.Now())),
				def.Sym("relative/path/to/symlink", "different/symlink/target/path",
					attr.OtherGid()),
			)

			err := filefactory.CreateFiles(tempRootDir, fileDeclarations...)
			Expect(err).ShouldNot(HaveOccurred())

			err = filefactory.VerifyFiles(tempRootDir, fileDeclarations...)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Describe("given a simple definition of a regular file, directory and symlink", func() {

			var fileDeclarations []file.File

			BeforeEach(func() {
				fileDeclarations = fileFactory.FilesToCreate(
					def.Reg("relative/path/to/regular-file"),
					def.Dir("relative/path/to/directory"),
					def.Sym("relative/path/to/symlink", "symlink/target/path"),
				)
			})

			Specify("create and verify real files exist. This is equivalent case to the very first example but without the boilerplate in the test case.", func() {

				//this will create defined files under specified root
				err := filefactory.CreateFiles(tempRootDir, fileDeclarations...)
				Expect(err).ShouldNot(HaveOccurred())

				//here you would do something that shouldn't/should change the files

				//this will verify that there are no real files corresponding to definitions
				err = filefactory.VerifyFiles(tempRootDir, fileDeclarations...)
				Expect(err).ShouldNot(HaveOccurred()) //just because I use Gomega but you can use anything else
			})

			Specify("verify real files don't exists", func() {

				//fresh temporary directory is used for each test case and filefactory.CreateFiles wasn't called so they definitely don't exist

				//this will verify that there are no real files corresponding to definitions
				err := filefactory.VerifyFiles(tempRootDir, fileDeclarations...)

				Expect(err).Should(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
				verErr := err.(*verify.Errors)
				Expect(verErr.CombinedFileDifference).To(Equal(diff.NotPresentOrNotAccessible))
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, filepath.Join(tempRootDir, "relative/path/to/regular-file")))
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, filepath.Join(tempRootDir, "relative/path/to/directory")))
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, filepath.Join(tempRootDir, "relative/path/to/symlink")))

				//or
				Expect(verErr.Errors).To(ConsistOf(
					verify.Error{FileDifference: diff.NotPresentOrNotAccessible, Path: filepath.Join(tempRootDir, "relative/path/to/regular-file"), Err: errors.New("file does not exist")},
					verify.Error{FileDifference: diff.NotPresentOrNotAccessible, Path: filepath.Join(tempRootDir, "relative/path/to/directory"), Err: errors.New("file does not exist")},
					verify.Error{FileDifference: diff.NotPresentOrNotAccessible, Path: filepath.Join(tempRootDir, "relative/path/to/symlink"), Err: errors.New("file does not exist")},
				))
			})

			Specify("demonstrate how to check a file does not exists (or is not accessible)", func() {
				filesToExpect := fileFactory.FilesToExpect(
					def.Reg("non/existent/file"),
				)

				err := filefactory.VerifyFiles(tempRootDir, filesToExpect...)

				//omitted nil and type check (can panic)

				verErr := err.(*verify.Errors)
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, abs("non/existent/file")))
			})
		})

		Specify("demonstrate the verification errors coming back. Induce some attribute verification errors by altering file definitions between create and verify invocations", func() {

			filesToCreate := fileFactory.FilesToCreate(
				def.Reg("relative/path/to/regular-file"),
				def.Dir("relative/path/to/directory"),
				def.Sym("relative/path/to/symlink", "symlink/target/path"),
			)

			err := filefactory.CreateFiles(tempRootDir, filesToCreate...)
			Expect(err).ShouldNot(HaveOccurred())

			//the following declarations are for the same file paths as above but with different attributes
			verifyDeclarations := fileFactory.FilesToExpect(
				def.Reg("relative/path/to/regular-file", attr.OtherGid(), attr.ModePerm(0765), attr.ModifiedTime(time.Now())),
				def.Dir("relative/path/to/directory", attr.ModePerm(0700), attr.AccessedTime(time.Now())),
				def.Sym("relative/path/to/symlink", "different/symlink/target/path"),
			)

			err = filefactory.VerifyFiles(tempRootDir, verifyDeclarations...)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
			verErr := err.(*verify.Errors)

			//this check gives me indication if particular type of error happened at least once for any of the declarations
			Expect(verErr.CombinedFileDifference & diff.ModePerm).To(Equal(diff.ModePerm))

			//but because each of the registered difference types takes a single bit we can say
			Expect(verErr.CombinedFileDifference & diff.ModePerm).ToNot(BeZero())

			//this tells me that each of the differences was detected
			Expect(verErr.CombinedFileDifference).To(Equal(diff.ModePerm | diff.ModTime | diff.AccTime | diff.Group | diff.LinkTarget))

			//and this is detailed list of the error type, path and original error itself (similar to os.PathError)
			Expect(verErr.Errors).To(HaveLen(6))
			Expect(verErr.HasDifference(diff.Group, abs("relative/path/to/regular-file"))).To(BeTrue())
			Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/regular-file"))).To(BeTrue())
			Expect(verErr.HasDifference(diff.ModTime, abs("relative/path/to/regular-file"))).To(BeTrue())
			Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/directory"))).To(BeTrue())
			Expect(verErr.HasDifference(diff.AccTime, abs("relative/path/to/directory"))).To(BeTrue())
			Expect(verErr.HasDifference(diff.LinkTarget, abs("relative/path/to/symlink"))).To(BeTrue())

			//also we can verify what differences were registered against a particular path
			Expect(verErr.DifferenceFor(abs("relative/path/to/directory"))).To(Equal(diff.ModePerm | diff.AccTime))
		})

		Specify("demonstrate file factory-level and file-level override for verification", func() {

			//create file factory which only verifies mode by default
			fileFactory = filefactory.New(verify.AllByDefault(false), verify.ModePerm(true))

			fileDeclarations := fileFactory.FilesToCreate(
				def.Reg("relative/path/to/regular-file"),
				def.Dir("relative/path/to/directory"),
				def.Sym("relative/path/to/symlink", "symlink/target/path"),
			)

			//these files will have incorrect attributes
			sameFilesButDifferentAttributes := fileFactory.FilesToExpect(
				def.Reg("relative/path/to/regular-file", attr.OtherGid(), attr.ModePerm(0765), attr.ModifiedTime(time.Now())),
				def.Dir("relative/path/to/directory", attr.ModePerm(0700), attr.AccessedTime(time.Now())),
				//for symlinks we have to override the the mode verification as it does not make sense
				def.Sym("relative/path/to/symlink", "different/symlink/target/path", verify.ModePerm(false)),
			)

			err := filefactory.CreateFiles(tempRootDir, fileDeclarations...)
			Expect(err).ShouldNot(HaveOccurred())

			err = filefactory.VerifyFiles(tempRootDir, sameFilesButDifferentAttributes...)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
			verErr := err.(*verify.Errors)

			// the only observable errors are the ones to do with regular file and directory's mode
			Expect(verErr.Errors).To(HaveLen(2))
			Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/regular-file")))
			Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/directory")))
		})

		Specify("not verifying mode permissions does not mean the mode type can be incompatible", func() {
			fileFactory = filefactory.New(verify.ModePerm(false))

			filesToCreate := fileFactory.FilesToCreate(
				def.Reg("something"),
				def.Reg("something-else"),
			)

			err := filefactory.CreateFiles(tempRootDir, filesToCreate...)
			Expect(err).ShouldNot(HaveOccurred())

			filesToExpect := fileFactory.FilesToExpect(
				def.Dir("something"),
				def.Sym("something-else", ""),
			)

			err = filefactory.VerifyFiles(tempRootDir, filesToExpect...)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))

			verErr := err.(*verify.Errors)
			Expect(verErr.Errors).To(HaveLen(2))
			Expect(verErr.HasDifference(diff.ModeType, abs("something")))
			Expect(verErr.HasDifference(diff.ModeType, abs("something-else")))
		})
	})
})
