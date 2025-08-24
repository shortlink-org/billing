@repo
Feature: Concurrent captures cause version conflict

  Scenario: Two captures on the same payment concurrently
    Given a stored payment "r1" exists in state "CREATED" with amount "USD 10.00"
      And two parallel workers load the same version
    When worker A captures "USD 6.00" and saves
      And worker B captures "USD 5.00" and tries to save
    Then worker B must get "VERSION_CONFLICT"
