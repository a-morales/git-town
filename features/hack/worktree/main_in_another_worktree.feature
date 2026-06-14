Feature: hack a new branch into a worktree while main is active in another worktree

  Background:
    Given a Git repo with origin
    And the branches
      | NAME     | TYPE    | PARENT | LOCATIONS     |
      | existing | feature | main   | local, origin |
    And the current branch is "existing"
    And branch "main" is active in another worktree
    When I run "git-town hack --worktree new"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH   | COMMAND                                                |
      | existing | git fetch --prune --tags                               |
      |          | git worktree add -b new {{ worktree-path "new" }} main |
    And the current branch is still "existing"
    And this lineage exists now
      """
      main
        existing
        new
      """

  Scenario: undo
    When I run "git-town undo"
    Then Git Town runs the commands
      | BRANCH   | COMMAND                                       |
      | existing | git worktree remove {{ worktree-path "new" }} |
      |          | git branch -D new                             |
    And the current branch is still "existing"
    And the initial branches and lineage exist now
