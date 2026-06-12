package cmd

import (
	"testing"

	"github.com/git-town/git-town/v23/internal/config/configdomain"
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/vm/opcodes"
	"github.com/git-town/git-town/v23/internal/vm/program"
	. "github.com/git-town/git-town/v23/pkg/prelude"
	"github.com/shoenig/test/must"
)

func TestMoveCommitsToAppendedBranch(t *testing.T) {
	t.Parallel()

	t.Run("non-worktree mode checks out the branches around the cherry-pick", func(t *testing.T) {
		t.Parallel()
		prog := NewMutable(&program.Program{})
		data := appendFeatureData{
			commitsToBeam:  gitdomain.Commits{{Message: "beamed", SHA: "111111"}},
			createWorktree: configdomain.CreateWorktree(false),
			initialBranch:  "existing",
			initialBranchInfo: &gitdomain.BranchInfo{
				Local:      Some(gitdomain.BranchData{Name: "existing", SHA: "222222"}),
				RemoteName: None[gitdomain.RemoteBranchName](),
				RemoteSHA:  None[gitdomain.SHA](),
				SyncStatus: gitdomain.SyncStatusLocalOnly,
			},
			targetBranch: "new",
			worktreePath: "",
		}
		moveCommitsToAppendedBranch(prog, data, true)
		want := program.Program{
			&opcodes.CherryPick{SHA: "111111"},
			&opcodes.Checkout{Branch: "existing"},
			&opcodes.CommitRemove{SHA: "111111"},
			&opcodes.Checkout{Branch: "new"},
		}
		must.Eq(t, want, prog.Immutable())
	})

	t.Run("worktree mode cherry-picks in the new worktree and removes commits in place", func(t *testing.T) {
		t.Parallel()
		prog := NewMutable(&program.Program{})
		data := appendFeatureData{
			commitsToBeam:  gitdomain.Commits{{Message: "beamed", SHA: "111111"}},
			createWorktree: configdomain.CreateWorktree(true),
			initialBranch:  "existing",
			initialBranchInfo: &gitdomain.BranchInfo{
				Local:      Some(gitdomain.BranchData{Name: "existing", SHA: "222222"}),
				RemoteName: None[gitdomain.RemoteBranchName](),
				RemoteSHA:  None[gitdomain.SHA](),
				SyncStatus: gitdomain.SyncStatusLocalOnly,
			},
			targetBranch: "new",
			worktreePath: "/code/my-project/new",
		}
		moveCommitsToAppendedBranch(prog, data, true)
		want := program.Program{
			// cherry-pick onto the new branch, which lives in its own worktree
			&opcodes.CherryPickInWorktree{Path: "/code/my-project/new", SHA: "111111"},
			// remove the beamed commit from the source branch (already checked out here)
			&opcodes.CommitRemove{SHA: "111111"},
		}
		must.Eq(t, want, prog.Immutable())
	})
}
