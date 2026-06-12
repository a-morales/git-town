package cmdhelpers

import (
	"fmt"
	"path/filepath"

	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/messages"
)

// WorktreePathFor computes the directory for a new worktree that holds the given
// branch, placed inside parentDir. Slashes in the branch name are preserved as
// nested directories (e.g. "feature/foo" -> "<parentDir>/feature/foo").
func WorktreePathFor(parentDir string, branch gitdomain.LocalBranchName) string {
	return filepath.Join(parentDir, branch.String())
}

// WorktreeParentDir provides the directory in which new worktrees should be
// created: the parent of the worktree in which the given anchor branch (usually
// the main branch) is checked out. This anchors new worktrees as siblings of the
// existing ones, independent of the current working directory.
//
// currentWorktreeRoot is the root of the worktree the command runs in (the repo
// root); it is used when the anchor branch is checked out in the current
// worktree, since the snapshot does not record a path for that branch.
func WorktreeParentDir(snapshot gitdomain.BranchesSnapshot, anchor gitdomain.LocalBranchName, currentWorktreeRoot string) (string, error) {
	branchInfo, hasBranchInfo := snapshot.Branches.FindByLocalName(anchor).Get()
	if !hasBranchInfo {
		return "", fmt.Errorf(messages.WorktreeAnchorBranchMissing, anchor)
	}
	if worktreePath, hasWorktreePath := branchInfo.WorktreePath.Get(); hasWorktreePath {
		// the anchor is checked out in another worktree
		return filepath.Dir(worktreePath), nil
	}
	if snapshot.Active.EqualSome(anchor) {
		// the anchor is checked out in the current worktree
		return filepath.Dir(currentWorktreeRoot), nil
	}
	return "", fmt.Errorf(messages.WorktreeAnchorNoWorktree, anchor)
}
