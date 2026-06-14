package opcodes

import (
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// CherryPickInWorktree cherry-picks the given commit in the worktree at the given
// path (instead of the current worktree). Used to beam commits onto a branch that
// is checked out in a newly created worktree.
type CherryPickInWorktree struct {
	Path string
	SHA  gitdomain.SHA
}

func (self *CherryPickInWorktree) Abort() []shared.Opcode {
	return []shared.Opcode{
		&CherryPickAbortInWorktree{Path: self.Path},
	}
}

func (self *CherryPickInWorktree) Continue() []shared.Opcode {
	return []shared.Opcode{
		&CherryPickContinueInWorktree{Path: self.Path},
	}
}

func (self *CherryPickInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, func() error {
		return args.Git.CherryPick(args.Frontend, self.SHA)
	})
}
