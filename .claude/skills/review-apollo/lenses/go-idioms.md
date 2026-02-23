# Go Idioms & Style Lens

## Focus Areas

### Error Handling
- Errors MUST be returned, never silently discarded
- `if err != nil { return ..., err }` — standard pattern
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)` — never bare `return err` from public functions
- Never use `panic` for recoverable errors
- `_ = someFunc()` that returns error = blocking if the error matters (DB writes, file ops)
- OK for fire-and-forget (logging, optional cleanup)

### Resource Cleanup
- `defer` for all cleanup: file handles, DB rows, HTTP response bodies, watcher stop functions
- Watcher returns `(<-chan Event, func(), error)` — stop func MUST be called on shutdown
- Channel close must happen in the goroutine that owns it (watcher goroutine closes `ch`)
- Missing cleanup = resource leak = blocking

### Naming
- Variables/functions: `camelCase`
- Exported: `PascalCase`
- Interfaces: single-method → `-er` suffix (`Notifier`, `Reader`)
- Receivers: short (1-2 chars), consistent within type (`m` for Model, `c` for Config)
- Acronyms: `ID` not `Id`, `URL` not `Url`, `API` not `Api`
- Package names: lowercase, single word, no underscores

### Function Length
- Functions >40 lines: consider splitting
- Long `switch` statements are OK if each case is short
- Long `View()` methods in Bubbletea are acceptable (rendering logic is inherently sequential)

### Zero Values & Initialization
- Prefer zero values over explicit initialization when semantically correct
- Use `var` for zero-value declarations, `:=` for initialized
- Struct literals: use field names, not positional args

### Interface Design
- Accept interfaces, return concrete types
- Small interfaces (1-3 methods) over large ones
- `Notifier` interface has one method (`Notify`) — keep it clean
- Don't define interfaces preemptively — define at point of use or when polymorphism exists

### Goroutines & Concurrency
- Goroutines must have clear ownership and shutdown path
- Watcher goroutine: owned by `Watch()`, shutdown via `done` channel + cleanup func
- Channel buffer sizing: intentional — watcher uses `make(chan Event, 1)` to avoid blocking sender
- `select { case ch <- ev: default: }` pattern drops events if consumer is slow — intentional for debounce
- Context propagation: pass `context.Context` as first param for cancellable work

### Imports
- Group: stdlib, blank line, external, blank line, internal
- No unused imports
- Avoid dot imports except in test files (and even then, prefer not)

### Dead Code
- Unused variables, functions, types
- Commented-out code blocks
- Unreachable code after return
- `fmt.Println` debug statements left in

### String Formatting
- `fmt.Sprintf` for complex formatting
- String concatenation (`+`) only for 2-3 simple parts
- `strings.Builder` for loops building strings

### Notifier Error Handling
- `Desktop.Notify` wraps `notify-send` lookup failure — good
- `Fallback.Notify` returns nil — intentional silent fallback
- `New()` selects implementation at startup — caller doesn't need to handle both
- If notifier errors surface to TUI, they must not crash the app

## Severity Guide

| Issue | Severity |
|-------|----------|
| Silently discarded error from DB/file/network op | blocking |
| panic for recoverable error | blocking |
| Goroutine leak (no shutdown path) | blocking |
| Missing watcher/resource cleanup on shutdown | blocking |
| Missing error context wrapping on public func | important |
| Function >40 lines with mixed responsibilities | important |
| Dead code / debug prints | important |
| Channel buffer sizing without clear rationale | important |
| Inconsistent receiver names | nit |
| Suboptimal string building in non-hot path | nit |
| Minor naming convention deviation | nit |
| Import grouping | nit |
