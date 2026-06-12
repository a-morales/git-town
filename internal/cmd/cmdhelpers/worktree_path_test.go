package cmdhelpers_test

import (
	"testing"

	"github.com/git-town/git-town/v23/internal/cmd/cmdhelpers"
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	. "github.com/git-town/git-town/v23/pkg/prelude"
	"github.com/shoenig/test/must"
)

func TestWorktreePathFor(t *testing.T) {
	t.Parallel()
	t.Run("branch name with slashes is preserved as nested path", func(t *testing.T) {
		t.Parallel()
		have := cmdhelpers.WorktreePathFor("/code/my-project", "feature/foo")
		must.EqOp(t, "/code/my-project/feature/foo", have)
	})
	t.Run("simple branch name", func(t *testing.T) {
		t.Parallel()
		have := cmdhelpers.WorktreePathFor("/code/my-project", "feature3")
		must.EqOp(t, "/code/my-project/feature3", have)
	})
}

func TestWorktreeParentDir(t *testing.T) {
	t.Parallel()

	t.Run("anchor branch does not exist", func(t *testing.T) {
		t.Parallel()
		snapshot := gitdomain.BranchesSnapshot{
			Active:   Some[gitdomain.LocalBranchName]("feature1"),
			Branches: gitdomain.BranchInfos{},
		}
		_, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "/code/my-project/feature1")
		must.Error(t, err)
	})

	t.Run("anchor exists but is not checked out anywhere", func(t *testing.T) {
		t.Parallel()
		snapshot := gitdomain.BranchesSnapshot{
			Active: Some[gitdomain.LocalBranchName]("feature1"),
			Branches: gitdomain.BranchInfos{
				{
					Local:      Some(gitdomain.BranchData{Name: "main"}),
					SyncStatus: gitdomain.SyncStatusLocalOnly,
				},
			},
		}
		_, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "/code/my-project/feature1")
		must.Error(t, err)
	})

	t.Run("anchor is checked out in another worktree (bare layout)", func(t *testing.T) {
		t.Parallel()
		snapshot := gitdomain.BranchesSnapshot{
			Active: Some[gitdomain.LocalBranchName]("feature1"),
			Branches: gitdomain.BranchInfos{
				{
					Local:        Some(gitdomain.BranchData{Name: "main"}),
					SyncStatus:   gitdomain.SyncStatusOtherWorktree,
					WorktreePath: Some("/code/my-project/main"),
				},
			},
		}
		have, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "/code/my-project/feature1")
		must.NoError(t, err)
		must.EqOp(t, "/code/my-project", have)
	})

	t.Run("anchor is the current worktree's branch (regular layout)", func(t *testing.T) {
		t.Parallel()
		snapshot := gitdomain.BranchesSnapshot{
			Active: Some[gitdomain.LocalBranchName]("main"),
			Branches: gitdomain.BranchInfos{
				{
					Local:      Some(gitdomain.BranchData{Name: "main"}),
					SyncStatus: gitdomain.SyncStatusLocalOnly,
				},
			},
		}
		have, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "/code/my-project")
		must.NoError(t, err)
		must.EqOp(t, "/code", have)
	})
}
