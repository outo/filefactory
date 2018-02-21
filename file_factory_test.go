package filefactory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"errors"
	"github.com/outo/filefactory"
	"path/filepath"
	"fmt"
	"os"
	"time"
	"github.com/outo/filefactory/testingaids/mock"
	"github.com/outo/filefactory/file"
	"github.com/outo/filefactory/attr"
	"github.com/outo/filefactory/def"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/diff"
)

var _ = Describe("pkg ff filefactory.go unit test", func() {

	var (
		//noError,
		anError error
	)

	BeforeEach(func() {
		anError = errors.New("just an error, not significant what it is")
	})

	It("will invoke user.CurrentIds() in init, to retrieve current user's id, primary group id and this user's other group id", func() {
		//I won't test if the routine has been invoked as it is in the init function of production part (invoked before I even get to this test case)
		//I can check that the ids have changed from zero values though.
		//To do so I will use predefined attribute types which are translated to arbitrary or default values
		actual := file.Meta{}
		Expect(actual.Uid).To(BeZero())
		Expect(actual.Gid).To(BeZero())
		actual.Populate("does not matter", attr.CurrentUid(), attr.PrimaryGid())
		Expect(actual.Uid).ToNot(BeZero())
		Expect(actual.Gid).ToNot(BeZero())

		//zero both ids
		actual.Populate("does not matter", attr.Uid(0), attr.Gid(0))
		Expect(actual.Gid).To(BeZero())
		actual.Populate("does not matter", attr.OtherGid())
		Expect(actual.Gid).ToNot(BeZero())

		//commonly otherGid will be different than primaryGid, however having user in only one group will mean these gids will be same
	})

	Describe("constructing FileFactory by invoking New function", func() {
		It("will create new FileFactory with default Uid and Gid (primary) set", func() {
			//tested above
		})
		It("will create new FileFactory with default Modified and Accessed times set", func() {
			fac1 := filefactory.New()
			files1 := fac1.FilesToCreate(def.Reg("does not matter"))
			Expect(files1).ToNot(BeEmpty())
			regular1 := files1[0].(*def.Regular)
			actualAccessedTime1 := regular1.Accessed
			actualModifiedTime1 := regular1.Modified

			//another FileFactory instance to demonstrate the principle of a fixed timestamp within an instance
			fac2 := filefactory.New()
			files2 := fac2.FilesToCreate(def.Reg("does not matter"))
			Expect(files2).ToNot(BeEmpty())
			regular2 := files2[0].(*def.Regular)
			actualAccessedTime2 := regular2.Accessed
			actualModifiedTime2 := regular2.Modified
			Expect(actualAccessedTime1).ToNot(BeTemporally("==", actualAccessedTime2))
			Expect(actualModifiedTime1).ToNot(BeTemporally("==", actualModifiedTime2))

			//now using instance 1 again to demonstrate the timestamps haven't changed if ff1 is used
			files1 = fac1.FilesToCreate(def.Reg("does not matter"))
			Expect(files1).ToNot(BeEmpty())
			regular1 = files1[0].(*def.Regular)
			Expect(actualAccessedTime1).To(BeTemporally("==", regular1.Accessed))
			Expect(actualModifiedTime1).To(BeTemporally("==", regular1.Modified))
		})
		It("will use the variadic parameters to overwrite common defaults for all files", func() {
			now := time.Now()
			expectedMode := os.FileMode(0765)
			expectedUid := uint32(3453)
			expectedGid := uint32(4334)
			expectedAccessed := now
			expectedModified := now.Add(-time.Hour)
			fac := filefactory.New(
				expectedMode,
				attr.Uid(expectedUid),
				attr.Gid(expectedGid),
				attr.AccessedTime(expectedAccessed),
				attr.ModifiedTime(expectedModified),
			)

			files := fac.FilesToCreate(
				def.Reg("a/path/to/regular"),
				def.Dir("a/path/to/dir"),
				def.Sym("a/path/to/symlink", "symlink/target"),
			)

			Expect(files).To(HaveLen(3))
			//ConsistOf matcher makes it harder to find out what does not match
			Expect(files).To(ContainElement(
				&def.Regular{
					Meta: file.Meta{
						Path:     "a/path/to/regular",
						Mode:     expectedMode,
						Uid:      expectedUid,
						Gid:      expectedGid,
						Accessed: expectedAccessed,
						Modified: expectedModified,
					},
					Size: 20, //default set within Reg()
				},
			))
			Expect(files).To(ContainElement(
				&def.Directory{
					Meta: file.Meta{
						Path:     "a/path/to/dir",
						Mode:     expectedMode | os.ModeDir,
						Uid:      expectedUid,
						Gid:      expectedGid,
						Accessed: expectedAccessed,
						Modified: expectedModified,
					},
				},
			))
			Expect(files).To(ContainElement(
				&def.Symlink{
					Meta: file.Meta{
						Path:     "a/path/to/symlink",
						Mode:     expectedMode | os.ModeSymlink,
						Uid:      expectedUid,
						Gid:      expectedGid,
						Accessed: expectedAccessed,
						Modified: expectedModified,
						VerificationInstructions: []verify.Instruction{ //default set within Sym()
							verify.ModePerm(false),
							verify.ModifiedTime(false),
							verify.AccessedTime(false),
						},
					},
					LinkTarget: "symlink/target",
				},
			))
		})
	})

	Describe("FilesToCreate constructs File definitions (takes constructor functions) ", func() {
		It("will invoke constructor", func() {
			fac := filefactory.New()
			invoked := false
			var wrapperConstructorFunc filefactory.DefinitionConstructor = func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File {
				invoked = true
				return &struct {
					file.File
				}{}
			}
			fac.FilesToCreate(wrapperConstructorFunc)
			Expect(invoked).To(BeTrue())
		})
		It("is not considered normal for constructor to return nil so that will panic", func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					Expect(recovered.(string)).To(Equal(filefactory.MsgConstructorDidNotDoItsJob))
				}
			}()
			fac := filefactory.New()
			var wrapperConstructorFunc filefactory.DefinitionConstructor = func(hardcodedFileFactoryDefaults []interface{}, extraFileFactoryDefaults []interface{}) file.File {
				return nil
			}

			fac.FilesToCreate(wrapperConstructorFunc)
			Fail("shouldn't get to this line due to panic within the above call")
		})
	})

	Describe("CreateFiles delegates creation of files and alignment of their attributes as per provided definitions", func() {
		const (
			createInvocationWithRoot    = "/tmp/root-323232"
			createInvocationFailure     = "/tmp/root-create-fail"
			alignAttrInvocationWithRoot = "/tmp/root-28643"
			alignAttributesFailure      = "/tmp/root-align-fail"
		)
		type alignAttributesInvocation struct {
			owner,
			mode,
			times bool
			optionalRoot []string
		}
		var (
			createFilesInvocations           []string
			alignAttributesInvocations       []alignAttributesInvocation
			files                            []file.File
			expectedErrorFromCreate          = errors.New("create failed")
			expectedErrorFromAlignAttributes = errors.New("apply attributes failed")
		)
		BeforeEach(func() {
			files = []file.File{}
			alignAttributesInvocations = []alignAttributesInvocation{}
			createFilesInvocations = []string{}
			//mock File.Create and File.AlignAttributes so that the outcome depends on root parameter value and the File instance count
			for i := 0; i < 3; i++ {
				file := mock.NewFile()
				file.CreateFunc = func(root string) error {
					createFilesInvocations = append(createFilesInvocations, root)
					if root == createInvocationFailure {
						return expectedErrorFromCreate
					}
					return nil
				}
				file.AlignAttributesFunc = func(ownershipInAnyCase, modeIfApplicable, timesIfApplicable bool, optionalRoot ...string) error {
					alignAttributesInvocations = append(alignAttributesInvocations, alignAttributesInvocation{
						owner:        ownershipInAnyCase,
						mode:         modeIfApplicable,
						times:        timesIfApplicable,
						optionalRoot: optionalRoot,
					})
					if filepath.Join(optionalRoot...) == alignAttributesFailure {
						return expectedErrorFromAlignAttributes
					}
					return nil
				}
				files = append(files, file)
			}
		})
		It("will invoke File.Create on each argument", func() {
			actualError := filefactory.CreateFiles(createInvocationWithRoot, files...)
			Expect(actualError).ShouldNot(HaveOccurred())
			Expect(createFilesInvocations).To(ConsistOf(
				createInvocationWithRoot,
				createInvocationWithRoot,
				createInvocationWithRoot,
			))
		})
		It("will return immediately if File.Create fails", func() {
			actualError := filefactory.CreateFiles(createInvocationFailure, files...)
			Expect(actualError).Should(MatchError(expectedErrorFromCreate))
			Expect(createFilesInvocations).To(ConsistOf(
				createInvocationFailure,
			))
		})
		It("will invoke File.AlignAttributes on each argument", func() {
			actualError := filefactory.CreateFiles(alignAttrInvocationWithRoot, files...)
			Expect(actualError).ShouldNot(HaveOccurred())
			Expect(alignAttributesInvocations).To(ConsistOf(
				alignAttributesInvocation{owner: true, mode: true, times: true, optionalRoot: []string{alignAttrInvocationWithRoot}},
				alignAttributesInvocation{owner: true, mode: true, times: true, optionalRoot: []string{alignAttrInvocationWithRoot}},
				alignAttributesInvocation{owner: true, mode: true, times: true, optionalRoot: []string{alignAttrInvocationWithRoot}},
			))
		})
		It("will return immediately if File.AlignAttributes fails", func() {
			actualError := filefactory.CreateFiles(alignAttributesFailure, files...)
			Expect(actualError).Should(MatchError(expectedErrorFromAlignAttributes))
			Expect(alignAttributesInvocations).To(ConsistOf(
				alignAttributesInvocation{owner: true, mode: true, times: true, optionalRoot: []string{alignAttributesFailure}},
			))
		})
	})

	Describe("VerifyFiles delegates verification that each definition has a file (regular, directory or symlink) with correct attributes. It also combines human readable results' report", func() {
		const (
			verifyInvocationSuccess                = "/tmp/root-323232"
			verifyInvocationVerificationFailure    = "/tmp/root-verification-failure"
			verifyInvocationNonVerificationFailure = "/tmp/root-non-verification-failure"
		)
		var (
			verifyInvocations            []string
			files                        []file.File
			expectedNonVerificationError = errors.New("verify failed")
			actualError                  error
		)
		BeforeEach(func() {
			files = []file.File{}
			verifyInvocations = []string{}
			//mock File.Verify so that the outcome depends on root parameter value and the File instance count
			for i := 0; i < 3; i++ {
				file := mock.NewFile()
				func(index int) {
					file.VerifyFunc = func(root string) error {
						verifyInvocations = append(verifyInvocations, root)
						if root == verifyInvocationVerificationFailure {
							verificationError := verify.Errors{}
							verificationError.Add(diff.FileDifference(1<<uint(i)), "some path", errors.New(fmt.Sprintf("error %d", index)))
							return &verificationError
						} else if root == verifyInvocationNonVerificationFailure {
							return expectedNonVerificationError
						}
						return nil
					}
					file.StringFunc = func() string {
						return fmt.Sprint("file ", index)
					}
				}(i)
				files = append(files, file)
			}
		})
		Context("given there was no failure during execution", func() {
			BeforeEach(func() {
				actualError = filefactory.VerifyFiles(verifyInvocationSuccess, files...)
			})
			It("will return no error (will be nil, as opposed to verification error instance with empty error list)", func() {
				Expect(actualError).ShouldNot(HaveOccurred())
			})
			It("will delegate to File.Verify for each of the arguments", func() {
				Expect(verifyInvocations).To(ConsistOf(
					verifyInvocationSuccess,
					verifyInvocationSuccess,
					verifyInvocationSuccess,
				))
			})
		})

		Context("given the invocation of File.Verify fails with non-verification error", func() {
			BeforeEach(func() {
				actualError = filefactory.VerifyFiles(verifyInvocationNonVerificationFailure, files...)
			})
			It("will interrupt verification with that error", func() {
				Expect(actualError).To(MatchError(expectedNonVerificationError))
			})
		})

		Context("given the invocation of File.Verify fails with verification error", func() {
			BeforeEach(func() {
				actualError = filefactory.VerifyFiles(verifyInvocationVerificationFailure, files...)
			})
			It("will return error of type Errors when done", func() {
				Expect(actualError).Should(HaveOccurred())
				Expect(actualError).Should(BeAssignableToTypeOf(&verify.Errors{}))
			})
			It("will continue verification and aggregate all verification errors", func() {
				verr := actualError.(*verify.Errors)
				Expect(verr.Errors).Should(HaveLen(3))
			})
			It("will return Errors where slice of errors contains error description", func() {
				verr := actualError.(*verify.Errors)
				Expect(verr.Errors).Should(ConsistOf(
					verify.Error{FileDifference: 8, Path: "some path", Err: errors.New("error 0")},
					verify.Error{FileDifference: 8, Path: "some path", Err: errors.New("error 1")},
					verify.Error{FileDifference: 8, Path: "some path", Err: errors.New("error 2")},
				))
			})
		})
	})
})
