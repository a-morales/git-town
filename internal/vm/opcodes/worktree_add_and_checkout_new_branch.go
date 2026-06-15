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
	startPoint, err := self.startPoint(args)
	if err != nil {
		return err
	}
	if err := args.Git.CreateWorktree(args.Frontend, self.Path, self.Branch, startPoint); err != nil {
		return err
	}
	args.FinalMessages.Addf(messages.WorktreeCreated, self.Path)
	return nil
}

// startPoint determines the Git ref the new worktree's branch is based on.
// It uses the nearest existing ancestor branch, but when that ancestor's local
// ref may be stale and a remote tracking branch for it exists, it bases off the
// remote ref instead so the new branch is current. CreateWorktree adds --no-track
// for a remote start point.
func (self *WorktreeAddAndCheckoutNewBranch) startPoint(args shared.RunArgs) (gitdomain.Location, error) {
	if ancestor, hasAncestor := args.Git.FirstExistingBranch(args.Backend, self.Ancestors...).Get(); hasAncestor {
		preferRemote, err := self.shouldPreferRemoteAncestor(args, ancestor)
		if err != nil {
			return "", err
		}
		if preferRemote {
			if remoteInfo, hasRemoteInfo := args.BranchInfos.FindRemoteNameMatchingLocal(ancestor).Get(); hasRemoteInfo {
				if remoteName, hasRemoteName := remoteInfo.RemoteName.Get(); hasRemoteName {
					return remoteName.BranchName().Location(), nil
				}
			}
		}
		return ancestor.Location(), nil
	}
	mainInfo, hasMainBranch := args.BranchInfos.FindLocalOrRemote(args.Config.Value.ValidatedConfigData.MainBranch).Get()
	if !hasMainBranch {
		return "", errors.New(messages.MainBranchNotFound)
	}
	return mainInfo.GetLocalOrRemoteName().Location(), nil
}

// shouldPreferRemoteAncestor reports whether the new worktree should be based on
// the ancestor's remote tracking branch rather than its local ref, because the
// local ref may be stale: the ancestor is checked out in another worktree (Git
// Town skips syncing such branches), or we run from a bare repository where the
// local ref is not fast-forwarded.
func (self *WorktreeAddAndCheckoutNewBranch) shouldPreferRemoteAncestor(args shared.RunArgs, ancestor gitdomain.LocalBranchName) (bool, error) {
	if args.BranchInfos.BranchIsActiveInAnotherWorktree(ancestor) {
		return true, nil
	}
	return args.Git.IsBareRepo(args.Backend)
}
