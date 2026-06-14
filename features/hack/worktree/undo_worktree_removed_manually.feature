@skipWindows
Feature: undo a worktree hack after the worktree directory was deleted by hand

  # Removing the worktree directory manually leaves a stale registration. Undo
  # must still clear it and delete the branch instead of aborting when it cannot
  # chdir into the missing directory.

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And I ran "git-town hack --worktree feature"
    And I ran "rm -rf ../feature"

  Scenario: undo
    When I run "git-town undo"
    Then Git Town runs the commands
      | BRANCH | COMMAND                                           |
      | main   | git worktree remove {{ worktree-path "feature" }} |
      |        | git branch -D feature                             |
    And the current branch is still "main"
    And the branches are now
      | REPOSITORY    | BRANCHES |
      | local, origin | main     |
    And no lineage exists now
