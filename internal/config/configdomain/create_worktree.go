package configdomain

import "strconv"

// CreateWorktree indicates whether the "hack" command should
// create the new branch in a new Git worktree instead of the current one.
type CreateWorktree bool

func (self CreateWorktree) ShouldCreateWorktree() bool {
	return bool(self)
}

func (self CreateWorktree) String() string {
	return strconv.FormatBool(bool(self))
}
