# TDD Development Prompt

## Task
Implement user authentication system using Test-Driven Development.

## Process (TDD Cycle)
Follow this cycle strictly:

### RED
1. Write a failing test for the next requirement
2. Run tests - confirm the new test fails
3. If test passes unexpectedly, requirement may already be met

### GREEN
4. Write the minimum code to make the test pass
5. Run tests
6. If tests fail, debug and fix
7. Repeat until test passes

### REFACTOR
8. Clean up code while keeping tests green
9. Run tests after each refactoring step
10. If tests break, undo and try again

## Requirements (implement in order)
1. User registration with email/password
2. Password hashing (bcrypt)
3. User login returning JWT token
4. Token validation middleware
5. Protected route that requires valid token
6. Token refresh endpoint
7. User logout (token invalidation)

## Commands to run
- Run tests: `npm test` or `pytest` (depending on stack)
- Run specific test: adjust as needed

## Iteration Protocol
Each iteration:
1. Identify next unimplemented requirement
2. Write test for it
3. Make test pass
4. Refactor if needed
5. Move to next requirement

## Completion Signal
When ALL requirements have passing tests, output:

**ALL_TESTS_PASS**

## Stuck Protocol
If stuck on same test for 5+ iterations:
- Skip to next requirement
- Document the blocker
- Return to it later
