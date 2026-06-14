package opcodes

import (
	"errors"

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
//
// If the pop conflicts, the command stops so the user can resolve the conflict
// inside the new worktree and then run "git town continue"; resuming does not
// re-pop (the stash entry is already consumed) - it only performs the remaining
// unstage step when one was requested.
type StashPopInWorktree struct {
	Path    string
	Unstage bool
}

func (self *StashPopInWorktree) Continue() []shared.Opcode {
	if self.Unstage {
		return []shared.Opcode{&ChangesUnstageAllInWorktree{Path: self.Path}}
	}
	return nil
}

func (self *StashPopInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, args.Git.CurrentBranchCache, func() error {
		if err := args.Git.PopStash(args.Frontend); err != nil {
			// The conflicted changes are now in the new worktree (PopStash applied
			// them and dropped the stash entry). Stop so the user can resolve them;
			// committing or continuing over an unmerged tree would fail or be wrong.
			return errors.New(messages.WorktreeWipConflict)
		}
		if self.Unstage {
			return args.Git.UnstageAll(args.Frontend)
		}
		return nil
	})
}
