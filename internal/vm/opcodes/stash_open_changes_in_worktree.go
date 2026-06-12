package opcodes

import (
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// StashOpenChangesInWorktree stashes the uncommitted changes in the worktree at
// the given path (instead of the current worktree). Used when undoing to move
// transported changes out of a worktree before it is removed.
type StashOpenChangesInWorktree struct {
	Path string
}

func (self *StashOpenChangesInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, func() error {
		return args.Git.Stash(args.Frontend)
	})
}
