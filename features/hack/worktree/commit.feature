Feature: hack into a new worktree and commit the open changes

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And an uncommitted file "wip.txt" with content "work in progress"
    When I run "git-town hack --worktree --commit -m work feature"

  Scenario: result
    Then Git Town runs the commands
      | BRANCH | COMMAND                                                       |
      | main   | git add -A                                                    |
      |        | git stash -m "Git Town WIP"                                   |
      |        | git worktree add -b feature {{ worktree-path "feature" }} main |
      |        | git stash pop                                                 |
      |        | git commit -m work                                            |
    And the current branch is still "main"
    And no uncommitted files exist now

  Scenario: undo
    When I run "git-town undo"
    Then Git Town runs the commands
      | BRANCH | COMMAND                                           |
      | main   | git worktree remove {{ worktree-path "feature" }} |
      |        | git branch -D feature                             |
    And the current branch is still "main"
    And the initial branches and lineage exist now
