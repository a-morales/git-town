package opcodes

import (
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// CherryPickContinueInWorktree continues a suspended cherry-pick operation in the
// worktree at the given path (instead of the current worktree). Used to resume a
// CherryPickInWorktree that stopped on a conflict.
type CherryPickContinueInWorktree struct {
	Path string
}

func (self *CherryPickContinueInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, args.Git.CurrentBranchCache, func() error {
		return args.Git.CherryPickContinue(args.Frontend)
	})
}
