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
// created, anchoring them as siblings of the existing worktrees:
//
//   - if the anchor branch (usually main) is checked out in a separate worktree,
//     new worktrees are created next to that worktree;
//   - otherwise they are created next to the current worktree.
//
// For the common flat layouts (all worktrees inside one container directory)
// these produce the same directory, so the result is independent of where the
// command runs.
//
// currentWorktreeRoot is the root of the worktree the command runs in (the repo
// root), or empty when running from a bare repository that has no working tree.
func WorktreeParentDir(snapshot gitdomain.BranchesSnapshot, anchor gitdomain.LocalBranchName, currentWorktreeRoot string) (string, error) {
	if branchInfo, hasBranchInfo := snapshot.Branches.FindByLocalName(anchor).Get(); hasBranchInfo {
		if worktreePath, hasWorktreePath := branchInfo.WorktreePath.Get(); hasWorktreePath {
			// the anchor is checked out in another worktree
			return filepath.Dir(worktreePath), nil
		}
	}
	if currentWorktreeRoot != "" {
		// the anchor is not in a separate worktree: place new worktrees next to the current one
		return filepath.Dir(currentWorktreeRoot), nil
	}
	return "", fmt.Errorf(messages.WorktreeAnchorNoWorktree, anchor)
}
