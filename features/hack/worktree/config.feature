Feature: the "create-worktree" config setting makes hack create worktrees by default

  Background:
    Given a Git repo with origin
    And the current branch is "main"
    And local Git setting "git-town.create-worktree" is "true"

  Scenario: enabled via config without a flag
    When I run "git-town hack feature"
    Then Git Town runs the commands
      | BRANCH | COMMAND                                                       |
      | main   | git fetch --prune --tags                                      |
      |        | git worktree add -b feature {{ worktree-path "feature" }} main |
    And the current branch is still "main"
    And this lineage exists now
      """
      main
        feature
      """

  Scenario: --no-worktree overrides the config setting
    When I run "git-town hack --no-worktree feature"
    Then Git Town runs the commands
      | BRANCH | COMMAND                  |
      | main   | git fetch --prune --tags |
      |        | git checkout -b feature  |
    And the current branch is "feature"
