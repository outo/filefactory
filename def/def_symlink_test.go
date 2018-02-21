package def_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"time"
	"github.com/outo/filefactory/def"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
	"github.com/outo/filefactory/attr"
)

var _ = Describe("pkg def file_symlink.go unit test", func() {

	const retrievedSymlinkTarget = "retrieved/target"

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
			modifyThis.OsReadlink = func(name string) (string, error) {
				return retrievedSymlinkTarget, noError
			}
		})
	})

	Describe("Symlink.Verify", func() {
		const expectedRoot = "/an/example/root"

		It("will invoke Meta.Verify with root parameter and indication that only ownership matters in symlinks", func() {
			actualRoot := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					actualRoot = root

					return anError
				}
			})

			symlink := def.Symlink{}

			symlink.Verify(expectedRoot)
			Expect(actualRoot).To(Equal(expectedRoot))
		})

		It("will return Meta.Verify error immediately if of type other than VerificationErrors", func() {
			invoked := false
			expectedError := errors.New("Meta.Verify error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return expectedError
				}
				modifyThis.OsReadlink = func(name string) (string, error) {
					invoked = true
					return "", anError
				}
			})

			symlink := def.Symlink{}
			actualError := symlink.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
			Expect(invoked).To(BeFalse())
		})
		It("will not return Meta.Verify error immediately if of type VerificationErrors (it contains list of verification specific failures). It will continue to os.Readlink", func() {
			invoked := false
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return &verify.Errors{}
				}
				modifyThis.OsReadlink = func(name string) (string, error) {
					invoked = true
					return "", anError
				}
			})

			symlink := def.Symlink{}
			symlink.Verify(expectedRoot)
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
				modifyThis.OsReadlink = func(name string) (string, error) {
					invoked = true
					return "", anError
				}
			})

			symlink := def.Symlink{}
			symlink.Verify(expectedRoot)
			Expect(invoked).To(BeFalse())
		})

		It("will not continue verification if file type is unexpected (e.g. found dir when it should be a symlink file)", func() {
			invoked := false
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					verErr := verify.Errors{}
					verErr.Add(diff.ModeType, "some path", anError)
					return &verErr
				}
				modifyThis.OsReadlink = func(name string) (string, error) {
					invoked = true
					return "", anError
				}
			})

			symlink := def.Symlink{}
			symlink.Verify(expectedRoot)
			Expect(invoked).To(BeFalse())
		})
		It("will invoke os.Readlink with a path joined to root parameter", func() {
			actualPath := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsReadlink = func(name string) (string, error) {
					actualPath = name
					return "", anError
				}
			})

			symlink := def.Symlink{}
			symlink.Path = "expected/relative/path/to/file"

			symlink.Verify(expectedRoot)
			Expect(actualPath).To(Equal(filepath.Join(expectedRoot, symlink.Path)))
		})
		It("will return os.Readlink error immediately", func() {
			expectedError := errors.New("os.Readlink error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsReadlink = func(name string) (string, error) {
					return "", expectedError
				}
			})

			symlink := def.Symlink{}
			actualError := symlink.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
		})
		It("will append error to VerificationErrors if target is different", func() {
			verificationErrors := verify.Errors{}
			nestedError := errors.New("owner error")
			verificationErrors.Add(diff.Owner, "some path", nestedError)

			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					return &verificationErrors
				}
			})

			symlink := def.Symlink{}
			symlink.LinkTarget = retrievedSymlinkTarget + "/make/it/different"
			actualError := symlink.Verify(expectedRoot)
			Expect(actualError).Should(HaveOccurred())
			Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
			actualVerificationErrors := actualError.(*verify.Errors)
			Expect(actualVerificationErrors.Errors).To(HaveLen(2))
			Expect(actualVerificationErrors.CombinedFileDifference).To(Equal(diff.Owner | diff.LinkTarget))
			Expect(actualVerificationErrors.HasDifference(diff.Owner, "some path"))
			Expect(actualVerificationErrors.HasDifference(diff.LinkTarget, "some path"))
		})
		It("will return nil error in case of success (rather than VerificationErrors with empty list)", func() {
			symlink := def.Symlink{}
			symlink.LinkTarget = retrievedSymlinkTarget
			actualError := symlink.Verify("does not matter")
			Expect(actualError).ShouldNot(HaveOccurred())
		})

	})

	Describe("Symlink.Create", func() {
		It("will invoke os.Symlink with path joined with root and with link target matching this file", func() {
			const (
				expectedTarget       = "expected/link/target"
				expectedRoot         = "/an/example/root/path"
				expectedRelativePath = "relative/path"
			)

			symlink := def.Symlink{}
			symlink.LinkTarget = expectedTarget
			symlink.Path = expectedRelativePath

			actualTarget := ""
			actualSymlinkPath := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					return noError
				}
				modifyThis.OsSymlink = func(oldname string, newname string) error {
					actualTarget = oldname
					actualSymlinkPath = newname
					return anError
				}
			})

			symlink.Create(expectedRoot)

			Expect(actualTarget).To(Equal(expectedTarget))
			Expect(actualSymlinkPath).To(Equal(filepath.Join(expectedRoot, expectedRelativePath)))
		})
		It("will return os.MkdirAll error", func() {
			symlink := def.Symlink{}
			expectedError := errors.New("os.MkdirAll error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					return expectedError
				}
			})

			actualError := symlink.Create("does not matter")
			Expect(actualError).Should(MatchError(expectedError))
		})
		It("will return os.Symlink error", func() {
			symlink := def.Symlink{}
			expectedError := errors.New("os.Symlink error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsSymlink = func(oldname string, newname string) error {
					return expectedError
				}
			})

			actualError := symlink.Create("does not matter")
			Expect(actualError).Should(MatchError(expectedError))
		})
	})

	It("will return human readable string representation of this Symlink upon calling Symlink.String", func() {
		symlink := def.Symlink{}
		now, err := time.Parse(file.TimeLayout, "2017-08-02 18:19:52.366534314")
		Expect(err).ShouldNot(HaveOccurred())
		symlink.Path = "expected/path"
		symlink.Mode = os.ModeSymlink
		symlink.Uid = 910
		symlink.Gid = 232
		symlink.Modified = now
		symlink.Accessed = now.Add(time.Hour)
		symlink.LinkTarget = "/target/of/this/link"
		actualString := symlink.String()
		Expect(actualString).To(Equal("L--------- 910 232 2017-08-02 18:19:52.366534314 2017-08-02 19:19:52.366534314 expected/path -> /target/of/this/link"))
	})

	It("will create symlink using constructor", func() {
		now := time.Now()

		expected := def.Symlink{}
		expected.Path = "expected/path"
		expected.Mode = os.ModeSymlink
		expected.Uid = 910
		expected.Gid = 232
		expected.Modified = now
		expected.Accessed = now.Add(time.Hour)
		expected.LinkTarget = "/target/of/this/link"

		actual := def.Sym("expected/path",
			expected.LinkTarget,
			attr.ModePerm(expected.Mode),
			attr.Uid(expected.Uid),
			attr.Gid(expected.Gid),
			attr.ModifiedTime(expected.Modified),
			attr.AccessedTime(expected.Accessed))(nil, nil)

		Expect(actual.GetPath()).To(Equal(expected.GetPath()))
		Expect(actual.GetUid()).To(Equal(expected.GetUid()))
		Expect(actual.GetGid()).To(Equal(expected.GetGid()))
		Expect(actual.GetMode()).To(Equal(expected.GetMode()))
		Expect(actual.GetModified()).To(BeTemporally("==", expected.GetModified()))
		Expect(actual.GetAccessed()).To(BeTemporally("==", expected.GetAccessed()))
		Expect(actual.(*def.Symlink).LinkTarget).To(Equal(expected.LinkTarget))
	})

})
