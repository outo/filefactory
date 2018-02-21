package user

import (
	"strconv"
)

// will return current user's uid, primary gid and an example of this user's other gid (same as primary gid if none available)
func CurrentIds() (uid, primaryGid, otherGid uint32, err error) {

	u, err := impl.UserCurrent()
	if err != nil {
		return
	}

	var t uint64

	t, err = strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return
	}
	uid = uint32(t)

	t, err = strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return
	}
	primaryGid = uint32(t)

	gids, err := impl.UserGroupIds(u)
	if err != nil {
		return
	}

	for _, gid := range gids {
		if gid == u.Gid {
			continue
		}
		t, err = strconv.ParseUint(gid, 10, 32)
		if err != nil {
			continue
		}
		otherGid = uint32(t)
	}

	if otherGid == 0 {
		otherGid = primaryGid
	}

	return
}
