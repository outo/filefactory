package def_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/outo/filefactory/testingaids/mock"
	"os"
	"path/filepath"
	"time"
	"github.com/outo/filefactory/def"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
	"github.com/outo/filefactory/attr"
)

var _ = Describe("pkg def file_regular.go unit test", func() {

	var (
		noError,
		anError error
	)

	BeforeEach(func() {
		def.ResetImplementation()
		anError = errors.New("just an error, not significant what it is")
		def.MockForTest(func(modifyThis *def.Implementation) {
			modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
				return noError
			}
			modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
				return mock.NewFileInfo().WithSize(20), noError
			}
			modifyThis.IoutilReadFile = func(filename string) ([]byte, error) {
				return def.ProvidePseudoRandomBytes(20, 18), noError
			}
		})
	})

	Describe("Regular.Verify", func() {
		const expectedRoot = "/an/example/root"

		It("will invoke Meta.Verify with root parameter and indication all three attributes needs verifying", func() {
			actualRoot := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					actualRoot = root
					return anError
				}
			})

			regular := def.Regular{}

			regular.Verify(expectedRoot)
			Expect(actualRoot).To(Equal(expectedRoot))
		})

		It("will return Meta.Verify error immediately if of type other than VerificationErrors", func() {
			invoked := false
			expectedError := errors.New("Meta.Verify error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return expectedError
				}
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					invoked = true
					return nil, anError
				}
			})

			regular := def.Regular{}
			actualError := regular.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
			Expect(invoked).To(BeFalse())
		})
		It("will not return Meta.Verify error immediately if of type VerificationErrors (it contains list of verification specific failures). It will continue to os.Lstat", func() {
			invoked := false
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return &verify.Errors{}
				}
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					invoked = true
					return nil, anError
				}
			})

			regular := def.Regular{}
			regular.Verify(expectedRoot)
			Expect(invoked).To(BeTrue())
		})

		It("will not continue verification if file is not present or is inaccessible", func() {
			invoked := false
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					verErr := verify.Errors{}
					verErr.Add(diff.NotPresentOrNotAccessible, "some path", anError)
					return &verErr
				}
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					invoked = true
					return nil, anError
				}
			})

			regular := def.Regular{}
			regular.Verify(expectedRoot)
			Expect(invoked).To(BeFalse())
		})

		It("will not continue verification if file type is unexpected (e.g. found dir when it should be a regular file)", func() {
			invoked := false
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					verErr := verify.Errors{}
					verErr.Add(diff.ModeType, "some path", anError)
					return &verErr
				}
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					invoked = true
					return nil, anError
				}
			})

			regular := def.Regular{}
			regular.Verify(expectedRoot)
			Expect(invoked).To(BeFalse())
		})

		It("will invoke os.Lstat with a path joined to root parameter", func() {
			actualPath := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					actualPath = name
					return nil, anError
				}
			})

			regular := def.Regular{}
			regular.Path = "expected/relative/path/to/file"

			regular.Verify(expectedRoot)
			Expect(actualPath).To(Equal(filepath.Join(expectedRoot, regular.Path)))
		})
		It("will return os.Lstat error immediately", func() {
			expectedError := errors.New("os.Lstat error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					return nil, expectedError
				}
			})

			regular := def.Regular{}
			actualError := regular.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
		})
		It("will append error to VerificationErrors if size is different", func() {
			retrievedSize := int64(12345)
			mockedFileInfo := mock.NewFileInfo().WithSize(retrievedSize)

			nestedError := errors.New("modTime error")
			verificationErrors := verify.Errors{}
			verificationErrors.Add(diff.ModTime, "some path", nestedError)

			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return &verificationErrors
				}
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					return mockedFileInfo, noError
				}
			})

			regular := def.Regular{}
			regular.Path = "file-with-different-size"
			regular.Size = 20
			regular.Seed = 18
			actualError := regular.Verify(expectedRoot)
			Expect(actualError).Should(HaveOccurred())
			Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
			actualVerificationErrors := actualError.(*verify.Errors)
			Expect(actualVerificationErrors.Errors).To(HaveLen(2))
			Expect(actualVerificationErrors.CombinedFileDifference).To(Equal(diff.ModTime | diff.Size))
			Expect(actualVerificationErrors.HasDifference(diff.ModTime, "some path")).To(BeTrue())
			Expect(actualVerificationErrors.HasDifference(diff.Size, filepath.Join(expectedRoot, "file-with-different-size"))).To(BeTrue())
		})
		It("will append error to VerificationErrors if contents are different", func() {
			verificationErrors := verify.Errors{}
			nestedError := errors.New("accTime error")
			verificationErrors.Add(diff.AccTime, "some path", nestedError)

			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return &verificationErrors
				}
				modifyThis.IoutilReadFile = func(filename string) ([]byte, error) {
					return def.ProvidePseudoRandomBytes(20, 99199), noError
				}
			})

			regular := def.Regular{}
			regular.Path = "file-with-different-contents"
			regular.Size = 20
			regular.Seed = 99999
			actualError := regular.Verify(expectedRoot)
			Expect(actualError).Should(HaveOccurred())
			Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
			actualVerificationErrors := actualError.(*verify.Errors)
			Expect(actualVerificationErrors.Errors).To(HaveLen(2))
			Expect(actualVerificationErrors.CombinedFileDifference).To(Equal(diff.AccTime | diff.Contents))
			Expect(actualVerificationErrors.HasDifference(diff.AccTime, "some path")).To(BeTrue())
			Expect(actualVerificationErrors.HasDifference(diff.Contents, filepath.Join(expectedRoot, "file-with-different-contents"))).To(BeTrue())
		})
		It("will return ioutil.ReadFile error immediately", func() {
			expectedError := errors.New("ioutil.ReadFile error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.IoutilReadFile = func(filename string) ([]byte, error) {
					return nil, expectedError
				}
			})

			regular := def.Regular{}
			actualError := regular.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
		})
		It("will return nil error in case of success (rather than VerificationErrors with empty list)", func() {
			regular := def.Regular{}
			regular.Size = 20
			regular.Seed = 18
			actualError := regular.Verify("does not matter")
			Expect(actualError).ShouldNot(HaveOccurred())
		})

	})

	Describe("Regular.Create", func() {
		It("will invoke ioutil.WriteFile with path joined with root, number of bytes and mode matching this file", func() {
			const (
				expectedSize         = 2736
				expectedMode         = os.FileMode(123)
				expectedRoot         = "/an/example/root/path"
				expectedRelativePath = "relative/path"
			)

			regular := def.Regular{}
			regular.Size = expectedSize
			regular.Mode = expectedMode
			regular.Path = expectedRelativePath

			actualPath := ""
			actualDataLength := 0
			actualMode := os.FileMode(0)
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					return noError
				}
				modifyThis.IoutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
					actualPath = filename
					actualDataLength = len(data)
					actualMode = perm
					return anError
				}
			})

			regular.Create(expectedRoot)

			Expect(actualPath).To(Equal(filepath.Join(expectedRoot, expectedRelativePath)))
			Expect(actualDataLength).To(Equal(expectedSize))
			Expect(actualMode).To(Equal(expectedMode))
		})
		It("will return os.MkdirAll error", func() {
			regular := def.Regular{}
			expectedError := errors.New("os.MkdirAll error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					return expectedError
				}
			})

			actualError := regular.Create("does not matter")
			Expect(actualError).Should(MatchError(expectedError))
		})
		It("will return ioutil.WriteFile error", func() {
			regular := def.Regular{}
			expectedError := errors.New("ioutil.WriteFile error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.IoutilWriteFile = func(filename string, data []byte, perm os.FileMode) error {
					return expectedError
				}
			})

			actualError := regular.Create("does not matter")
			Expect(actualError).Should(MatchError(expectedError))
		})
	})

	It("will return human readable string representation of this Regular file upon calling Regular.String", func() {
		regular := def.Regular{}
		now, err := time.Parse(file.TimeLayout, "2017-08-02 18:19:52.366534314")
		Expect(err).ShouldNot(HaveOccurred())
		regular.Path = "expected/path"
		regular.Mode = 0664
		regular.Uid = 910
		regular.Gid = 232
		regular.Modified = now
		regular.Accessed = now.Add(time.Hour)
		regular.Size = 2367
		actualString := regular.String()
		Expect(actualString).To(Equal("-rw-rw-r-- 910 232 2017-08-02 18:19:52.366534314 2017-08-02 19:19:52.366534314 expected/path 2367"))
	})

	It("will create regular file using constructor", func() {
		now := time.Now()

		expected := def.Regular{}
		expected.Path = "expected/path"
		expected.Mode = 0765
		expected.Uid = 910
		expected.Gid = 232
		expected.Modified = now
		expected.Accessed = now.Add(time.Hour)
		expected.Size = 4323
		expected.Seed = 39474

		actual := def.Reg("expected/path",
			attr.ModePerm(expected.Mode),
			attr.Uid(expected.Uid),
			attr.Gid(expected.Gid),
			attr.ModifiedTime(expected.Modified),
			attr.AccessedTime(expected.Accessed),
			attr.Size(expected.Size),
			attr.Seed(expected.Seed))(nil, nil)

		Expect(actual.GetPath()).To(Equal(expected.GetPath()))
		Expect(actual.GetUid()).To(Equal(expected.GetUid()))
		Expect(actual.GetGid()).To(Equal(expected.GetGid()))
		Expect(actual.GetMode()).To(Equal(expected.GetMode()))
		Expect(actual.GetModified()).To(BeTemporally("==", expected.GetModified()))
		Expect(actual.GetAccessed()).To(BeTemporally("==", expected.GetAccessed()))
		Expect(actual.(*def.Regular).Size).To(Equal(expected.Size))
		Expect(actual.(*def.Regular).Seed).To(Equal(expected.Seed))
	})
})
