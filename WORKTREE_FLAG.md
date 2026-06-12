# `--worktree` flag for `hack` and `append`

## Summary

Add a `--worktree` flag to `git town hack` and `git town append` that creates
the new branch in a **new Git worktree** instead of checking it out in the
current worktree. A `git-town.create-worktree` config setting makes this the
default behavior; `--worktree` / `--no-worktree` override it per invocation.

The feature targets a worktree-centric workflow where each branch lives in its
own directory alongside the others.

## Motivation

Two common worktree layouts are supported by a single rule.

### Bare repository layout

```
code/
  my-project/
    .git/        <- bare metadata (git clone --bare <repo>.git .git)
    main/        <- worktree for the main branch
    feature1/    <- worktree for feature1
    feature2/    <- worktree for feature2
```

Running `git town hack --worktree feature3` from anywhere inside the project
(`my-project/`, `my-project/main/`, `my-project/feature1/`, ...) creates the
new worktree at `code/my-project/feature3`.

### Regular repository layout

```
code/
  my-project/    <- regular repo, main branch checked out here
    .git/
    src/
  feature1/      <- worktree for feature1
  feature2/      <- worktree for feature2
```

Running `git town hack --worktree feature3` creates the new worktree at
`code/feature3`.

## Design decisions

### 1. Worktree location

The new worktree is created as a **sibling of the worktree that has the `main`
branch checked out**, named after the new branch:

```
new worktree path = <parent-of-main-branch-worktree> / <branch-name>
```

| `main` checked out at      | New branch | New worktree path                 |
| -------------------------- | ---------- | --------------------------------- |
| `code/my-project/main`     | `feature3` | `code/my-project/feature3`        |
| `code/my-project` (direct) | `feature3` | `code/feature3`                   |

This rule is **location-independent**: it produces the same result no matter
which directory the command runs from, including the bare-repo container
directory where there is no "current worktree".

Branch names containing slashes are **preserved as nested paths**
(`feature/foo` → `<parent>/feature/foo`), so the directory always matches the
branch name. `git worktree add` creates intermediate directories. The tradeoff
is that an empty prefix directory (`feature/`) may linger after the worktree is
removed.

The path is computed, not configurable beyond this rule.

### 2. End state and shell behavior

- The new branch is **always** created and checked out in the **new worktree**,
  never in the current worktree.
- The user's shell working directory **never changes**. Git Town runs as a
  subprocess; it may `chdir` into the new worktree *internally* to run commit /
  cherry-pick / stash-pop operations, but that does not affect the parent shell.
- After creating the worktree, Git Town prints the new worktree's path so the
  user can `cd` into it.

```
$ pwd
code/my-project/main
$ git town hack --worktree feature3
Created worktree at code/my-project/feature3
$ pwd
code/my-project/main   <- unchanged
```

### 3. Uncommitted changes

Open changes always **move** to the new worktree, mirroring normal `hack` /
`append` semantics (which move WIP onto the new branch). The current worktree
ends up clean; the new worktree carries the changes.

Mechanism: stash in the current worktree → `git worktree add` → `chdir` into the
new worktree → `git stash pop`. `git stash` is repo-global, so the stash entry
is available from the new worktree.

### 4. Flag compatibility

All existing flags are supported. There are no incompatible-flag errors.

| Flag                          | Behavior with `--worktree`                                     |
| ----------------------------- | ------------------------------------------------------------- |
| `--commit` / `--commit-message` | Moved WIP is committed onto the new branch in the new worktree |
| `--beam`                      | Selected commits move from the current branch onto the new branch (cherry-pick in the new worktree, removal on the current branch in the current worktree) |
| `--propose`                   | New branch is pushed and a proposal is opened (forge op, worktree-independent) |
| `--prototype`                 | Sets the branch type (worktree-independent)                    |
| `--detached`                  | Affects which branches sync (worktree-independent)            |
| `--sync`                      | Syncs the parent; sync of `main` is skipped when `main` is in another worktree (existing behavior) |
| `--stash`                     | Moot — WIP movement is handled by the worktree flow            |
| `--dry-run`, `--verbose`, `--interactive`, `--auto-resolve` | Orthogonal, unchanged          |

### 5. Sync of the parent branch

`hack` bases the new branch on `main`; `append` on the current branch. Git Town
already **skips syncing any branch that is checked out in another worktree**
(`SyncStatusOtherWorktree`, see `internal/cmd/sync/sync_branch.go`). When `main`
lives in another worktree (or we run from a bare repo), it cannot be
fast-forwarded locally, so the **local `main` ref may be stale**.

