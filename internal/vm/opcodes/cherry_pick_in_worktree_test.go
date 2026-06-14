package opcodes_test

import (
	"testing"

	"github.com/git-town/git-town/v23/internal/vm/opcodes"
	"github.com/git-town/git-town/v23/internal/vm/shared"
	"github.com/shoenig/test/must"
)

func TestCherryPickInWorktree(t *testing.T) {
	t.Parallel()

	t.Run("Continue resumes the cherry-pick inside the same worktree", func(t *testing.T) {
		t.Parallel()
		opcode := &opcodes.CherryPickInWorktree{Path: "/code/my-project/new", SHA: "111111"}
		want := []shared.Opcode{
			&opcodes.CherryPickContinueInWorktree{Path: "/code/my-project/new"},
		}
		must.Eq(t, want, opcode.Continue())
	})

	t.Run("Abort aborts the cherry-pick inside the same worktree", func(t *testing.T) {
		t.Parallel()
		opcode := &opcodes.CherryPickInWorktree{Path: "/code/my-project/new", SHA: "111111"}
		want := []shared.Opcode{
			&opcodes.CherryPickAbortInWorktree{Path: "/code/my-project/new"},
		}
		must.Eq(t, want, opcode.Abort())
	})
}
