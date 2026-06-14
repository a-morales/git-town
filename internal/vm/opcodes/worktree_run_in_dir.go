package opcodes

import (
	"os"

	"github.com/git-town/git-town/v23/internal/config/configdomain"
)

// runInWorktree runs the given function with the process working directory
// temporarily changed to the given worktree path, restoring the original
// directory afterwards. This lets an opcode operate inside another worktree
// (e.g. to pop a stash or commit there) without affecting the user's shell or
// the working directory seen by subsequent opcodes and snapshots.
//
// os.Chdir mutates process-global state, but the interpreter runs opcodes
// sequentially in a single goroutine, so changing and restoring the directory
// within a single opcode is safe.
//
// In dry-run mode the new worktree was never actually created (the worktree-add
// command only printed), so there is no directory to change into. The chdir is
// therefore skipped and fn still runs against the dry-run frontend, which prints
// the intended commands without executing them.
func runInWorktree(path string, dryRun configdomain.DryRun, fn func() error) error {
	if bool(dryRun) {
		return fn()
	}
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
