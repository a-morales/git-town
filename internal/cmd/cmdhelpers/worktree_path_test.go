package cmdhelpers_test

import (
	"os"
	"path/filepath"
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

	t.Run("anchor checked out in another worktree (bare/multi-worktree layout)", func(t *testing.T) {
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

	t.Run("anchor not checked out and no current worktree returns an error", func(t *testing.T) {
		t.Parallel()
		snapshot := gitdomain.BranchesSnapshot{
			Active: None[gitdomain.LocalBranchName](),
			Branches: gitdomain.BranchInfos{
				{
					Local:      Some(gitdomain.BranchData{Name: "main"}),
					SyncStatus: gitdomain.SyncStatusLocalOnly,
				},
			},
		}
		_, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "")
		must.Error(t, err)
	})

	t.Run("anchor not checked out, falls back to the current worktree (on another branch)", func(t *testing.T) {
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
		have, err := cmdhelpers.WorktreeParentDir(snapshot, "main", "/code/my-project")
		must.NoError(t, err)
		must.EqOp(t, "/code", have)
	})
}

func TestCheckWorktreePathAvailable(t *testing.T) {
	t.Parallel()
	emptySnapshot := gitdomain.BranchesSnapshot{
		Active:   None[gitdomain.LocalBranchName](),
		Branches: gitdomain.BranchInfos{},
	}

	t.Run("path does not exist", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "new")
		must.NoError(t, cmdhelpers.CheckWorktreePathAvailable(path, emptySnapshot))
	})

	t.Run("path is an empty directory", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "new")
		must.NoError(t, os.Mkdir(path, 0o700))
		must.NoError(t, cmdhelpers.CheckWorktreePathAvailable(path, emptySnapshot))
	})

	t.Run("path is a non-empty directory", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "new")
		must.NoError(t, os.Mkdir(path, 0o700))
		must.NoError(t, os.WriteFile(filepath.Join(path, "file.txt"), []byte("content"), 0o600))
		must.Error(t, cmdhelpers.CheckWorktreePathAvailable(path, emptySnapshot))
	})

	t.Run("path is an existing file", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "new")
		must.NoError(t, os.WriteFile(path, []byte("content"), 0o600))
		must.Error(t, cmdhelpers.CheckWorktreePathAvailable(path, emptySnapshot))
	})

	t.Run("path is registered as a worktree in the snapshot (directory already removed)", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "new") // never created on disk
		snapshot := gitdomain.BranchesSnapshot{
			Active: None[gitdomain.LocalBranchName](),
			Branches: gitdomain.BranchInfos{
				{
					Local:        Some(gitdomain.BranchData{Name: "other"}),
					SyncStatus:   gitdomain.SyncStatusOtherWorktree,
					WorktreePath: Some(path),
				},
			},
		}
		must.Error(t, cmdhelpers.CheckWorktreePathAvailable(path, snapshot))
	})
}
