Feature: hack into a new worktree with uncommitted changes

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And an uncommitted file "wip.txt" with content "work in progress"
    When I run "git-town hack --worktree feature"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH  | COMMAND                                                        |
      | main    | git add -A                                                     |
      |         | git stash -m "Git Town WIP"                                    |
      |         | git worktree add -b feature {{ worktree-path "feature" }} main |
      | feature | git stash pop                                                  |
      |         | git restore --staged .                                         |
    And the current branch is still "main"
    And no uncommitted files exist now
    And this lineage exists now
      """
      main
        feature
      """

  Scenario: undo
    When I run "git-town undo"
    Then Git Town runs the commands
      | BRANCH  | COMMAND                                           |
      | feature | git add -A                                        |
      |         | git stash -m "Git Town WIP"                       |
      | main    | git worktree remove {{ worktree-path "feature" }} |
      |         | git branch -D feature                             |
      |         | git stash pop                                     |
      |         | git restore --staged .                            |
    And the current branch is still "main"
    And an uncommitted file "wip.txt" exists now
    And the initial branches and lineage exist now
