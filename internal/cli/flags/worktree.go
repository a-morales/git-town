package flags

import (
	"github.com/git-town/git-town/v23/internal/config/configdomain"
	. "github.com/git-town/git-town/v23/pkg/prelude"
	"github.com/spf13/cobra"
)

const worktreeLong = "worktree"

// type-safe access to the CLI arguments of type configdomain.CreateWorktree
func Worktree() (AddFunc, ReadWorktreeFlagFunc) {
	addFlag := func(cmd *cobra.Command) {
		cmd.Flags().Bool(worktreeLong, false, "create the new branch in a new worktree")
		defineNegatedFlag(cmd.Flags(), worktreeLong, "create the new branch in the current worktree")
	}
	readFlag := func(cmd *cobra.Command) (Option[configdomain.CreateWorktree], error) {
		return readNegatableFlag[configdomain.CreateWorktree](cmd.Flags(), worktreeLong)
	}
	return addFlag, readFlag
}

// ReadWorktreeFlagFunc is the type signature for the function that reads the "worktree" flag from the args to the given Cobra command.
type ReadWorktreeFlagFunc func(*cobra.Command) (Option[configdomain.CreateWorktree], error)
