package diff

type FileDifference uint64

const (
	All                       FileDifference = 1 << iota
	NotPresentOrNotAccessible  //not present or no permissions
	ModeType
	ModePerm
	Owner
	Group
	ModTime
	AccTime
	Size
	LinkTarget
	Contents
)
