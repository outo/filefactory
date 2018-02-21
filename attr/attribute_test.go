package attr_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/outo/filefactory/attr"
	"os"
	"time"
	"errors"
)

var _ = Describe("pkg attr file attribute test", func() {

	Specify("what file permission should be. Although this function returns os.FileMode the only portion that will be used is permissions, i.e. os.FileMode&os.ModePerm", func() {
		actual := attr.ModePerm(0750)
		Expect(actual).To(Equal(os.FileMode(0750)))
	})

	It("will only set the ModePerm bits with ModePerm", func() {
		actual := attr.ModePerm(0750|os.ModeDir)
		Expect(actual).To(Equal(os.FileMode(0750)))
	})

	Specify("what user id file's owner will be set to", func() {
		actual := attr.ArbitraryUid(1001)
		Expect(actual).To(Equal(attr.Uid(1001)))
	})

	Specify("what group id file's group will be set to", func() {
		actual := attr.ArbitraryGid(501)
		Expect(actual).To(Equal(attr.Gid(501)))
	})

	Describe("given current user id's. ", func() {
		BeforeEach(func() {
			attr.MockForTest(func(modifyThis *attr.Implementation) {
				modifyThis.UserCurrentIds = func() (uid uint32, primaryGid uint32, otherGid uint32, err error) {
					return 15, 25, 35, nil
				}
			})

			/*
			It will normally be populated on package initialisation.
			In this test it had to be triggered explicitly as - in normal circumstances - the call happens before this test is run,
			so I couldn't mock it.
			 */
			attr.RetrieveCurrentUserIds()
		})

		Specify("that file's owner needs to be set to current user's id. ", func() {
			actual := attr.CurrentUid()
			Expect(actual).To(BeAssignableToTypeOf(attr.Uid(0)))
			Expect(actual).To(BeEquivalentTo(15))
		})

		Specify("that file's group needs to be set to current user's primary group id. ", func() {
			actual := attr.PrimaryGid()
			Expect(actual).To(BeAssignableToTypeOf(attr.Gid(0)))
			Expect(actual).To(BeEquivalentTo(25))
		})

		Specify("that file's group needs to be set to current user's group id, other than primary group id (if available). ", func() {
			actual := attr.OtherGid()
			Expect(actual).To(BeAssignableToTypeOf(attr.Gid(0)))
			Expect(actual).To(BeEquivalentTo(35))
		})
	})

	Describe("given failing user.CurrentIds() invocation", func() {
		BeforeEach(func() {
			attr.MockForTest(func(modifyThis *attr.Implementation) {
				modifyThis.UserCurrentIds = func() (uid uint32, primaryGid uint32, otherGid uint32, err error) {
					err = errors.New("user.CurrentIds() error")
					return
				}
			})
		})

		It("will panic", func() {
			Expect(attr.RetrieveCurrentUserIds).To(Panic())
		})
	})

	Specify("what file's accessed time will be set to", func() {
		now := time.Now()
		actual := attr.AccessedTime(now)
		Expect(actual).To(BeAssignableToTypeOf(attr.AccessedTime(time.Time{})))
		Expect(actual).To(BeEquivalentTo(now))
	})

	Specify("what file's modified time will be set to", func() {
		now := time.Now()
		actual := attr.ModifiedTime(now)
		Expect(actual).To(BeAssignableToTypeOf(attr.ModifiedTime(time.Time{})))
		Expect(actual).To(BeEquivalentTo(now))
	})

	Specify("number of bytes that this file will consist of", func() {
		actual := attr.Size(1709)
		Expect(actual).To(BeAssignableToTypeOf(attr.Size(0)))
		Expect(actual).To(BeEquivalentTo(1709))
	})

	Specify("the seed for the pseudo-random generator that will populate this file's contents", func() {
		actual := attr.Seed(82736)
		Expect(actual).To(BeAssignableToTypeOf(attr.Seed(0)))
		Expect(actual).To(BeEquivalentTo(82736))
	})

	It("will retrieve current user ids on production package initialisation, but I am unable to test this invocation", func() {

	})
})
