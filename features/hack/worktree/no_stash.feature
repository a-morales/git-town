Feature: hack into a new worktree with uncommitted changes and --no-stash

# In worktree mode the "stash" setting is moot: there is no
# checkout-carries-changes path across worktrees, so open changes always move
# into the new worktree even with --no-stash.

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And an uncommitted file "wip.txt" with content "work in progress"
    When I run "git-town hack --worktree --no-stash feature"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH  | COMMAND                                                        |
      | main    | git add -A                                                     |
      |         | git stash -m "Git Town WIP"                                    |
      |         | git worktree add -b feature {{ worktree-path "feature" }} main |
      | feature | git stash pop                                                  |
      |         | git restore --staged .                                         |
    And the current branch is still "main"
    And this lineage exists now
      """
      main
        feature
      """
    And no uncommitted files exist now
