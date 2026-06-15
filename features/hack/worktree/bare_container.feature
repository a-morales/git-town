@skipWindows
Feature: hack a new branch in a new worktree from a bare repository container

# The developer's project is a bare repository whose worktrees are siblings in a
# container directory. Running "git town hack --worktree" from the container
# itself (which has no working tree) creates the new worktree there, based on
# origin/main.

  Background:
    Given a Git repo with origin
    And the repository is a bare repo container

  Scenario: create a worktree from the bare container
    When I run "git-town hack --worktree feature2" in the bare repo container
    Then Git Town runs the commands
      | BRANCH | COMMAND                                                                            |
      | main   | git fetch --prune --tags                                                           |
      |        | git worktree add -b feature2 --no-track {{ worktree-path "feature2" }} origin/main |

  Scenario: undo
    When I run "git-town hack --worktree feature2" in the bare repo container
    And I run "git-town undo" in the bare repo container
    Then Git Town runs the commands
      | BRANCH | COMMAND                                            |
      | main   | git worktree remove {{ worktree-path "feature2" }} |
      |        | git branch -D feature2                             |

  Scenario: worktree mode is required from a bare container
    When I run "git-town hack feature2" in the bare repo container
    Then Git Town prints the error:
      """
      this is a bare repository with no working tree
      """

  Scenario: --commit is rejected from a bare container
    When I run "git-town hack --worktree --commit -m work feature2" in the bare repo container
    Then Git Town prints the error:
      """
      cannot use --commit from a bare repository
      """

  Scenario: --beam is rejected from a bare container
    When I run "git-town hack --worktree --beam feature2" in the bare repo container
    Then Git Town prints the error:
      """
      cannot use --beam from a bare repository
      """
