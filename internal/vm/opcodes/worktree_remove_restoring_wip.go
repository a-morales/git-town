package opcodes

import (
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// WorktreeRemoveRestoringWIP removes the worktree at the given path and deletes
// the branch checked out there, used when undoing a worktree-creating hack.
//
// If that worktree contains uncommitted changes (transported there by a previous
// hack), they are moved back into the current worktree: the changes are stashed
// inside the worktree, the worktree is removed, the branch is deleted, and the
// changes are popped back in the current worktree. A clean worktree is simply
// removed and its branch deleted, so the common case runs no stash commands.
type WorktreeRemoveRestoringWIP struct {
	Branch gitdomain.LocalBranchName
	Path   string
}

func (self *WorktreeRemoveRestoringWIP) Run(args shared.RunArgs) error {
	hasOpenChanges := false
	if err := runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, func() error {
		status, err := args.Git.RepoStatus(args.Backend)
		if err != nil {
			return err
		}
		hasOpenChanges = status.OpenChanges
		return nil
	}); err != nil {
		return err
	}
	if hasOpenChanges {
		stashSize, err := args.Git.StashSize(args.Backend)
		if err != nil {
			return err
		}
		args.PrependOpcodes(
			&StashOpenChangesInWorktree{Path: self.Path},
			&WorktreeRemove{Path: self.Path},
			&BranchLocalDelete{Branch: self.Branch},
			&StashPopIfNeeded{InitialStashSize: stashSize},
		)
	} else {
		args.PrependOpcodes(
			&WorktreeRemove{Path: self.Path},
			&BranchLocalDelete{Branch: self.Branch},
		)
	}
	return nil
}
