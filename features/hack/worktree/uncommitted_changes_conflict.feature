Feature: hacking into a new worktree when the uncommitted changes conflict

  Background:
    Given a Git repo with origin
    And the commits
      | BRANCH | LOCATION | MESSAGE     | FILE NAME | FILE CONTENT |
      | main   | local    | main commit | file.txt  | main content |
    And the branches
      | NAME     | TYPE    | PARENT | LOCATIONS |
      | existing | feature | main   | local     |
    And the commits
      | BRANCH   | LOCATION | MESSAGE         | FILE NAME | FILE CONTENT     |
      | existing | local    | existing commit | file.txt  | existing content |
    And the current branch is "existing"
    And an uncommitted file "file.txt" with content "wip content"
    When I run "git-town hack --worktree new"

  Scenario: result
    Then Git Town prints the error:
      """
      moving your uncommitted changes into the new worktree caused conflicts
      """
    And the current branch is still "existing"
