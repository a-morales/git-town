package configdomain

// CreateWorktree indicates whether the "hack" command should
// create the new branch in a new Git worktree instead of the current one.
type CreateWorktree bool

func (self CreateWorktree) ShouldCreateWorktree() bool {
	return bool(self)
}
