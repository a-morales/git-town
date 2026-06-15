Feature: dry-run hacking into a new worktree with uncommitted changes

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And an uncommitted file "wip.txt" with content "work in progress"
    When I run "git-town hack --worktree --commit -m work feature --dry-run"

  Scenario: result
    Then an uncommitted file "wip.txt" exists now
    And Git Town runs the commands
      | BRANCH | COMMAND                                                        |
      | main   | git add -A                                                     |
      |        | git stash -m "Git Town WIP"                                    |
      |        | git worktree add -b feature {{ worktree-path "feature" }} main |
      |        | git stash pop                                                  |
      |        | git commit -m work                                             |
    And the current branch is still "main"
    And the initial branches and lineage exist now
  #
  # Cannot test undo because dry-run now doesn't create a runstate.
