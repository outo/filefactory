package verify_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/outo/filefactory/verify"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("fs/copy.go unit test", func() {

	Specify("construct new Instruction (to extend the functionality) with NewInstruction", func() {

		expectedVerify := true
		expectedAspect := "brand-new"

		actualInstruction := verify.NewInstruction(expectedVerify, expectedAspect)
		Expect(actualInstruction.Verify).To(Equal(expectedVerify))
		Expect(actualInstruction.Aspect).To(Equal(expectedAspect))

		expectedVerify = false
		expectedAspect = "something-else"

		actualInstruction = verify.NewInstruction(expectedVerify, expectedAspect)
		Expect(actualInstruction.Verify).To(Equal(expectedVerify))
		Expect(actualInstruction.Aspect).To(Equal(expectedAspect))
	})

	DescribeTable("predefined verification instructions are simple bool and string holders", func(constructor func(bool) verify.Instruction, expectedVerify bool, expectedAspect string) {
		actualInstruction := constructor(expectedVerify)
		Expect(actualInstruction.Verify).To(Equal(expectedVerify))
		Expect(actualInstruction.Aspect).To(Equal(expectedAspect))
	},
		Entry("AllByDefault", verify.AllByDefault, true, "all"),
		Entry("ModePerm", verify.ModePerm, true, "mode-perm"),
		Entry("ModifiedTime", verify.ModifiedTime, true, "modified"),
		Entry("AccessedTime", verify.AccessedTime, true, "accessed"),
		Entry("Uid", verify.Uid, true, "uid"),
		Entry("Gid", verify.Gid, true, "gid"),
		Entry("Size", verify.Size, true, "size"),
		Entry("SymlinkTarget", verify.SymlinkTarget, true, "symlink-target"),
		Entry("Contents", verify.Contents, true, "contents"),
		Entry("AllByDefault", verify.AllByDefault, false, "all"),
		Entry("ModePerm", verify.ModePerm, false, "mode-perm"),
		Entry("ModifiedTime", verify.ModifiedTime, false, "modified"),
		Entry("AccessedTime", verify.AccessedTime, false, "accessed"),
		Entry("Uid", verify.Uid, false, "uid"),
		Entry("Gid", verify.Gid, false, "gid"),
		Entry("Size", verify.Size, false, "size"),
		Entry("SymlinkTarget", verify.SymlinkTarget, false, "symlink-target"),
		Entry("Contents", verify.Contents, false, "contents"),
	)
})
