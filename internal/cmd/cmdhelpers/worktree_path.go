package cmdhelpers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/messages"
)

// CheckWorktreePathAvailable verifies that a new worktree can be created at the
// given path, reporting failure mode #1 (path already taken) before any Git
// command runs. It returns an error when the path is already in use:
//
//   - by an existing or registered worktree (per the branches snapshot, which
//     still reports the path even when its directory was deleted but not pruned);
//   - on disk by a file or a non-empty directory.
//
// A non-existent path or an empty directory is considered available - "git
// worktree add" creates intermediate directories and reuses an empty one.
func CheckWorktreePathAvailable(path string, snapshot gitdomain.BranchesSnapshot) error {
	for _, branchInfo := range snapshot.Branches {
		if worktreePath, hasWorktreePath := branchInfo.WorktreePath.Get(); hasWorktreePath && worktreePath == path {
			return fmt.Errorf(messages.WorktreePathInUse, path)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		// the path does not exist (or cannot be inspected) - treat as available
		// and let "git worktree add" surface any remaining problem.
		return nil
	}
	if !info.IsDir() {
		return fmt.Errorf(messages.WorktreePathExists, path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf(messages.WorktreePathExists, path)
	}
	if len(entries) > 0 {
		return fmt.Errorf(messages.WorktreePathExists, path)
	}
	return nil
}

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
