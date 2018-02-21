package user_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	osuser "os/user"
	"github.com/outo/filefactory/dependencies/user"
)

var _ = Describe("pkg user unit test", func() {

	Describe("CurrentIds", func() {

		mockUserCurrent := func(uid, gid string, err error) func() (*osuser.User, error) {
			return func() (*osuser.User, error) {
				return &osuser.User{
					Uid: uid,
					Gid: gid,
				}, err
			}
		}

		It("will return error if user.Current fails", func() {
			expectedError := errors.New("user.Current error")
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = func() (*osuser.User, error) {
					return nil, expectedError
				}
			})
			_, _, _, err := user.CurrentIds()
			Expect(err).To(Equal(expectedError))
		})

		It("will return error if uid is non-numeric", func() {
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("non-numeric", "", nil)
			})

			_, _, _, err := user.CurrentIds()
			Expect(err).Should(HaveOccurred())
		})

		It("will return error if gid is non-numeric", func() {
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("4", "non-numeric", nil)
			})

			_, _, _, err := user.CurrentIds()
			Expect(err).Should(HaveOccurred())
		})

		It("will return error if user groups cannot be retrieved", func() {
			expectedError := errors.New("User.GroupsIds error")
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("4", "4", nil)
				modifyThis.UserGroupIds = func(u *osuser.User) ([]string, error) {
					return nil, expectedError
				}
				_, _, _, err := user.CurrentIds()
				Expect(err).To(Equal(expectedError))
			})

			_, _, _, err := user.CurrentIds()
			Expect(err).Should(HaveOccurred())
		})
		It("will return primary gid as other gid, if there isn't other gid", func() {
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("3424", "4", nil)
				modifyThis.UserGroupIds = func(u *osuser.User) ([]string, error) {
					return []string{"4"}, nil
				}
			})

			uid, primaryGid, otherGid, err := user.CurrentIds()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(uid).To(Equal(uint32(3424)))
			Expect(primaryGid).To(Equal(uint32(4)))
			Expect(otherGid).To(Equal(primaryGid))
		})
		It("will return gid other than primary as other, if available", func() {
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("3424", "4", nil)
				modifyThis.UserGroupIds = func(u *osuser.User) ([]string, error) {
					return []string{"4", "5"}, nil
				}
			})

			uid, primaryGid, otherGid, err := user.CurrentIds()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(uid).To(Equal(uint32(3424)))
			Expect(primaryGid).To(Equal(uint32(4)))
			Expect(otherGid).To(Equal(uint32(5)))
		})
		It("will skip gid that is not numeric", func() {
			user.MockForTest(func(modifyThis *user.Implementation) {
				modifyThis.UserCurrent = mockUserCurrent("3424", "4", nil)
				modifyThis.UserGroupIds = func(u *osuser.User) ([]string, error) {
					return []string{"4", "non-numeric", "6"}, nil
				}
			})

			uid, primaryGid, otherGid, err := user.CurrentIds()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(uid).To(Equal(uint32(3424)))
			Expect(primaryGid).To(Equal(uint32(4)))
			Expect(otherGid).To(Equal(uint32(6)))
		})
	})
})
