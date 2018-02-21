package verify_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"errors"
	"github.com/outo/filefactory/diff"
	"github.com/outo/filefactory/verify"
)

var _ = Describe("pkg ff verification_errors.go unit test", func() {

	var (
		anError,
		anotherError,
		someError,
		yetAnotherError error
		fileDifferenceOf4  = diff.FileDifference(4)
		fileDifferenceOf8  = diff.FileDifference(8)
		fileDifferenceOf16 = diff.FileDifference(16)
		fileDifferenceOf1  = diff.FileDifference(1)
	)

	BeforeEach(func() {
		anError = errors.New("just an error, not significant what it is")
		anotherError = errors.New("just another error, not significant what it is")
		someError = errors.New("some error, not significant what it is")
		yetAnotherError = errors.New("yet another error, not significant what it is")
	})

	It("will add an error and perform bit union to capture the difference", func() {
		verr := verify.Errors{}
		Expect(verr.CombinedFileDifference).To(BeZero())
		Expect(verr.Errors).To(BeEmpty())

		verr.Add(fileDifferenceOf4, "some path", anError)
		Expect(verr.CombinedFileDifference).To(Equal(fileDifferenceOf4))
		Expect(verr.Errors).To(ConsistOf(
			verify.Error{FileDifference: fileDifferenceOf4, Path: "some path", Err: anError},
		))

		verr.Add(fileDifferenceOf8, "some other path", anotherError)
		Expect(verr.CombinedFileDifference).To(Equal(fileDifferenceOf4 | fileDifferenceOf8))
		Expect(verr.Errors).To(ConsistOf(
			verify.Error{FileDifference: fileDifferenceOf4, Path: "some path", Err: anError},
			verify.Error{FileDifference: fileDifferenceOf8, Path: "some other path", Err: anotherError},
		))
	})

	Describe("given sample of 2 verification errors inside verErr", func() {
		var verErr verify.Errors
		BeforeEach(func() {
			verErr = verify.Errors{}
			verErr.Add(fileDifferenceOf8, "some path (8)", anError)
			verErr.Add(fileDifferenceOf4, "some path (4)", anotherError)
		})
		It("will print all error messages", func() {
			Expect(verErr.Error()).To(ContainSubstring(anError.Error()))
			Expect(verErr.Error()).To(ContainSubstring(anotherError.Error()))
		})

		Describe("merging errors", func() {
			It("will merge provided error into this, if of type Errors, then will return nil", func() {
				providedVerErr := &verify.Errors{}
				providedVerErr.Add(fileDifferenceOf16, "some path (16)", someError)
				providedVerErr.Add(fileDifferenceOf1, "some path (1)", yetAnotherError)

				returnedErr := verErr.Merge(providedVerErr)
				Expect(returnedErr).ShouldNot(HaveOccurred())
				Expect(verErr.CombinedFileDifference).To(Equal(
					fileDifferenceOf1 | fileDifferenceOf4 | fileDifferenceOf8 | fileDifferenceOf16,
				))
				Expect(verErr.Errors).To(ConsistOf(
					verify.NewErr(fileDifferenceOf1, "some path (1)", yetAnotherError),
					verify.NewErr(fileDifferenceOf4, "some path (4)", anotherError),
					verify.NewErr(fileDifferenceOf8, "some path (8)", anError),
					verify.NewErr(fileDifferenceOf16, "some path (16)", someError),
				))

			})
			It("will not merge provided error if not of type Errors, then it will return provided error", func() {
				nonVerErr := errors.New("failure that could happen during verification but it is not business logic related")
				actualErr := verErr.Merge(nonVerErr)
				Expect(actualErr).To(MatchError(nonVerErr))
			})
		})
	})

	Describe("mapping to nil (on return usually)", func() {
		It("will return itself if there are errors in this verification errors instance", func() {
			verErr := verify.Errors{}
			verErr.Add(fileDifferenceOf16, "some path", anError)
			actualErr := verErr.MapToNilIfNone()
			Expect(actualErr).Should(Equal(&verErr))
		})

		It("will return nil if there are no errors in this verification errors instance", func() {
			verErr := verify.Errors{}
			actualErr := verErr.MapToNilIfNone()
			Expect(actualErr).Should(BeNil())
		})

	})
	Describe("given differences were detected for paths", func() {
		var verErr verify.Errors
		BeforeEach(func() {
			verErr = verify.Errors{}
			verErr.Add(diff.ModTime, "expected/path", errors.New("some error"))
			verErr.Add(diff.AccTime, "expected/path", errors.New("some error"))
			verErr.Add(diff.Size, "different/path", errors.New("some error"))
		})

		Describe("checking if a specific difference was detected for given path", func() {
			It("will return true if the difference was detected for this path", func() {
				Expect(verErr.HasDifference(diff.ModTime, "expected/path")).To(BeTrue())
			})
			It("will return false if the difference was detected, but not for this path", func() {
				Expect(verErr.HasDifference(diff.Size, "expected/path")).To(BeFalse())
			})
			It("will return false if other difference was detected for this path", func() {
				Expect(verErr.HasDifference(diff.LinkTarget, "expected/path")).To(BeFalse())
			})
		})

		Describe("retrieving all detected differences for a given path", func() {
			It("will return all FileDifference(s) detected for a path", func() {
				Expect(verErr.DifferenceFor("expected/path")).To(Equal(diff.ModTime|diff.AccTime))
			})
			It("will return no file difference (value of 0) if a path has no differences detected", func() {
				Expect(verErr.DifferenceFor("unknown/path")).To(BeZero())
			})
		})
	})

	Describe("checking if verification could not detect file (not present or not accessible)", func() {
		It("will return true if file difference is equal diff.NotPresentOrNotAccessible", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.NotPresentOrNotAccessible}
			Expect(verErr.IsFileNotPresentOrNotAccessible()).To(BeTrue())
		})
		It("will return true if file difference is contains diff.NotPresentOrNotAccessible bitmask", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.NotPresentOrNotAccessible | diff.AccTime}
			Expect(verErr.IsFileNotPresentOrNotAccessible()).To(BeTrue())
		})
		It("will return false if file difference does not contain diff.NotPresentOrNotAccessible bitmask", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.Contents}
			Expect(verErr.IsFileNotPresentOrNotAccessible()).To(BeFalse())
		})
	})

	Describe("checking if verification detected file type was unexpected (e.g. directory instead of a file)", func() {
		It("will return true if file difference is equal diff.ModeType", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.ModeType}
			Expect(verErr.IsFileTypeUnexpected()).To(BeTrue())
		})
		It("will return true if file difference is contains diff.ModeType bitmask", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.ModeType | diff.AccTime}
			Expect(verErr.IsFileTypeUnexpected()).To(BeTrue())
		})
		It("will return false if file difference does not contain diff.ModeType bitmask", func() {
			verErr := verify.Errors{CombinedFileDifference: diff.Size}
			Expect(verErr.IsFileNotPresentOrNotAccessible()).To(BeFalse())
		})
	})

})
