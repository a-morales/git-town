Feature: hacking into a new worktree with --commit but no uncommitted changes

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    When I run "git-town hack --worktree --commit -m work feature"

  Scenario: result
    Then Git Town prints the error:
      """
      you used "--commit" but there are no uncommitted changes to commit
      """
    And the current branch is still "main"
    And the initial branches and lineage exist now
