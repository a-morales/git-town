package opcodes

import (
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// ChangesUnstageAllInWorktree unstages all staged changes in the worktree at the
// given path (instead of the current worktree). Used when resuming a
// StashPopInWorktree after the user resolved a transport conflict, to turn the
// resolved changes back into ordinary uncommitted changes.
type ChangesUnstageAllInWorktree struct {
	Path string
}

func (self *ChangesUnstageAllInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, args.Git.CurrentBranchCache, func() error {
		return args.Git.UnstageAll(args.Frontend)
	})
}
