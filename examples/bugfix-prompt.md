# Bug Fixing Prompt

## Bug Description
[Describe the bug here]

Example: Users report that the shopping cart total shows incorrect values
when items with discounts are added.

## Steps to Reproduce
1. Add item with 20% discount to cart
2. Add regular price item
3. View cart total
4. Expected: Correct discounted total
5. Actual: Discount not applied

## Investigation Steps
1. Find relevant code (cart calculation, discount logic)
2. Add logging/debugging to trace the issue
3. Write a failing test that reproduces the bug
4. Identify root cause

## Fix Protocol
1. Understand the root cause
2. Implement the fix
3. Run the reproduction test - should pass
4. Run full test suite - should all pass
5. Manual verification if possible

## Regression Prevention
- Write test case that would catch this bug
- Consider edge cases:
  - Multiple discounts
  - Zero-value discounts
  - Negative values
  - Currency rounding

## Verification Checklist
- [ ] Bug is reproducible before fix
- [ ] Root cause identified
- [ ] Fix implemented
- [ ] Regression test written
- [ ] All existing tests pass
- [ ] No new issues introduced

## Completion Signal
When bug is fixed and verified, output:

**FIXED**

## Escalation
After 15 iterations without fix:
- Document all findings
- List attempted solutions
- Identify what additional information is needed
- Output: **NEEDS_HUMAN_REVIEW**
