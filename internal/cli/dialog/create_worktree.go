package dialog

import (
	"fmt"

	"github.com/git-town/git-town/v23/internal/cli/dialog/dialogcomponents"
	"github.com/git-town/git-town/v23/internal/cli/dialog/dialogcomponents/list"
	"github.com/git-town/git-town/v23/internal/cli/dialog/dialogdomain"
	"github.com/git-town/git-town/v23/internal/config/configdomain"
	"github.com/git-town/git-town/v23/internal/messages"
	. "github.com/git-town/git-town/v23/pkg/prelude"
)

const (
	createWorktreeTitle = `Create worktree`
	CreateWorktreeHelp  = `
Should "git town hack" create the new branch in its own new Git worktree
instead of checking it out in the current worktree?

`
)

func CreateWorktree(args Args[configdomain.CreateWorktree]) (Option[configdomain.CreateWorktree], dialogdomain.Exit, error) {
	entries := list.Entries[Option[configdomain.CreateWorktree]]{}
	if global, hasGlobal := args.Global.Get(); hasGlobal {
		entries = append(entries, list.Entry[Option[configdomain.CreateWorktree]]{
			Data: None[configdomain.CreateWorktree](),
			Text: fmt.Sprintf(messages.DialogUseGlobalValue, global),
		})
	}
	entries = append(entries, list.Entries[Option[configdomain.CreateWorktree]]{
		{
			Data: Some(configdomain.CreateWorktree(false)),
			Text: "no, check the new branch out in the current worktree",
		},
		{
			Data: Some(configdomain.CreateWorktree(true)),
			Text: "yes, create a new worktree for the new branch",
		},
	}...)
	defaultPos := entries.IndexOf(args.Local)
	selection, exit, err := dialogcomponents.RadioList(entries, defaultPos, createWorktreeTitle, CreateWorktreeHelp, args.Inputs, args.Interactive, "create-worktree")
	fmt.Printf(messages.CreateWorktreeResult, dialogcomponents.FormattedOption(selection, args.Global.IsSome(), exit))
	return selection, exit, err
}
