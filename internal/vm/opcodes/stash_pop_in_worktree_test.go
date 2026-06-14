package opcodes_test

import (
	"testing"

	"github.com/git-town/git-town/v23/internal/vm/opcodes"
	"github.com/git-town/git-town/v23/internal/vm/shared"
	"github.com/shoenig/test/must"
)

func TestStashPopInWorktree(t *testing.T) {
	t.Parallel()

	t.Run("Unstage resumes by unstaging the resolved changes in the worktree", func(t *testing.T) {
		t.Parallel()
		opcode := &opcodes.StashPopInWorktree{Path: "/code/my-project/new", Unstage: true}
		want := []shared.Opcode{
			&opcodes.ChangesUnstageAllInWorktree{Path: "/code/my-project/new"},
		}
		must.Eq(t, want, opcode.Continue())
	})

	t.Run("non-Unstage resumes without re-popping or unstaging (changes stay staged for the commit)", func(t *testing.T) {
		t.Parallel()
		opcode := &opcodes.StashPopInWorktree{Path: "/code/my-project/new", Unstage: false}
		must.Nil(t, opcode.Continue())
	})
}
