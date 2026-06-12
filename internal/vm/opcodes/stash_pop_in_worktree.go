package opcodes

import (
	"github.com/git-town/git-town/v23/internal/messages"
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// StashPopInWorktree pops the most recent stash entry into the worktree at the
// given path (instead of the current worktree). Used to transport uncommitted
// changes into a newly created worktree.
//
// When Unstage is true the popped changes are unstaged afterwards, so they appear
// as ordinary uncommitted changes. When false they are left staged, which lets a
// following commit consume them.
type StashPopInWorktree struct {
	Path    string
	Unstage bool
}

func (self *StashPopInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, func() error {
		if err := args.Git.PopStash(args.Frontend); err != nil {
			args.FinalMessages.Add(messages.DiffConflictWithMain)
			return nil
		}
		if self.Unstage {
			return args.Git.UnstageAll(args.Frontend)
		}
		return nil
	})
}
