package opcodes

import (
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// CherryPickAbortInWorktree aborts a suspended cherry-pick operation in the
// worktree at the given path (instead of the current worktree). Used to abort a
// CherryPickInWorktree that stopped on a conflict.
type CherryPickAbortInWorktree struct {
	Path string
}

func (self *CherryPickAbortInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, args.Git.CurrentBranchCache, func() error {
		return args.Git.CherryPickAbort(args.Frontend)
	})
}
