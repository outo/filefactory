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
	"github.com/outo/filefactory/attr"
)

var _ = Describe("pkg def file_directory.go unit test", func() {

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
			modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
				return noError
			}
		})
	})

	Describe("Directory.Verify", func() {
		const expectedRoot = "/an/example/root"

		It("will invoke Meta.Verify with root parameter and indication all three attributes needs verifying", func() {
			actualRoot := ""
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.MetaVerify = func(fileMeta file.Meta, root string) error {
					actualRoot = root
					return anError
				}
			})

			dir := def.Directory{}

			dir.Verify(expectedRoot)
			Expect(actualRoot).To(Equal(expectedRoot))
		})

		It("will return Meta.Verify error immediately", func() {
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

			dir := def.Directory{}
			actualError := dir.Verify(expectedRoot)
			Expect(actualError).Should(MatchError(expectedError))
			Expect(invoked).To(BeFalse())
		})
	})

	Describe("Directory.Create", func() {
		It("will invoke os.MkdirAll with path joined with root and mode matching this file", func() {
			const (
				expectedMode         = os.FileMode(123)
				expectedRoot         = "/an/example/root/path"
				expectedRelativePath = "relative/path"
			)

			dir := def.Directory{}
			dir.Mode = expectedMode
			dir.Path = expectedRelativePath

			var actualPaths []string
			var actualModes []os.FileMode
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					actualPaths = append(actualPaths, path)
					actualModes = append(actualModes, perm)
					return noError
				}
			})

			dir.Create(expectedRoot)

			Expect(actualPaths).To(ConsistOf(
				filepath.Join(expectedRoot, "relative"),
				filepath.Join(expectedRoot, "relative/path"),
			))
			Expect(actualModes).To(ConsistOf(
				os.FileMode(0777),
				expectedMode,
			))
		})
		It("will return os.MkdirAll error", func() {
			dir := def.Directory{}
			expectedError := errors.New("os.MkdirAll error")
			def.MockForTest(func(modifyThis *def.Implementation) {
				modifyThis.OsMkdirAll = func(path string, perm os.FileMode) error {
					return expectedError
				}
			})

			actualError := dir.Create("does not matter")
			Expect(actualError).Should(MatchError(expectedError))
		})
	})

	It("will return human readable string representation of this Dir file upon calling Dir.String", func() {
		directory := def.Directory{}
		now, err := time.Parse(file.TimeLayout, "2017-08-02 18:19:52.366534314")
		Expect(err).ShouldNot(HaveOccurred())
		directory.Path = "expected/path"
		directory.Mode = os.ModeDir
		directory.Uid = 910
		directory.Gid = 232
		directory.Modified = now
		directory.Accessed = now.Add(time.Hour)
		actualString := directory.String()
		Expect(actualString).To(Equal("d--------- 910 232 2017-08-02 18:19:52.366534314 2017-08-02 19:19:52.366534314 expected/path"))
	})

	It("will create directory using constructor", func() {
		now := time.Now()

		expected := def.Directory{}
		expected.Path = "expected/path"
		expected.Mode = 0753 | os.ModeDir
		expected.Uid = 910
		expected.Gid = 232
		expected.Modified = now
		expected.Accessed = now.Add(time.Hour)

		actual := def.Dir("expected/path",
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
	})
})
