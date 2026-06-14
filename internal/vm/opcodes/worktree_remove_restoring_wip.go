package opcodes

import (
	"os"

	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/messages"
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// WorktreeRemoveRestoringWIP removes the worktree at the given path and deletes
// the branch checked out there, used when undoing a worktree-creating hack.
//
// If that worktree contains uncommitted changes (transported there by a previous
// hack), they are moved back into the current worktree: the changes are stashed
// inside the worktree, the worktree is removed, the branch is deleted, and the
// changes are popped back in the current worktree. This relocation is announced
// via a final message, since it can also pick up changes the user made in the
// worktree after the hack. A clean worktree is simply removed and its branch
// deleted, so the common case runs no stash commands. If the worktree directory
// was already removed by hand, the stale registration is cleared and the branch
// deleted without aborting.
type WorktreeRemoveRestoringWIP struct {
	Branch gitdomain.LocalBranchName
	Path   string
}

func (self *WorktreeRemoveRestoringWIP) Run(args shared.RunArgs) error {
	hasOpenChanges := false
	// Only inspect the worktree for uncommitted changes when its directory still
	// exists. If the user removed it by hand, there is nothing to restore and
	// chdir-ing into it would fail; "git worktree remove" below still clears the
	// now-stale registration.
	if info, statErr := os.Stat(self.Path); statErr == nil && info.IsDir() {
		if err := runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, args.Git.CurrentBranchCache, func() error {
			status, err := args.Git.RepoStatus(args.Backend)
			if err != nil {
				return err
			}
			hasOpenChanges = status.OpenChanges
			return nil
		}); err != nil {
			return err
		}
	}
	if hasOpenChanges {
		stashSize, err := args.Git.StashSize(args.Backend)
		if err != nil {
			return err
		}
		args.FinalMessages.Addf(messages.WorktreeWipRelocated, self.Path)
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
