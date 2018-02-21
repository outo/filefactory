//attribute describes meta information about a file
// it is used at the time file is created and verified
package attr

import (
	"time"
	"os"
)

var (
	currentUserUid,
	currentUserPrimaryGid,
	currentUserOtherGid uint32
)

func CurrentUserUid() uint32        { return currentUserUid }
func CurrentUserPrimaryGid() uint32 { return currentUserPrimaryGid }
func CurrentUserOtherGid() uint32   { return currentUserOtherGid }

func RetrieveCurrentUserIds() {
	var err error
	currentUserUid, currentUserPrimaryGid, currentUserOtherGid, err = impl.UserCurrentIds()
	if err != nil {
		panic(err)
	}
}

func init() {
	RetrieveCurrentUserIds()
}

//attribute constructors
func ModePerm(mode os.FileMode) os.FileMode { return mode.Perm() }
func ArbitraryUid(val uint32) Uid           { return Uid(val) }
func ArbitraryGid(val uint32) Gid           { return Gid(val) }
func CurrentUid() Uid                       { return ArbitraryUid(CurrentUserUid()) }
func PrimaryGid() Gid                       { return ArbitraryGid(CurrentUserPrimaryGid()) }
func OtherGid() Gid                         { return ArbitraryGid(CurrentUserOtherGid()) }

//attributes
type AccessedTime time.Time
type ModifiedTime time.Time
type Uid uint32
type Gid uint32
type Size int64
type Seed int64
