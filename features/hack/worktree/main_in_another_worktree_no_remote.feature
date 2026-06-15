Feature: hack a new branch into a worktree while main is active in another worktree without a remote

  # Without a remote tracking branch for "main" there is nothing more current to
  # base off, so the new worktree falls back to the local "main" ref.

  Background:
    Given a local Git repo
    And the branches
      | NAME     | TYPE    | PARENT | LOCATIONS |
      | existing | feature | main   | local     |
    And the current branch is "existing"
    And branch "main" is active in another worktree
    When I run "git-town hack --worktree new"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH   | COMMAND                                                |
      | existing | git worktree add -b new {{ worktree-path "new" }} main |
    And the current branch is still "existing"
    And this lineage exists now
      """
      main
        existing
        new
      """
