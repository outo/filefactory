package path_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/outo/filefactory/dependencies/path"
	"os"
	"errors"
	"github.com/outo/filefactory/testingaids/mock"
	"syscall"
)

var _ = Describe("pkg path file path.go", func() {
	BeforeEach(func() {
		path.ResetImplementation()
	})

	It("will return true and no error if file exists", func() {
		path.MockForTest(func(modifyThis *path.Implementation) {
			modifyThis.OsLstat = func(path string) (info os.FileInfo, err error) {
				return mock.FileInfo{}, nil
			}
		})
		actualExists, actualError := path.Exists("does not matter")
		Expect(actualError).ShouldNot(HaveOccurred())
		Expect(actualExists).Should(BeTrue())
	})

	It("will return false and no error if os.Lstat returns os.ErrNotExist error", func() {
		path.MockForTest(func(modifyThis *path.Implementation) {
			modifyThis.OsLstat = func(path string) (info os.FileInfo, err error) {
				return nil, os.ErrNotExist
			}
		})
		actualExists, actualError := path.Exists("does not matter")
		Expect(actualError).ShouldNot(HaveOccurred())
		Expect(actualExists).Should(BeFalse())
	})

	It("will return false and no error if os.Lstat returns syscall.ENOENT error", func() {
		path.MockForTest(func(modifyThis *path.Implementation) {
			modifyThis.OsLstat = func(path string) (info os.FileInfo, err error) {
				return nil, syscall.ENOENT
			}
		})
		actualExists, actualError := path.Exists("does not matter")
		Expect(actualError).ShouldNot(HaveOccurred())
		Expect(actualExists).Should(BeFalse())
	})

	It("will return error from os.Lstat if the error does not indicate that file does not exists", func() {
		path.MockForTest(func(modifyThis *path.Implementation) {
			modifyThis.OsLstat = func(path string) (info os.FileInfo, err error) {
				return nil, errors.New("os.Lstat error")
			}
		})
		_, actualError := path.Exists("does not matter")
		Expect(actualError).Should(HaveOccurred())
		Expect(actualError).Should(MatchError("os.Lstat error"))
	})
})
