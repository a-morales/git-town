Feature: hack a new branch in a new worktree

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    When I run "git-town hack --worktree feature"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH | COMMAND                                                        |
      | main   | git fetch --prune --tags                                       |
      |        | git worktree add -b feature {{ worktree-path "feature" }} main |
    And the current branch is still "main"
    And this lineage exists now
      """
      main
        feature
      """

  Scenario: undo
    When I run "git-town undo"
    Then Git Town runs the commands
      | BRANCH | COMMAND                                           |
      | main   | git worktree remove {{ worktree-path "feature" }} |
      |        | git branch -D feature                             |
    And the current branch is still "main"
    And the initial branches and lineage exist now
