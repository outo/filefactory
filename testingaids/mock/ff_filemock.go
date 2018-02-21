package mock

import (
	"time"
	"os"
)

type File struct {
	CreateFunc          func(root string) error
	AlignAttributesFunc func(ownershipInAnyCase, modeIfApplicable, timesIfApplicable bool, optionalRoot ...string) (err error)
	VerifyFunc          func(root string) error
	StringFunc          func() string
	GetPathFunc         func() string
	GetModeFunc         func() os.FileMode
	GetAccessedFunc     func() time.Time
	GetModifiedFunc     func() time.Time
	GetUidFunc          func() uint32
	GetGidFunc          func() uint32
}

func NewFile() File {
	return File{
		CreateFunc:          func(root string) error { return nil },
		StringFunc:          func() string { return "String() result" },
		VerifyFunc:          func(root string) error { return nil },
		AlignAttributesFunc: func(ownershipInAnyCase, modeIfApplicable, timesIfApplicable bool, optionalRoot ...string) (err error) { return nil },
	}
}

func (m File) Create(root string) error                                                    { return m.CreateFunc(root) }
func (m File) AlignAttributes(owner, mode, times bool, optionalRoot ...string) (err error) { return m.AlignAttributesFunc(owner, mode, times, optionalRoot...) }
func (m File) Verify(root string) error                                                    { return m.VerifyFunc(root) }
func (m File) String() string                                                              { return m.StringFunc() }
func (m File) GetPath() string                                                             { return m.GetPathFunc() }
func (m File) GetMode() os.FileMode                                                        { return m.GetModeFunc() }
func (m File) GetAccessed() time.Time                                                      { return m.GetAccessedFunc() }
func (m File) GetModified() time.Time                                                      { return m.GetModifiedFunc() }
func (m File) GetUid() uint32                                                              { return m.GetUidFunc() }
func (m File) GetGid() uint32                                                              { return m.GetGidFunc() }
