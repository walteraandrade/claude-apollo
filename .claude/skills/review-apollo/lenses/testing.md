# Testing Lens

## Focus Areas

### Test Existence
- Every new/modified package should have corresponding `_test.go` file
- Tests live next to source: `model_test.go` beside `model.go`
- New exported function without test = important (unless trivial getter)
- New package without any test file = blocking

### Review State Transition Tests
Review state is core domain — tests must verify:
- Default status on insert = `unreviewed`
- Transition to `reviewed` sets `reviewed_at` to non-nil
- Transition to `ignored` sets `reviewed_at` to nil
- Transition back to `unreviewed` clears `reviewed_at`
- `InsertCommit` creates both `commits` and `review_state` rows
- Duplicate commit insert handled gracefully (INSERT OR IGNORE)
- Missing status transition test = important

### Watcher Debounce Tests
- Rapid successive events produce single debounced output
- Event after debounce period produces new output
- Cleanup function stops goroutine and closes channel
- Watcher ignores non-Write/Create events
- Missing debounce behavior test = important

### TUI Partition & Navigation Tests
- `partitionCommits` distributes by status to correct columns
- `ClampCursor` prevents out-of-bounds after partition
- Column navigation wraps correctly at boundaries (doesn't go negative or past NumColumns-1)
- `Selected()` returns nil for empty column
- Expanded hash clears on navigation
- Missing partition/navigation test = important

### Git Temp Repo Tests
- Tests that exercise git operations should use `t.TempDir()` with `git init`
- Create real commits in temp repo, then verify `ReadNewCommits` output
- Don't mock git when testing git package — use real temp repos
- Test edge cases: empty repo, single commit, branch detection

### Table-Driven Tests
- Prefer table-driven pattern for functions with multiple input/output combinations
- Each test case: name, input, expected output
- `t.Run(name, func(t *testing.T) { ... })` for subtests
- Not every test needs table-driven — simple functions can use direct assertions

### Test Isolation
- Tests must not depend on external state (real filesystem paths, network, real DB)
- Use `t.TempDir()` for file-based tests
- Use in-memory SQLite (`:memory:`) or temp file for DB tests
- No shared mutable state between test functions
- Tests depending on real git repo or network = important

### Error Path Testing
- Test both success and error paths
- Verify error messages contain expected context
- Test edge cases: empty input, nil values, boundary conditions
- Functions that return errors: test at least one error case

### Mocking
- Use interfaces for external dependencies (`Notifier`)
- Mock implementations in test files
- Don't mock what you own unnecessarily — prefer real implementations for internal packages
- Test real DB operations with in-memory SQLite, not mocks

### Test Naming
- Test function: `TestFunctionName` or `TestType_Method`
- Subtest names: descriptive of the scenario, snake_case OK
- Avoid test names that just repeat the function name

### Assertions
- Use standard `testing` package — `t.Errorf`, `t.Fatalf`
- Fail fast on setup errors: `t.Fatal` not `t.Error`
- Compare with `==` for simple types, `reflect.DeepEqual` or custom for structs
- Check error existence with `err != nil` / `err == nil`, not string matching

### Build Tags & Running
- Default `go test ./...` must pass without special setup
- Integration tests that need real resources: use build tags
- Race detector: `go test -race ./...` should pass

### Regression Tests
- Bug fixes should include a test that would have caught the bug
- Edge cases discovered during review deserve tests

## Severity Guide

| Issue | Severity |
|-------|----------|
| New package with no tests | blocking |
| `go test ./...` fails | blocking |
| Test depends on network/real filesystem (non-temp) | important |
| Missing review state transition test | important |
| Missing watcher debounce test | important |
| Missing TUI partition/navigation test | important |
| New exported function without test | important |
| Missing error path test for function that can fail | important |
| Test has shared mutable state | important |
| Could use table-driven pattern | nit |
| Test naming inconsistency | nit |
| Missing edge case test | nit |
