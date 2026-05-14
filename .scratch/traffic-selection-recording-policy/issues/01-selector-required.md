# Issue 01: Require Ruleset Selector

Status: completed
Blocked by: none

## Scope

- Reject rulesets whose `selector.all` is empty.
- Update tests and fixtures that relied on empty selector behavior.
- Keep protocol selector field/operator validation unchanged.

## Acceptance

- Empty selector fails validation with a clear `selector` issue.
- Publish/compile cannot accept empty selectors.
- Tests cover the new validation behavior.

## Result

- `engine.ValidateRuleSet` rejects empty `selector.all`.
- Ruleset settings UI now requires at least one selector condition before save.
- Engine validation tests cover the empty selector failure.
