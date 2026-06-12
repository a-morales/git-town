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
- **Add `WorktreePath Option[string]` to `BranchInfo`.** `BranchesSnapshot`
  already fetches `worktreepath:%(worktreepath)` and discards it after computing
  the `Worktree bool` (`commands.go:946,968`); retain it instead. This single
  field feeds both anchor resolution (Slice 2) and undo (Slice 3), so it lands
  first. Update the parser and any `BranchInfo` constructors/tests.
- `MainWorktreeParentDir(branchInfos)`: read the `WorktreePath` of the `main`
  branch's `BranchInfo` and return its parent directory. Returns an error if
  `main` has no worktree path (failure mode #2). Pure function over the
  snapshot — no extra Git query.

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
  `git.WorktreeAdd` (path passed as an argument; runs fine from the repo root,
  no `chdir` needed). Replaces `BranchCreateAndCheckoutExistingParent` in
  worktree mode.
- `WorktreeRemove{ Path }` — runs `git.WorktreeRemove`; used by undo.
- Worktree-targeted mutation opcodes for the operations that must run *inside*
  the new worktree (stash-pop, commit, cherry-pick). Each is **atomic**: it
  captures the current dir, `os.Chdir`es into the worktree path, runs the
  existing `git.Commands` method, and `defer`s a `chdir` back to the original
  dir — all within one `Run()`. Examples:
  `StashPopInWorktree{ Path }`, `CommitInWorktree{ Path, Message, ... }`,
  `CherryPickInWorktree{ Path, SHA }`.

### Notes — why no standalone `ChangeDir` opcode

Both runners execute Git in the **process CWD**: `FrontendRunner` never sets
`subProcess.Dir`, and in production `BackendRunner` is built with
`Dir: None[string]()` (`internal/execute/open_repo.go:37`). The end snapshot is
taken **after** the program runs, via the backend, in the then-current CWD
(`finished.go`, `errored.go`, `exit_to_shell.go` all call
`Git.BranchesSnapshot(args.Backend)`), and `BranchesSnapshot` derives the
*active* branch from the **current worktree's HEAD**.

A cross-opcode `ChangeDir` that leaves the CWD changed would therefore (a) make
the end snapshot record the *new* worktree's branch as active, corrupting the
undo diff, and (b) break interrupt/resume, which restarts at the original root
and would run the remaining worktree-targeted opcodes in the wrong directory.

The **atomic chdir-with-defer per opcode** pattern avoids both: the process CWD
is always restored to the repo root after every opcode (snapshot-safe), each
opcode is self-contained (resume-safe), and the existing `git.Commands` logic is
reused (no reimplementation of commit/cherry-pick/hooks). `os.Chdir` is
process-global, but the interpreter is single-threaded and sequential, so this
is safe.

### Registration

`internal/vm/program/json.go` (de)serializes opcodes by type name via
`opcodes.Lookup` → `opcodes.All()`. `internal/vm/opcodes/all.go` is **generated**
(`generate_opcodes_all.sh`, `// DO NOT EDIT`). Just add the opcode files and run
`make fix`; registration is automatic.

### Tests

- Per-opcode unit tests where the existing opcodes have them.
- A test asserting the process CWD is unchanged after a worktree-targeted opcode
  runs (atomicity guarantee).

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

`undobranches/branch_changes.go` generates undo by diffing begin/end
`BranchInfos` — it has no Git access, so the worktree path comes from the
`BranchInfo.WorktreePath` field added in Slice 0.

### `internal/undo/undobranches/branch_changes.go`

- The new branch is detected as `LocalAdded`. Today undo emits
  `CheckoutIfNeeded` + `BranchLocalDelete`, which fails for a branch checked out
  in a worktree. Make the added-branch undo path worktree-aware: when the
  added branch's end-snapshot `WorktreePath` is set, emit
  `WorktreeRemove{Path}` **before** `BranchLocalDelete` (and skip the
  `CheckoutIfNeeded`, since we are not checked out on the added branch in the
  current worktree).

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

- When there are open changes: prepend `StashOpenChanges` (runs in the current
  worktree at the repo root), then after `WorktreeAddAndCheckoutNewBranch` emit
  the atomic `StashPopInWorktree{newPath}`. No standalone `ChangeDir` — the pop
  opcode chdirs in and back internally.

### Undo

- Full reversal: when undoing, the new worktree may be dirty. Reverse the
  transport with atomic opcodes — `StashOpenChangesInWorktree{newPath}` (stash
  the WIP inside the new worktree) → `StashPopIfNeeded` (pop in the original
  worktree at the repo root) → `WorktreeRemove{newPath}`. Sequence so WIP lands
  back in the original worktree before the worktree is removed.

### Tests

- Cucumber: open changes present → after command, current worktree clean, new
  worktree carries the changes; `undo` moves them back.

## Slice 5 — `--commit` / `--commit-message`

### `appendProgram`

- After WIP is popped in the new worktree (Slice 4), emit the atomic
  `CommitInWorktree{newPath, ...}` so the commit lands on the new branch in the
  new worktree.

### Tests

- Cucumber: `hack --worktree --commit -m "msg"` → commit lands on the new branch
  in the new worktree; current worktree clean; `undo` reverses.

## Slice 6 — `--beam`

### `appendProgram` (`moveCommitsToAppendedBranch`)

- Cherry-pick the beamed commits in the **new worktree** via the atomic
  `CherryPickInWorktree{newPath, SHA}` opcodes.
- `CommitRemove` and the optional `PushCurrentBranchForceIgnoreError` run on the
  **current** branch in the **current** worktree at the repo root (unchanged
  opcodes — no chdir).

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

## Slice 11 — Base off `origin/main` when `main` can't be synced locally

Refines the start point used by `WorktreeAddAndCheckoutNewBranch` so the new
branch is current even when local `main` is stale (it lives in another worktree,
or we are in a bare repo). See WORKTREE_FLAG.md section 5.

### `internal/git/commands.go`

- Extend `CreateWorktree` to accept a start point that may be a remote ref and add
  `--no-track` in that case (mirror `CreateAndCheckoutBranchWithParent`, which
  appends `--no-track` when the parent `IsRemoteBranchName()`).

### `WorktreeAddAndCheckoutNewBranch` opcode

- After resolving the ancestor branch (e.g. `main`), decide the start point:
  - if the ancestor is checked out in another worktree (or the repo is bare)
    **and** a tracking branch `origin/<ancestor>` exists in `args.BranchInfos`,
    use that remote ref as the start point (with `--no-track`);
  - otherwise use the local ancestor branch.
- Ensure a `git fetch` still precedes this (the existing sync/fetch step in
  `appendProgram` already fetches unless offline / open changes).

### Tests

- Update `features/hack/worktree/main_in_another_worktree.feature`: the expected
  command becomes `git worktree add -b new --no-track {{ worktree-path "new" }}
  origin/main` (main is in another worktree, so we base off `origin/main`).
- Add a scenario without a remote (offline / no origin) asserting the fallback to
  local `main`.

## Slice 12 — Running `hack --worktree` from a bare repository container

Implements WORKTREE_FLAG.md section 9: `git town hack --worktree feature2` run
from the bare container directory (no working tree). Depends on Slices 2, 3, 11.

### Detection and `OpenRepo` (`internal/execute/open_repo.go`)

- Add `AllowBare bool` to `OpenRepoArgs`; `hack` and `undo` set it true, all other
  commands leave it false (so they keep rejecting bare repos).
- Query `git rev-parse --is-bare-repository`. When bare:
  - if `!AllowBare`, keep returning `messages.RepoOutside`;
  - if `AllowBare`, do **not** error on the missing work-tree root. Set the repo
    root to the **parent of `git rev-parse --git-common-dir`** (the container).
    Add a `git.Commands` helper `CommonDirParent` / reuse a `RootDirectory`-style
    query. Skip the startup `chdir`-to-root if already there.
  - return an `IsBare bool` (or `Option`-typed root) on `OpenRepoResult` so the
    command can branch on it.

### `internal/cmd/hack.go` (`determineHackData`)

- When `repo.IsBare`:
  - **require worktree mode**; if worktree mode is off, return a clear error
    (`messages.WorktreeBareNeedsFlag`).
  - error on `--commit` / `--beam` (`messages.WorktreeBareNoCommit` /
    `messages.WorktreeBareNoBeam`).
  - **skip `RepoStatus`** (`git status` fails with no work tree) → treat as no
    open changes.
  - compute the worktree path as `repo.RootDir/<targetBranch>` (the container is
    the anchor); still apply the path-exists failure check.
  - `initialBranch`: the worktree-mode program never references it, but
    `branchesSnapshot.Active` may report `main` (the bare HEAD). Ensure nothing in
    the bare path depends on a real current branch.
- Confirm `BranchesSnapshot` works from the bare repo (it uses `for-each-ref`, not
  `git status`) — it does.

### Undo from the bare container

- `undo` must also pass `AllowBare: true` to `OpenRepo`.
- The added-branch undo (Slice 3) already skips `CheckoutIfNeeded` for a branch in
  a worktree and emits `WorktreeRemove` before `BranchLocalDelete`; verify this
  path runs cleanly with no working tree (no stray checkout opcodes).

### Messages (`internal/messages/en.go`)

- `WorktreeBareNeedsFlag`, `WorktreeBareNoCommit`, `WorktreeBareNoBeam`.

### Tests

- New `features/hack/worktree/bare_container.feature`: set up a bare repo with a
  `main` worktree, run `git-town hack --worktree feature2` **from the container**;
  assert the `git worktree add ... origin/main` command, that `feature2` exists in
  a new worktree at `<container>/feature2`, and the undo scenario.
- Failure scenarios: plain `hack feature2` from the container errors;
  `hack --worktree --commit` / `--beam` from the container error.
- The cucumber harness already models a bare linked worktree
  (`AddBareRepoLinkedWorktree`); extend the fixture as needed to run a command
  from the bare container directory itself.

## Cross-cutting checklist

- [ ] Worktree-targeted opcodes are atomic (chdir in + `defer` chdir back); the
      process CWD is the repo root whenever `BranchesSnapshot` runs
      (`finished`/`errored`/`exit_to_shell`).
- [ ] Interrupted-command resume (runstate persistence) works with the new
      opcodes — no CWD state carried across opcodes; run `make fix` so
      `opcodes.All()` includes the new opcodes for (de)serialization.
- [ ] Dry-run prints the intended worktree operations without executing them.
- [ ] Windows path handling for the computed worktree path.
- [ ] `git town undo` never removes a worktree without first recovering its WIP.
- [ ] Run the end-to-end suite (see `docs/DEVELOPMENT.md`).

## Suggested PR breakdown

1. Slices 0–3 (primitives, opcodes, clean case, clean-case undo) — the core.
2. Slices 4–6 (WIP, commit, beam) with full-reversal undo.
3. Slice 7 (`append`).
4. Slices 8–9 (config + setup assistant).
5. Slice 11 (`origin/main` base ref) + Slice 12 (bare container support).
6. Slice 10 (docs).
