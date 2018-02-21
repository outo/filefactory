package verify

import (
	"fmt"
	"github.com/outo/filefactory/diff"
)

type Error struct {
	diff.FileDifference
	Path string
	Err  error
}

func NewErr(difference diff.FileDifference, path string, err error) Error {
	return Error{
		FileDifference: difference,
		Path:           path,
		Err:            err,
	}
}

type Errors struct {
	error
	CombinedFileDifference diff.FileDifference
	Errors                 []Error
}

func (ves *Errors) Add(fileDifference diff.FileDifference, path string, err error) {
	ves.CombinedFileDifference |= fileDifference
	ves.Errors = append(ves.Errors, Error{
		FileDifference: fileDifference,
		Path:           path,
		Err:            err,
	})
}

func (ves *Errors) Merge(err error) (nonVerificationError error) {
	if err != nil {
		if verrs, ok := err.(*Errors); !ok {
			nonVerificationError = err
		} else {
			for _, verr := range verrs.Errors {
				ves.Add(verr.FileDifference, verr.Path, verr.Err)
			}
		}
	}
	return
}

func (ves *Errors) IsFileNotPresentOrNotAccessible() bool {
	return ves.CombinedFileDifference&diff.NotPresentOrNotAccessible != 0
}

func (ves *Errors) IsFileTypeUnexpected() bool {
	return ves.CombinedFileDifference&diff.ModeType != 0
}

func (ves *Errors) MapToNilIfNone() error {
	if len(ves.Errors) == 0 {
		return nil
	}
	return ves
}

func (ves *Errors) Error() string {
	collated := ""
	for _, err := range ves.Errors {
		collated += fmt.Sprintf("%s\n", err.Err.Error())
	}
	return collated
}

func (ves *Errors) HasDifference(difference diff.FileDifference, absolutePath string) bool {
	for _, ve := range ves.Errors {
		if ve.FileDifference == difference && ve.Path == absolutePath {
			return true
		}
	}
	return false
}

func (ves *Errors) DifferenceFor(absolutePath string) (diffUnion diff.FileDifference) {
	for _, ve := range ves.Errors {
		if ve.Path == absolutePath {
			diffUnion |= ve.FileDifference
		}
	}
	return
}
