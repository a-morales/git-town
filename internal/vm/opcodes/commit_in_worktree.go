package opcodes

import (
	"github.com/git-town/git-town/v23/internal/config/configdomain"
	"github.com/git-town/git-town/v23/internal/git/gitdomain"
	"github.com/git-town/git-town/v23/internal/vm/shared"
	. "github.com/git-town/git-town/v23/pkg/prelude"
)

// CommitInWorktree commits all open changes in the worktree at the given path
// (instead of the current worktree) as a new commit. Used to commit changes that
// were transported into a newly created worktree.
type CommitInWorktree struct {
	AuthorOverride                 Option[gitdomain.Author]
	FallbackToDefaultCommitMessage bool
	Message                        Option[gitdomain.CommitMessage]
	Path                           string
}

func (self *CommitInWorktree) Run(args shared.RunArgs) error {
	return runInWorktree(self.Path, args.Config.Value.NormalConfig.DryRun, func() error {
		return args.Git.Commit(args.Frontend, configdomain.UseMessageWithFallbackToDefault(self.Message, self.FallbackToDefaultCommitMessage), self.AuthorOverride, configdomain.CommitHookEnabled)
	})
}
