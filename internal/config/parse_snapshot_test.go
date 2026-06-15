package config_test

import (
	"testing"

	"github.com/git-town/git-town/v23/internal/config"
	"github.com/git-town/git-town/v23/internal/config/configdomain"
	. "github.com/git-town/git-town/v23/pkg/prelude"
	"github.com/shoenig/test/must"
)

func TestNewPartialConfigFromSnapshotCreateWorktree(t *testing.T) {
	t.Parallel()
	t.Run("set to true", func(t *testing.T) {
		t.Parallel()
		snapshot := configdomain.SingleSnapshot{
			"git-town.create-worktree": "true",
		}
		have, err := config.NewPartialConfigFromSnapshot(snapshot, false, false, nil)
		must.NoError(t, err)
		must.Eq(t, Some(configdomain.CreateWorktree(true)), have.CreateWorktree)
	})
	t.Run("set to false", func(t *testing.T) {
		t.Parallel()
		snapshot := configdomain.SingleSnapshot{
			"git-town.create-worktree": "false",
		}
		have, err := config.NewPartialConfigFromSnapshot(snapshot, false, false, nil)
		must.NoError(t, err)
		must.Eq(t, Some(configdomain.CreateWorktree(false)), have.CreateWorktree)
	})
	t.Run("not set", func(t *testing.T) {
		t.Parallel()
		snapshot := configdomain.SingleSnapshot{}
		have, err := config.NewPartialConfigFromSnapshot(snapshot, false, false, nil)
		must.NoError(t, err)
		must.Eq(t, None[configdomain.CreateWorktree](), have.CreateWorktree)
	})
}

func TestNewBranchTypeOverridesFromSnapshot(t *testing.T) {
	t.Parallel()
	snapshot := configdomain.SingleSnapshot{
		"git-town-branch.branch-1.branchtype": "feature",
		"git-town-branch.branch-2.branchtype": "observed",
		"git-town-branch.branch-3.parent":     "main",
		"git-town.prototype-branches":         "foo",
	}
	have, err := config.NewBranchTypeOverridesInSnapshot(snapshot, false, nil)
	must.NoError(t, err)
	want := configdomain.BranchTypeOverrides{
		"branch-1": configdomain.BranchTypeFeatureBranch,
		"branch-2": configdomain.BranchTypeObservedBranch,
	}
	must.Eq(t, want, have)
}
