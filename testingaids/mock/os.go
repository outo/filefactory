package mock

import (
	"os"
	"time"
)

func OsChmod(returnErr error, recordName **string, recordMode **os.FileMode) func(name string, mode os.FileMode) error {
	return func(name string, mode os.FileMode) error {
		if recordName != nil {
			*recordName = &name
		}
		if recordMode != nil {
			*recordMode = &mode
		}
		return returnErr
	}
}

func OsChtimes(returnErr error, recordName **string, recordAtime, recordMtime **time.Time) func(name string, atime time.Time, mtime time.Time) error {
	return func(name string, atime time.Time, mtime time.Time) error {
		if recordName != nil {
			*recordName = &name
		}
		if recordAtime != nil {
			*recordAtime = &atime
		}
		if recordMtime != nil {
			*recordMtime = &mtime
		}
		return returnErr
	}
}

func OsLchown(returnErr error, recordName **string, recordUid, recordGid **int) func(name string, uid int, gid int) error {
	return func(name string, uid int, gid int) error {
		if recordName != nil {
			*recordName = &name
		}
		if recordUid != nil {
			*recordUid = &uid
		}
		if recordGid != nil {
			*recordGid = &gid
		}
		return returnErr
	}
}

func OsLstat(returnInfo os.FileInfo, returnErr error) func(name string) (os.FileInfo, error) {
	return func(name string) (os.FileInfo, error) {
		return returnInfo, returnErr
	}

}
