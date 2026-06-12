package opcodes

import (
	"errors"

	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/messages"
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// WorktreeAddAndCheckoutNewBranch creates a new branch with the first existing entry
// from the given ancestor list as its parent, in a new worktree at the given path,
// and checks the new branch out in that worktree.
// This is the worktree-mode counterpart of BranchCreateAndCheckoutExistingParent.
type WorktreeAddAndCheckoutNewBranch struct {
	Ancestors gitdomain.LocalBranchNames // list of ancestors - uses the first existing ancestor in this list
	Branch    gitdomain.LocalBranchName
	Path      string
}

func (self *WorktreeAddAndCheckoutNewBranch) Run(args shared.RunArgs) error {
	var ancestorToUse gitdomain.BranchName
	if nearestAncestor, hasNearestAncestor := args.Git.FirstExistingBranch(args.Backend, self.Ancestors...).Get(); hasNearestAncestor {
		ancestorToUse = nearestAncestor.BranchName()
	} else {
		mainInfo, hasMainBranch := args.BranchInfos.FindLocalOrRemote(args.Config.Value.ValidatedConfigData.MainBranch).Get()
		if !hasMainBranch {
			return errors.New(messages.MainBranchNotFound)
		}
		ancestorToUse = mainInfo.GetLocalOrRemoteName()
	}
	if err := args.Git.CreateWorktree(args.Frontend, self.Path, self.Branch, ancestorToUse.Location()); err != nil {
		return err
	}
	args.FinalMessages.Addf(messages.WorktreeCreated, self.Path)
	return nil
}
