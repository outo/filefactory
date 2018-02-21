package file_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/outo/filefactory/testingaids/mock"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/attr"
)

var _ = Describe("pkg file meta.go unit test", func() {

	var (
		noError,
		anError error
	)

	mockedFileInfo := func(
		expectedPath string,
		expectedMode os.FileMode,
		expectedUid uint32,
		expectedGid uint32,
		expectedAccessedTime time.Time,
		expectedModifiedTime time.Time,
	) os.FileInfo {
		fi := mock.FileInfo{
			SysFunc: func() interface{} {
				return &syscall.Stat_t{
					Uid: expectedUid,
					Gid: expectedGid,
					Atim: syscall.Timespec{
						Sec:  expectedAccessedTime.Truncate(time.Second).Unix(),
						Nsec: int64(expectedAccessedTime.Nanosecond()),
					},
					Mtim: syscall.Timespec{
						Sec:  expectedModifiedTime.Truncate(time.Second).Unix(),
						Nsec: int64(expectedModifiedTime.Nanosecond()),
					},
				}
			},
		}
		fi.WithMode(expectedMode)
		return fi
	}

	BeforeEach(func() {
		anError = errors.New("just an error, not significant what it is")
		file.ResetImplementation()
		file.MockForTest(func(modifyThis *file.Implementation) {
			modifyThis.PathExists = func(path string) (exists bool, err error) {
				return true, noError
			}
		})
	})

	Describe("Meta.NewFromPath", func() {
		It("will call os.Lstat with the given path to retrieve its os.FileInfo", func() {
			actualPath := ""
			file.MockForTest(func(modifyThis *file.Implementation) {
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					actualPath = name
					return nil, anError
				}
			})
			const expectedPath = "path/which/we/need/file/info/from"
			file.NewFromPath(expectedPath)
			Expect(actualPath).To(Equal(expectedPath))
		})
		It("will return os.Lstat error", func() {
			expectedError := errors.New("os.Lstat error")
			file.MockForTest(func(modifyThis *file.Implementation) {
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					return nil, expectedError
				}
			})
			_, actualError := file.NewFromPath("does not matter")
			Expect(actualError).To(MatchError(expectedError))
		})
		It("will use retrieved os.FileInfo to construct Meta with it", func() {
			const (
				expectedPath = "path/which/we/need/file/info/from"
				expectedMode = os.FileMode(232)
				expectedUid  = uint32(1883)
				expectedGid  = uint32(980)
			)
			var (
				expectedAccessedTime = time.Now()
				expectedModifiedTime = time.Now()
			)

			mockedExpectedFileInfo := mockedFileInfo(
				expectedPath,
				expectedMode,
				expectedUid,
				expectedGid,
				expectedAccessedTime,
				expectedModifiedTime,
			)

			actualPath := ""
			file.MockForTest(func(modifyThis *file.Implementation) {
				modifyThis.OsLstat = func(name string) (os.FileInfo, error) {
					actualPath = name
					return mockedExpectedFileInfo, noError
				}
			})
			actualFileMeta, err := file.NewFromPath(expectedPath)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(actualPath).To(Equal(expectedPath))
			Expect(actualFileMeta.Path).To(Equal(expectedPath))
			Expect(actualFileMeta.Mode).To(Equal(expectedMode))
			Expect(actualFileMeta.Uid).To(Equal(expectedUid))
			Expect(actualFileMeta.Gid).To(Equal(expectedGid))
			Expect(actualFileMeta.Accessed).To(BeTemporally("==", expectedAccessedTime))
			Expect(actualFileMeta.Modified).To(BeTemporally("==", expectedModifiedTime))
		})
	})
	Describe("given Meta", func() {

		originalPath := "original path"
		originalMode := os.ModeDir
		originalUid := uint32(1883)
		originalGid := uint32(980)
		originalAccessed := time.Now()
		originalModified := time.Now()

		var m file.Meta

		BeforeEach(func() {
			m = file.Meta{
				Path:     originalPath,
				Mode:     originalMode,
				Uid:      originalUid,
				Gid:      originalGid,
				Accessed: originalAccessed,
				Modified: originalModified,
			}
		})

		Describe("Meta.AlignAttributes", func() {
			Describe("path handling", func() {
				It("will report an error when root path is specified but Meta.Path is absolute", func() {
					m.Path = "/when/Meta/path/is/absolute..."
					actualErr := m.AlignAttributes(true, true, true, "...anything/here/will/upset/it")
					Expect(actualErr).Should(MatchError(file.ErrorMessageRootCannotBeUsedWithAbsoluteMetaPath))
				})
				It("will report an error when root path isn't specified but Meta.Path is relative", func() {
					m.Path = "a/relative/path"
					actualErr := m.AlignAttributes(true, true, true)
					Expect(actualErr).Should(MatchError(file.ErrorMessageRelativePathHasToBeUsedWithRoot))
				})
				It("will use absolute Meta.Path as is (and root parameter wasn't provided)", func() {
					const expectedPath = "/an/absolute/path"
					m.Path = expectedPath
					actualPath := ""
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsLchown = func(name string, uid int, gid int) error {
							actualPath = name
							return anError
						}
					})
					m.AlignAttributes(true, true, true)
					Expect(actualPath).To(Equal(expectedPath))
				})
				It("will join relative Meta.Path to root, when root parameter was provided)", func() {
					const expectedPath = "a/relative/path"
					m.Path = expectedPath
					actualPath := ""
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsLchown = func(name string, uid int, gid int) error {
							actualPath = name
							return anError
						}
					})
					specifiedRootPath := "/a/root/path/passed/in"
					m.AlignAttributes(true, true, true, specifiedRootPath)
					Expect(actualPath).To(Equal(filepath.Join(specifiedRootPath, expectedPath)))
				})
			})

			It("will not invoke os.Lchown/os.Chmod/os.Chtimes if none of the attributes is to be aligned, tested with non-symlink (here regular file)", func() {
				var (
					actualOsLchownPath,
					actualOsChmodPath,
					actualOsChtimesPath *string
				)
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.OsLchown = mock.OsLchown(noError, &actualOsLchownPath, nil, nil)
					modifyThis.OsChmod = mock.OsChmod(noError, &actualOsChmodPath, nil)
					modifyThis.OsChtimes = mock.OsChtimes(noError, &actualOsChtimesPath, nil, nil)
				})
				m.Mode = 0666
				m.AlignAttributes(false, false, false, "anything just ot get past path routine")
				Expect(actualOsLchownPath).To(BeNil())
				Expect(actualOsChmodPath).To(BeNil())
				Expect(actualOsChtimesPath).To(BeNil())
			})

			Describe("file ownership alignment", func() {
				It("will invoke os.Lchown with correct parameters if ownership is to be aligned", func() {
					var (
						actualPath *string
						actualUid  *int
						actualGid  *int
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsLchown = mock.OsLchown(anError, &actualPath, &actualUid, &actualGid)
					})

					const (
						expectedPath = "/expectedPath"
						expectedUid  = 123
						expectedGid  = 1621
					)

					m.Path = expectedPath
					m.Uid = expectedUid
					m.Gid = expectedGid
					m.AlignAttributes(true, true, true)
					Expect(*actualPath).To(Equal(expectedPath))
					Expect(*actualUid).To(Equal(expectedUid))
					Expect(*actualGid).To(Equal(expectedGid))
				})
				It("will return immediately with os.Lchown error", func() {
					expectedError := errors.New("os.Lchown error")
					var (
						actualOsChmodPath,
						actualOsChtimesPath *string
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsLchown = mock.OsLchown(expectedError, nil, nil, nil)
						modifyThis.OsChmod = mock.OsChmod(anError, &actualOsChmodPath, nil)
						modifyThis.OsChtimes = mock.OsChtimes(anError, &actualOsChtimesPath, nil, nil)
					})

					actualError := m.AlignAttributes(true, true, true, "just to jump through the path routine")
					Expect(actualError).Should(MatchError(expectedError))
					Expect(actualOsChmodPath).To(BeNil())
					Expect(actualOsChtimesPath).To(BeNil())
				})
			})

			Describe("file mode alignment", func() {
				It("will invoke os.Chmod with correct parameters if it is not a symlink and mode is to be aligned", func() {
					var (
						actualPath *string
						actualMode *os.FileMode
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsChmod = mock.OsChmod(anError, &actualPath, &actualMode)
					})

					const (
						expectedPath = "/expectedPath"
						expectedMode = 0751
					)

					m.Path = expectedPath
					m.Mode = expectedMode
					m.AlignAttributes(false, true, false)
					Expect(*actualPath).To(Equal(expectedPath))
					Expect(*actualMode).To(Equal(os.FileMode(expectedMode)))
				})
				It("will not invoke os.Chmod if it is a symlink even when mode is to be aligned", func() {
					var (
						actualPath *string
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsChmod = mock.OsChmod(anError, &actualPath, nil)
					})

					m.Mode = 0751 | os.ModeSymlink
					m.AlignAttributes(false, true, false)
					Expect(actualPath).To(BeNil())
				})
			})

			Describe("file times alignment", func() {
				It("will invoke os.Chtimes with correct parameters if it is not a symlink and times are to be aligned", func() {
					var (
						actualPath    *string
						actualAccTime,
						actualModTime *time.Time
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsChtimes = mock.OsChtimes(anError, &actualPath, &actualAccTime, &actualModTime)
					})

					const expectedPath = "/expectedPath"
					var (
						expectedAccessed = time.Now()
						expectedModified = time.Now()
					)

					m.Path = expectedPath
					m.Accessed = expectedAccessed
					m.Modified = expectedModified
					m.AlignAttributes(false, false, true)
					Expect(*actualPath).To(Equal(expectedPath))
					Expect(*actualAccTime).To(BeTemporally("==", expectedAccessed))
					Expect(*actualModTime).To(BeTemporally("==", expectedModified))
				})
				It("will not invoke os.Chtimes if it is a symlink even when times are to be aligned", func() {
					var (
						actualPath *string
					)
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.OsChtimes = mock.OsChtimes(anError, &actualPath, nil, nil)
					})

					m.Mode = os.ModeSymlink
					m.AlignAttributes(false, false, true)
					Expect(actualPath).To(BeNil())
				})
			})
		})
		Describe("Verify", func() {

			It("will leave immediately without error if none of the attributes are to be verified", func() {
				m.VerificationInstructions = append(m.VerificationInstructions, verify.AllByDefault(false))
				invoked := false
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						invoked = true
						return file.Meta{}, anError
					}
				})
				actualError := m.Verify("does not matter")
				Expect(actualError).ShouldNot(HaveOccurred())
				Expect(invoked).To(BeFalse())
			})

			It("will add verification error and return immediately if file existence check returns with an error (file most likely inaccessible)", func() {
				invoked := false
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.PathExists = func(path string) (exists bool, err error) {
						err = errors.New("exists() error")
						return
					}
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						invoked = true
						return file.Meta{}, anError
					}
				})

				m.Path = "rel path"
				err := m.Verify("does not matter")
				Expect(invoked).To(BeFalse())
				Expect(err).Should(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
				verErr := err.(*verify.Errors)
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, filepath.Join("does not matter", "rel path")))
			})

			It("will add verification error and return immediately if file existence check returns false without error (file does not exists)", func() {
				invoked := false
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.PathExists = func(path string) (exists bool, err error) {
						return
					}
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						invoked = true
						return file.Meta{}, anError
					}
				})

				m.Path = "rel path"
				err := m.Verify("does not matter")
				Expect(invoked).To(BeFalse())
				Expect(err).Should(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
				verErr := err.(*verify.Errors)
				Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, filepath.Join("does not matter", "rel path")))
			})

			It("will invoke base.NewFromPath with absolute path to create a wrapper for file's attributes (Meta)", func() {
				const expectedRelativePath = "relative/path"
				m.Path = expectedRelativePath
				actualPath := ""
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						actualPath = path
						return file.Meta{}, anError
					}
				})
				const expectedRoot = "/an/example/root/path"
				m.Verify(expectedRoot)
				Expect(actualPath).To(Equal(filepath.Join(expectedRoot, expectedRelativePath)))
			})

			It("will return base.NewFromPath error", func() {
				expectedError := errors.New("base.NewFromPath error")
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						return file.Meta{}, expectedError
					}
				})
				actualError := m.Verify("does not matter")
				Expect(actualError).Should(MatchError(expectedError))
			})

			It("will collate all errors and return within Errors", func() {
				retrievedFm := file.Meta{
					Mode:     0123,
					Uid:      234,
					Gid:      345,
					Accessed: time.Now(),
					Modified: time.Now(),
				}
				file.MockForTest(func(modifyThis *file.Implementation) {
					modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
						return retrievedFm, noError
					}
				})
				fm := file.Meta{
					Mode:     0765,
					Uid:      654,
					Gid:      543,
					Accessed: time.Now(),
					Modified: time.Now(),
				}
				actualError := fm.Verify("/some/root")
				Expect(actualError).Should(HaveOccurred())
				Expect(actualError).Should(BeAssignableToTypeOf(&verify.Errors{}))
				actualVerificationErrors := actualError.(*verify.Errors)
				Expect(actualVerificationErrors.CombinedFileDifference).To(Equal(
					diff.ModePerm | diff.Owner | diff.Group | diff.AccTime | diff.ModTime,
				))
				Expect(actualVerificationErrors.Errors).To(HaveLen(5))
			})

			Describe("mode verification", func() {

				BeforeEach(func() {
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
							m.Mode = 12
							return m, noError
						}
					})
				})

				It("will interrupt verification if file type is unexpected (e.g. found dir when it should be a regular file)", func() {
					retrievedMeta := file.Meta{
						Mode:0765|os.ModeSymlink,
					}
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
							return retrievedMeta, noError
						}
					})

					m.Mode = 0500|os.ModeDir
					err := m.Verify("does not matter")
					Expect(err).Should(HaveOccurred())
					Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
					verErr := err.(*verify.Errors)
					Expect(verErr.HasDifference(diff.ModeType, filepath.Join("does not matter", originalPath))).To(BeTrue())

					//because the verification will be interrupted
					Expect(verErr.HasDifference(diff.ModePerm, filepath.Join("does not matter", originalPath))).To(BeFalse())
				})

				It("will verify mode permissions if it was requested", func() {
					m.Mode = 176
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.ModePerm))

					m.Mode = 12
					actualError = m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will not error when mode is different but check wasn't requested", func() {
					m.Mode = 54
					m.VerificationInstructions = append(m.VerificationInstructions, verify.ModePerm(false))
					actualError := m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
			})

			Describe("ownership verification", func() {
				BeforeEach(func() {
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
							m.Uid = originalUid
							m.Gid = originalGid
							return m, noError
						}
					})
				})

				It("will verify owner if ownership check was requested", func() {
					m.Uid = 765
					m.VerificationInstructions = append(m.VerificationInstructions,
						verify.AllByDefault(false),
						verify.Uid(true),
					)
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.Owner))

					m.Uid = originalUid
					actualError = m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will verify group if ownership check was requested", func() {
					m.Gid = 900
					m.VerificationInstructions = append(m.VerificationInstructions,
						verify.AllByDefault(false),
						verify.Gid(true),
					)
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.Group))

					m.Gid = originalGid
					actualError = m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will not error when owner is different but check wasn't requested", func() {
					m.Uid = 82
					m.VerificationInstructions = append(m.VerificationInstructions, verify.Uid(false))
					actualError := m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will not error when group is different but check wasn't requested", func() {
					m.Gid = 533
					m.VerificationInstructions = append(m.VerificationInstructions, verify.Gid(false))
					actualError := m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will error if both owner and group are different and ownership check was required", func() {
					m.Uid = 7632
					m.Gid = 882
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.Group | diff.Owner))
				})
			})

			Describe("time verification", func() {

				BeforeEach(func() {
					file.MockForTest(func(modifyThis *file.Implementation) {
						modifyThis.WrapNewFromPath = func(path string) (meta file.Meta, err error) {
							m.Accessed = originalAccessed
							m.Modified = originalModified
							return m, noError
						}
					})
				})

				It("will verify access time if time check was requested", func() {
					m.Accessed = time.Now()
					m.Modified = originalModified
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.AccTime))

					m.Accessed = originalAccessed
					actualError = m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will verify modified time if times check was requested", func() {
					m.Accessed = originalAccessed
					m.Modified = time.Now()
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.ModTime))

					m.Modified = originalModified
					actualError = m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will not error when access time is different but check wasn't requested", func() {
					m.Accessed = time.Now()
					m.Modified = originalModified
					m.VerificationInstructions = append(m.VerificationInstructions, verify.AccessedTime(false))
					actualError := m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will not error when modified time is different but check wasn't requested", func() {
					m.Accessed = originalAccessed
					m.Modified = time.Now()
					m.VerificationInstructions = append(m.VerificationInstructions, verify.ModifiedTime(false))
					actualError := m.Verify("does not matter")
					Expect(actualError).ShouldNot(HaveOccurred())
				})
				It("will error if both access and modified time are different and check was required", func() {
					m.Accessed = time.Now()
					m.Modified = time.Now()
					actualError := m.Verify("does not matter")
					Expect(actualError).Should(HaveOccurred())
					Expect(actualError).To(BeAssignableToTypeOf(&verify.Errors{}))
					vErr := actualError.(*verify.Errors)
					Expect(vErr.CombinedFileDifference).To(Equal(diff.AccTime | diff.ModTime))
				})
			})
		})
		Describe("Populate", func() {
			var m file.Meta
			BeforeEach(func() {
				m = file.Meta{}
			})
			Describe("with attributes", func() {
				It("will populate Path with arbitrary value", func() {
					m.Populate("/expected/path")
					Expect(m.Path).To(Equal("/expected/path"))
				})
				It("will populate ModePerm arbitrary value", func() {
					m.Populate("", attr.ModePerm(0755))
					Expect(m.Mode).To(Equal(os.FileMode(0755)))
				})
				It("will populate Accessed arbitrary value", func() {
					now := time.Now()
					m.Populate("", attr.AccessedTime(now))
					Expect(m.Accessed).To(BeTemporally("==", now))
				})
				It("will populate Modified arbitrary value", func() {
					now := time.Now()
					m.Populate("", attr.ModifiedTime(now))
					Expect(m.Modified).To(BeTemporally("==", now))
				})
				It("will populate Uid arbitrary value", func() {
					m.Populate("", attr.Uid(501))
					Expect(m.Uid).To(Equal(uint32(501)))
				})
				It("will populate Gid arbitrary value", func() {
					m.Populate("", attr.Gid(1001))
					Expect(m.Gid).To(Equal(uint32(1001)))
				})
			})
			Describe("with verification instructions", func() {
				isLast := func(m file.Meta, sought verify.Instruction) (valueOfVerify bool) {
					if len(m.VerificationInstructions) == 0 {
						return
					}
					last := m.VerificationInstructions[len(m.VerificationInstructions) - 1]
					if last.Aspect == sought.Aspect {
						valueOfVerify = last.Verify
					}
					return
				}

				It("will append ModePerm verification instruction", func() {
					Expect(isLast(m, verify.ModePerm(true))).To(BeFalse())
					m.Populate("", verify.ModePerm(true))
					Expect(isLast(m, verify.ModePerm(true))).To(BeTrue())
				})
				It("will append any verification instruction", func() {
					newInstruction := verify.NewInstruction(true, "not even known yet")

					Expect(isLast(m, newInstruction)).To(BeFalse())
					m.Populate("", newInstruction)
					Expect(isLast(m, newInstruction)).To(BeTrue())
				})
			})
		})
		It("will return human readable string representation of this File upon calling File.String", func() {
			now, err := time.Parse(file.TimeLayout, "2017-08-02 18:19:52.366534314")
			Expect(err).ShouldNot(HaveOccurred())
			m.Modified = now
			m.Accessed = now.Add(time.Hour)
			actualString := m.String()
			Expect(actualString).To(Equal("d--------- 1883 980 2017-08-02 18:19:52.366534314 2017-08-02 19:19:52.366534314 original path"))
		})
		Describe("Getters", func() {
			It("will provide corresponding, unchanged values", func() {
				m.Path = "expected/path"
				m.Mode = 234234
				m.Accessed = time.Now()
				m.Modified = time.Now().Add(time.Hour) //not necessary as there would be ns difference form Accessed
				m.Uid = 4374
				m.Gid = 1324
				Expect(m.GetPath()).To(Equal(m.Path))
				Expect(m.GetMode()).To(Equal(m.Mode))
				Expect(m.GetAccessed()).To(Equal(m.Accessed))
				Expect(m.GetModified()).To(Equal(m.Modified))
				Expect(m.GetUid()).To(Equal(m.Uid))
				Expect(m.GetGid()).To(Equal(m.Gid))
			})
		})
	})

})
