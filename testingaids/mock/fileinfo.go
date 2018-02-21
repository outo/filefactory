package mock

import (
	"os"
	"time"
)

type FileInfo struct {
	os.FileInfo
	NameFunc    func() string
	SizeFunc    func() int64
	ModeFunc    func() os.FileMode
	ModTimeFunc func() time.Time
	IsDirFunc   func() bool
	SysFunc     func() interface{}
}

func (fi FileInfo) Name() string       { return fi.NameFunc() }
func (fi FileInfo) Size() int64        { return fi.SizeFunc() }
func (fi FileInfo) Mode() os.FileMode  { return fi.ModeFunc() }
func (fi FileInfo) ModTime() time.Time { return fi.ModTimeFunc() }
func (fi FileInfo) IsDir() bool        { return fi.IsDirFunc() }
func (fi FileInfo) Sys() interface{}   { return fi.SysFunc() }

func NewFileInfo() *FileInfo {
	return &FileInfo{}
}

func (fi *FileInfo) WithMode(mode os.FileMode) *FileInfo {
	fi.IsDirFunc = func() bool {
		return mode&os.ModeDir != 0
	}
	fi.ModeFunc = func() os.FileMode {
		return mode
	}
	return fi
}

func (fi *FileInfo) WithSize(size int64) *FileInfo {
	fi.SizeFunc = func() int64 {
		return size
	}
	return fi
}
