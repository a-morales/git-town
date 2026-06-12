package opcodes

import "os"

// runInWorktree runs the given function with the process working directory
// temporarily changed to the given worktree path, restoring the original
// directory afterwards. This lets an opcode operate inside another worktree
// (e.g. to pop a stash or commit there) without affecting the user's shell or
// the working directory seen by subsequent opcodes and snapshots.
//
// os.Chdir mutates process-global state, but the interpreter runs opcodes
// sequentially in a single goroutine, so changing and restoring the directory
// within a single opcode is safe.
func runInWorktree(path string, fn func() error) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err = os.Chdir(path); err != nil {
		return err
	}
	defer func() { _ = os.Chdir(originalDir) }()
	return fn()
}
