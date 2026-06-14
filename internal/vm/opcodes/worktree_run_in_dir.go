package opcodes

import (
	"os"

	"github.com/git-town/git-town/v23/internal/config/configdomain"
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/gohacks/cache"
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
// Inside the target worktree the process's cached "current branch" (the branch
// of the original worktree) is stale: commands here run on the branch checked
// out in the target worktree. The branch cache is therefore invalidated while
// inside the worktree, so branch resolution and verbose output re-query the
// branch from the target worktree, and invalidated again on the way out so
// later opcodes back at the repo root re-read the original branch.
//
// In dry-run mode the new worktree was never actually created (the worktree-add
// command only printed), so there is no directory to change into. The chdir and
// cache handling are skipped and fn still runs against the dry-run frontend,
// which prints the intended commands without executing them.
func runInWorktree(path string, dryRun configdomain.DryRun, currentBranchCache *cache.WithPrevious[gitdomain.LocalBranchName], fn func() error) error {
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
	currentBranchCache.Invalidate()
	defer currentBranchCache.Invalidate()
	return fn()
}
