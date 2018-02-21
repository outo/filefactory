package path

import "os"

func Exists(path string) (exists bool, err error) {
	if _, err = impl.OsLstat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return
}