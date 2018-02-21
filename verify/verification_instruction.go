package verify

type Instruction struct {
	Verify     bool
	Aspect     string
}

func NewInstruction(verify bool, aspect string) Instruction {
	return Instruction{
		Verify:     verify,
		Aspect:     aspect,
	}
}

func AllByDefault(verify bool) Instruction  { return NewInstruction(verify, "all") }
func ModePerm(verify bool) Instruction      { return NewInstruction(verify, "mode-perm") }
func ModifiedTime(verify bool) Instruction  { return NewInstruction(verify, "modified") }
func AccessedTime(verify bool) Instruction  { return NewInstruction(verify, "accessed") }
func Uid(verify bool) Instruction           { return NewInstruction(verify, "uid") }
func Gid(verify bool) Instruction           { return NewInstruction(verify, "gid") }
func Size(verify bool) Instruction          { return NewInstruction(verify, "size") }
func SymlinkTarget(verify bool) Instruction { return NewInstruction(verify, "symlink-target") }
func Contents(verify bool) Instruction      { return NewInstruction(verify, "contents") }
