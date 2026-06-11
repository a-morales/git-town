# Implementation plan: `--worktree` flag for `hack` and `append`

This plan implements the design in [WORKTREE_FLAG.md](WORKTREE_FLAG.md).

It is organized into incremental slices. Each slice compiles, passes tests, and
delivers a usable subset. The clean case lands first to prove the opcodes and
undo wiring; WIP/commit/beam and the config default layer on top.

## Slice 0 — Git primitives and path computation

Establish the low-level Git operations and the path rule, fully unit-tested,
before touching any command.

### `internal/git/commands.go` (+ `commands_test.go`)

- `WorktreeAdd(runner, path, branch, parent)` →
  `git worktree add -b <branch> <path> <parent>`.
- `WorktreeRemove(runner, path)` → `git worktree remove <path>`.
- `WorktreeList(...)` / reuse existing `git worktree list --porcelain` parsing
  to find the path where a given branch is checked out. (Branch→worktree path
  data already exists via the `worktreepath` field used in `BranchesSnapshot`;
  factor out a query that returns the worktree path for a branch.)
- `MainWorktreeParentDir(...)`: resolve the directory of the worktree holding
  the `main` branch, return its parent. Returns an error if `main` is not
  checked out in any worktree (failure mode #2).

### Path computation helper

- A pure function `WorktreePathFor(parentDir, branchName) string` that joins
  `parentDir` and the branch name (slashes preserved). Lives near the command
  layer (e.g. `internal/cmd/cmdhelpers/`) and is unit-tested in isolation,
  including slash-containing branch names and the two layouts from the design
  doc.

### Tests

- Unit tests for `WorktreeAdd` / `WorktreeRemove` against a real temp repo
  (follow the pattern in `internal/git/commands_test.go`, which already uses
  `AddWorktree`).
- Unit tests for `MainWorktreeParentDir` in both layouts and the
  no-anchor error case.
- Unit tests for `WorktreePathFor`.

## Slice 1 — Opcodes

New opcodes in `internal/vm/opcodes/`. Follow the existing single-`Run`-method
pattern (see `commit.go`, `branch_local_delete.go`).

- `WorktreeAddAndCheckoutNewBranch{ Branch, Path, Parent }` — runs
  `git.WorktreeAdd`. Replaces `BranchCreateAndCheckoutExistingParent` in
  worktree mode.
- `ChangeDir{ Path }` — `os.Chdir(path)` so subsequent opcodes run inside the
  target worktree. Add a corresponding "change back" usage (store the original
  dir, or emit an explicit `ChangeDir` back to the repo root).
- `WorktreeRemove{ Path }` — runs `git.WorktreeRemove`; used by undo.

### Notes

- `ChangeDir` changes the Git Town **process** working directory only; the
  parent shell is unaffected.
- Verify how the interpreter resolves the repo root after a `chdir` (the runner
  executes Git in the process CWD). Ensure snapshot/runstate operations that
  assume the repo root still work, or always `chdir` back before they run.
- Register opcodes wherever opcodes are enumerated for serialization (check how
  existing opcodes are registered for runstate persistence so an interrupted
  command can resume).

### Tests

- Per-opcode unit tests where the existing opcodes have them.

## Slice 2 — Clean case wiring (`hack`, no WIP/commit/beam)

Goal: `git town hack --worktree feature3` with a clean working tree creates the
worktree and branch, leaves the shell put, and prints the path. No config yet.

### `internal/cli/flags/`

- Add a `Worktree()` tri-state flag (set true / set false / unset), mirroring
  the existing `Detached()` flag. Provides `--worktree` and `--no-worktree`.

### `internal/cmd/hack.go`

- Wire `addWorktreeFlag` / `readWorktreeFlag` into `hackCmd`.
- Thread the value into `hackArgs` and on into `appendFeatureData`.

### `internal/cmd/append.go` (`appendProgram`)

- Add a `worktree` decision to `appendFeatureData` plus the resolved worktree
  `path` and `parent`.
- When worktree mode is on, replace:
  ```
  prog.Add(&opcodes.BranchCreateAndCheckoutExistingParent{...})
  ```
  with:
  ```
  prog.Add(&opcodes.WorktreeAddAndCheckoutNewBranch{Branch, Path, Parent})
  ```
- Skip the in-place checkout-back / stash `Wrap` for the clean case (the current
  worktree is never left, so there is nothing to restore).
- Keep `LineageParentSetFirstExisting` and `BranchTypeOverrideSet` (they operate
  on refs/config, not the working tree).

### `internal/cmd/hack.go` (`determineHackData`)

- Resolve the anchor and compute the target path here so failure mode #2 (no
  anchor) and #1 (path exists & non-empty) are reported **before** any opcode
  runs. Store path/parent in the data struct.

### `internal/messages/en.go`

- `WorktreeCreated` ("Created worktree at %q").
- `WorktreePathExists` / `WorktreeNoAnchor` error messages.

### Tests

- Cucumber feature `features/hack/worktree/` covering: bare layout and regular
  layout, path correctness, shell-cwd unchanged, branch created in the new
  worktree, current branch unchanged.