To keep the new branch current in that situation, the base (start point) is
chosen as follows:

- `git fetch` runs first (unless offline).
- If `main` cannot be synced locally (it is checked out in another worktree, or
  we are in a bare repo) **and** a remote tracking branch `origin/main` exists,
  the new worktree is created off `origin/main` with `--no-track` (so the new
  branch does not track `origin/main`).
- Otherwise the new worktree is created off local `main`.

This rule applies both to the bare-container case and to the case where `main`
is in another worktree while running from a different worktree, so the two behave
consistently.

### 6. Configuration

A new boolean config setting `git-town.create-worktree`:

```
git config git-town.create-worktree true

git town hack feature3            # -> creates a worktree
git town hack --no-worktree foo   # -> in-place (override)
```

The flag is tri-state (set true / set false / unset → use config), following
the existing pattern (e.g. `detached`). The setting is also exposed in the
interactive **setup assistant** (`git town setup`).

### 7. Undo

`git town undo` performs a **full reversal**:

- Move WIP back from the new worktree to the original worktree.
- Restore beamed commits onto the original branch.
- `git worktree remove` the new worktree.
- Delete the new branch.

Because Git Town's undo is snapshot-diff based and snapshots capture branch /
commit SHAs and stash size — **not** arbitrary working-tree contents across
directories — moved WIP cannot be reversed by the diff engine alone. Full
reversal requires explicit worktree-aware undo steps: reverse stash-transport
and `git worktree remove` **before** the branch deletion (a checked-out branch
cannot be deleted).

### 8. Failure modes

Git Town aborts with a clear error, without modifying the repository, when:

1. The computed worktree directory **exists and is non-empty**.
2. The **anchor cannot be resolved** (a non-bare repo where `main` is not checked
   out in any worktree, so there is no directory to anchor against).
3. The **branch already exists** locally or remotely (existing `hack` / `append`
   behavior, unchanged).

### 9. Running from a bare repository container

A common worktree setup uses a bare repository whose worktrees are siblings
inside a container directory:

```
my-project/
  .git/        <- bare repository
  main/        <- worktree for main
  feature1/    <- worktree for feature1
```

Running `git town hack --worktree feature2` from `my-project/` itself (the bare
container, **not** inside any worktree) is supported. Because `git worktree add`
works fine from a bare repository, the new worktree is created at
`my-project/feature2`.

This requires Git Town to operate **without a working tree**, which the rest of
Git Town normally assumes. The scope is therefore deliberately narrow:

- **`hack` only.** `append --worktree` run from the bare container errors,
  because it has no current branch to use as a parent. (`append --worktree` still
  works from inside a worktree.)
- **Worktree mode is required.** A plain in-place `hack` (worktree mode off) from
  the bare container errors with guidance to use `--worktree` (or to set
  `create-worktree`), since there is no working tree to check the branch out
  into. With `create-worktree` enabled, `hack feature2` from the bare container
  just works.
- **Path / anchor.** The new worktree is created at `<container>/<branch>`, where
  `<container>` is the parent of `git rev-parse --git-common-dir`. This is the
  same directory whether `main` already has a worktree at `<container>/main` or
  the bare repo has no worktrees yet (a fresh bare clone).
- **Root directory.** Git Town uses the container (parent of the common Git dir)
  as its repo root for loading the config file and storing runstate. This value
  is identical whether computed from the container or from any worktree.
- **Base ref.** As in section 5: fetch, then base off `origin/main` (`--no-track`)
  when it exists, else local `main`.
- **Flags.** `--commit` and `--beam` error from the bare container (there is no
  working tree to take changes from and no current branch to move commits off).
  `--propose`, `--prototype`, `--sync`, `--detached` and the rest work.
- **Undo.** `git town undo` run from the bare container reverses the operation
  (`git worktree remove` + delete the branch), skipping any checkout step since
  there is no working tree to check out into.

Detection uses `git rev-parse --is-bare-repository`. Permission to proceed in a
bare repository is opt-in per command (only `hack` and `undo` allow it); all
other commands keep rejecting bare repositories as before.

## Out of scope (for the first iteration)

- `prepend` and other branch-creating commands (can reuse the same opcodes
  later).
- A configurable worktree base location beyond the rules above.
- Shell integration that automatically `cd`s into the new worktree.
- `append --worktree` from the bare container (no current branch).
