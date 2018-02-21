package file

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"fmt"
	"github.com/outo/filefactory/diff"
	"github.com/outo/filefactory/verify"
	"github.com/outo/filefactory/attr"
)

const (
	ErrorMessageRootCannotBeUsedWithAbsoluteMetaPath = "root cannot be used when meta path is absolute"
	ErrorMessageRelativePathHasToBeUsedWithRoot      = "relative meta path requires root"
	TimeLayout                                       = "2006-01-02 15:04:05.000000000"
)

// wrapper for file's (file in generic terms) attributes
type Meta struct {
	AttributesAligner
	Verifier
	fmt.Stringer
	Path                     string
	Mode                     os.FileMode
	Accessed                 time.Time
	Modified                 time.Time
	Uid                      uint32
	Gid                      uint32
	VerificationInstructions []verify.Instruction
}

func NewFromPath(path string) (meta Meta, err error) {
	info, err := impl.OsLstat(path)
	if err != nil {
		return
	}

	st := info.Sys().(*syscall.Stat_t)

	meta = Meta{
		Path:     path,
		Mode:     info.Mode(),
		Accessed: time.Unix(st.Atim.Sec, st.Atim.Nsec),
		Modified: time.Unix(st.Mtim.Sec, st.Mtim.Nsec),
		Uid:      st.Uid,
		Gid:      st.Gid,
	}
	return
}

func (m Meta) isSymlink() bool {
	return m.Mode&os.ModeSymlink != 0
}

func (m Meta) String() string {
	return fmt.Sprintf("%s %d %d %s %s %s", m.Mode.String(), m.Uid, m.Gid, m.Modified.Format(TimeLayout), m.Accessed.Format(TimeLayout), m.Path)
}

func (m Meta) Should(verification verify.Instruction) bool {
	doVerify := true
	for _, verificationInstruction := range m.VerificationInstructions {
		if verificationInstruction == verify.AllByDefault(false) {
			doVerify = false
			//keep going, allow to overwrite instruction
		}
	}
	for _, verificationInstruction := range m.VerificationInstructions {
		if verificationInstruction.Aspect == verification.Aspect {
			doVerify = verificationInstruction.Verify
			//keep going, allow to overwrite instruction
		}
	}
	return doVerify
}

func (m Meta) Verify(root string) (err error) {
	if len(m.VerificationInstructions) == 1 && m.VerificationInstructions[0] == verify.AllByDefault(false) {
		return
	}

	verr := &verify.Errors{}

	path := filepath.Join(root, m.Path)

	ex, err := impl.PathExists(path)
	if err != nil {
		verr.Add(diff.NotPresentOrNotAccessible, path, err)
		return verr
	} else if !ex {
		verr.Add(diff.NotPresentOrNotAccessible, path, errors.New(fmt.Sprintf("file does not exist")))
		return verr
	}

	meta, err := impl.WrapNewFromPath(path)
	if err != nil {
		return
	}

	if meta.Mode & os.ModeType != m.Mode & os.ModeType {
		verr.Add(diff.ModeType, path, errors.New(fmt.Sprintf("expected %s, actual %s", m.Mode, meta.Mode)))
		return verr
	}

	if m.Should(verify.ModePerm(true)) {
		if meta.Mode & os.ModePerm != m.Mode & os.ModePerm{
			verr.Add(diff.ModePerm, path, errors.New(fmt.Sprintf("expected %s, actual %s", m.Mode, meta.Mode)))
		}
	}

	if m.Should(verify.Uid(true)) {
		if meta.Uid != m.Uid {
			verr.Add(diff.Owner, path, errors.New(fmt.Sprintf("expected %d, actual %d", m.Uid, meta.Uid)))
		}
	}

	if m.Should(verify.Gid(true)) {
		if meta.Gid != m.Gid {
			verr.Add(diff.Group, path, errors.New(fmt.Sprintf("expected %d, actual %d", m.Gid, meta.Gid)))
		}
	}

	if m.Should(verify.AccessedTime(true)) {
		if !meta.Accessed.Equal(m.Accessed) {
			verr.Add(diff.AccTime, path, errors.New(fmt.Sprintf("expected %s, actual %s", m.Accessed.Format(TimeLayout), meta.Accessed.Format(TimeLayout))))
		}
	}

	if m.Should(verify.ModifiedTime(true)) {
		if !meta.Modified.Equal(m.Modified) {
			verr.Add(diff.ModTime, path, errors.New(fmt.Sprintf("expected %s, actual %s", m.Modified.Format(TimeLayout), meta.Modified.Format(TimeLayout))))
		}
	}

	return verr.MapToNilIfNone()
}

func (m Meta) AlignAttributes(ownership, mode, times bool, optionalRoot ...string) (err error) {

	var path string
	if filepath.IsAbs(m.Path) {
		if filepath.Join(optionalRoot...) != "" {
			return errors.New(ErrorMessageRootCannotBeUsedWithAbsoluteMetaPath)
		} else {
			path = m.Path
		}
	} else {
		if filepath.Join(optionalRoot...) == "" {
			return errors.New(ErrorMessageRelativePathHasToBeUsedWithRoot)
		}
		path = filepath.Join(filepath.Join(optionalRoot...), m.Path)
	}

	if ownership {
		err = impl.OsLchown(path, int(m.Uid), int(m.Gid))
		if err != nil {
			return err
		}
	}

	if mode && !m.isSymlink() {
		err = impl.OsChmod(path, m.Mode)
		if err != nil {
			return err
		}
	}

	if times && !m.isSymlink() {
		err = impl.OsChtimes(path, m.Accessed, m.Modified)
		if err != nil {
			return err
		}
	}
	return
}

// will interpret variadic input with attributes and instructions and set fields of this Meta
func (m *Meta) Populate(relPath string, attributesAndInstructions ...interface{}) {
	for _, attribute := range attributesAndInstructions {
		switch catt := attribute.(type) {
		case os.FileMode:
			m.Mode = catt
		case attr.AccessedTime:
			m.Accessed = time.Time(catt)
		case attr.ModifiedTime:
			m.Modified = time.Time(catt)
		case attr.Uid:
			m.Uid = uint32(catt)
		case attr.Gid:
			m.Gid = uint32(catt)
		case verify.Instruction:
			m.VerificationInstructions = append(m.VerificationInstructions, catt)
		}
	}
	m.Path = relPath
}


func (m Meta) GetPath() string        { return m.Path }
func (m Meta) GetMode() os.FileMode   { return m.Mode }
func (m Meta) GetAccessed() time.Time { return m.Accessed }
func (m Meta) GetModified() time.Time { return m.Modified }
func (m Meta) GetUid() uint32         { return m.Uid }
func (m Meta) GetGid() uint32         { return m.Gid }
