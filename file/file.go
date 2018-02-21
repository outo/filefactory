package file

import (
	"fmt"
	"os"
	"time"
)

type Creator interface {
	Create(root string) error
}

type AttributesAligner interface {
	AlignAttributes(ownershipInAnyCase, modeIfApplicable, timesIfApplicable bool, optionalRoot ...string) (err error)
}

type Verifier interface {
	Verify(root string) error
}

type CommonAttributesGetter interface {
	GetPath() string
	GetMode() os.FileMode
	GetAccessed() time.Time
	GetModified() time.Time
	GetUid() uint32
	GetGid() uint32
}

type File interface {
	Creator
	AttributesAligner
	Verifier
	CommonAttributesGetter
	fmt.Stringer
}
