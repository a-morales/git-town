@messyoutput
Feature: beam a commit onto a new branch in a new worktree

  Background:
    Given a Git repo with origin
    And the branches
      | NAME     | TYPE    | PARENT | LOCATIONS |
      | existing | feature | main   | local     |
    And the commits
      | BRANCH   | LOCATION | MESSAGE |
      | existing | local    | beamed  |
    And the current branch is "existing"
    When I run "git-town hack --worktree --beam new" and enter into the dialog:
      | DIALOG          | KEYS        |
      | commits to beam | space enter |

  Scenario: result
    Then Git Town runs the commands
      | BRANCH   | COMMAND                                                                                             |
      | existing | git worktree add -b new {{ worktree-path "new" }} main                                              |
      | new      | git cherry-pick {{ sha-initial 'beamed' }}                                                          |
      | existing | git -c rebase.updateRefs=false rebase --onto {{ sha-initial 'beamed' }}^ {{ sha-initial 'beamed' }} |
    And the current branch is still "existing"
    And this lineage exists now
      """
      main
        existing
        new
      """
