package opcodes

import (
	"github.com/git-town/git-town/v23/internal/vm/shared"
)

// WorktreeRemove removes the worktree at the given path.
type WorktreeRemove struct {
	Path string
}

func (self *WorktreeRemove) Run(args shared.RunArgs) error {
	return args.Git.RemoveWorktree(args.Frontend, self.Path)
}