- Failure-mode features: directory exists & non-empty; no anchor.

## Slice 3 — Undo for the clean case

Goal: `git town undo` after a clean `hack --worktree` removes the worktree and
deletes the branch.

### `internal/undo/undobranches/branch_changes.go`

- The new branch is detected as `LocalAdded`. Today undo emits
  `CheckoutIfNeeded` + `BranchLocalDelete`, which fails for a branch checked out
  in a worktree. Make the added-branch undo path worktree-aware: when the added
  branch is checked out in a (non-main) worktree, emit `WorktreeRemove{path}`
  **before** `BranchLocalDelete`.
- Determine the worktree path during undo from the begin/end snapshots or by
  querying Git at undo time (the branch→worktree-path query from Slice 0).

### Tests

- Extend `internal/undo/undobranches/branch_changes_test.go` with a case where
  an added branch lives in a worktree → expect `WorktreeRemove` then
  `BranchLocalDelete`.
- Cucumber: `hack --worktree` followed by `git town undo` restores the original
  state (no worktree dir, no branch, original branch still checked out).

## Slice 4 — WIP transport

Goal: `hack --worktree` with uncommitted changes moves the WIP to the new
worktree.

### `appendProgram`

- When there are open changes: prepend `StashOpenChanges` (current worktree),
  then after `WorktreeAddAndCheckoutNewBranch`, emit `ChangeDir{newPath}` and
  `StashPopIfNeeded`, then `ChangeDir` back to the repo root.

### Undo

- Full reversal: when undoing, the new worktree may be dirty. Reverse the
  transport — `ChangeDir{newPath}` → `StashOpenChanges` → `ChangeDir{origRoot}`
  → `StashPopIfNeeded` → `WorktreeRemove`. Sequence the stash ops so WIP lands
  back in the original worktree before the worktree is removed.

### Tests

- Cucumber: open changes present → after command, current worktree clean, new
  worktree carries the changes; `undo` moves them back.

## Slice 5 — `--commit` / `--commit-message`

### `appendProgram`

- After WIP is popped in the new worktree (Slice 4), emit `Commit` while the
  process CWD is the new worktree, then `ChangeDir` back.

### Tests

- Cucumber: `hack --worktree --commit -m "msg"` → commit lands on the new branch
  in the new worktree; current worktree clean; `undo` reverses.

## Slice 6 — `--beam`

### `appendProgram` (`moveCommitsToAppendedBranch`)

- Cherry-pick the beamed commits in the **new worktree** (`ChangeDir{newPath}`
  around the `CherryPick` ops).
- `CommitRemove` and the optional `PushCurrentBranchForceIgnoreError` run on the
  **current** branch in the **current** worktree (`ChangeDir` back first).

### Undo

- Restore the removed commits on the current branch and remove the worktree;
  rely on the branch-changes snapshot diff for the commit-restore plus the
  worktree removal from Slice 3.

### Tests

- Cucumber: `hack --worktree --beam` moves selected commits to the new branch;
  current branch loses them; `undo` restores.

## Slice 7 — `append --worktree`

- Wire the same flag and data through `internal/cmd/append.go`'s command
  definition (parent = current branch; path still anchored to the main
  worktree's parent).
- Cucumber features mirroring the `hack` ones, with a non-`main` parent.

## Slice 8 — Config setting `git-town.create-worktree`

### `internal/config/configdomain/`

- New `CreateWorktree` bool type and accessor.
- Register the key `git-town.create-worktree` in config parsing/loading
  (follow an existing bool setting, e.g. how `detached`/`offline` style keys are
  read).

### Flag resolution

- In `hackCmd` / `appendCmd`: effective worktree mode =
  `flag if set, else config value, else false`.

### Tests

- Unit tests for parsing the config key.
- Cucumber: setting `git-town.create-worktree true` makes `hack <branch>` create
  a worktree; `--no-worktree` overrides.

## Slice 9 — Setup assistant

### `internal/setup/`

- Add a dialog step for `create-worktree` (a yes/no prompt), following an
  existing boolean dialog in the setup flow.
- Persist the answer through the existing setup write path.

### Tests

- Follow existing setup-assistant test patterns.

## Slice 10 — Documentation

- Website docs (`website/`) for `hack` / `append` describing `--worktree` and
  the `create-worktree` config setting.
- Changelog entry.

## Cross-cutting checklist

- [ ] `ChangeDir` always returns to the repo root before snapshot/runstate code
      runs.
- [ ] Interrupted-command resume (runstate persistence) works with the new
      opcodes — verify serialization registration.
- [ ] Dry-run prints the intended worktree operations without executing them.
- [ ] Windows path handling for the computed worktree path.
- [ ] `git town undo` never removes a worktree without first recovering its WIP.
- [ ] Run the end-to-end suite (see `docs/DEVELOPMENT.md`).

## Suggested PR breakdown

1. Slices 0–3 (primitives, opcodes, clean case, clean-case undo) — the core.
2. Slices 4–6 (WIP, commit, beam) with full-reversal undo.
3. Slice 7 (`append`).
4. Slices 8–9 (config + setup assistant).
5. Slice 10 (docs).
