---
name: review-apollo
description: Multi-agent parallel code review for Apollo (Go TUI commit triage). Spawns 5 specialist lens agents (Go Idioms, Bubbletea/TUI, Architecture, Database, Testing) to review git diff changes, then merges findings into a deduplicated, conflict-aware action plan.
allowed-tools: Task, Read, Write, Edit, Bash, Glob, Grep
---

# Apollo — Multi-Agent Code Review

Orchestrates 5 parallel specialist review agents using the Scatter-Gather pattern. Each agent reviews changes through a specific lens, then findings are merged into a single deduplicated action plan.

## COMMAND SYNTAX

```
/review-apollo [--base=BRANCH_OR_COMMIT] [--files=GLOB] [--report-only]
```

**Parameters:**
- `--base=REF`: Base for diff (default: auto-detect — `HEAD` for unstaged, `master` for branch diff)
- `--files=GLOB`: Limit review to specific files (e.g., `internal/tui/**`)
- `--report-only`: Generate report without fix suggestions (findings only)

---

## EXECUTION PROCESS

### Phase 0: Input Collection

**Step 0.1 — Detect git state:**

```bash
git rev-parse --is-inside-work-tree
git branch --show-current
git merge-base HEAD master 2>/dev/null || echo "no-merge-base"
```

**Step 0.2 — Determine diff base:**

If `--base` provided, use it. Otherwise:
- If there are unstaged/staged changes → `BASE=HEAD`
- If on a feature branch with no local changes → `BASE=$(git merge-base HEAD master)`
- Fallback → `BASE=master`

**Step 0.3 — Collect diff:**

```bash
# Full diff content (for agents)
git diff $BASE -- $FILES_FILTER

# File list with status
git diff $BASE --name-status -- $FILES_FILTER

# Stats summary
git diff $BASE --stat -- $FILES_FILTER
```

If `--files` provided, append the glob to each command.

**Step 0.4 — Classify changed files:**

From `--name-status`, build a classification map:
```
Config:      internal/config/**
DB:          internal/db/**
Git:         internal/git/**
Watcher:     internal/watcher/**
Notifier:    internal/notifier/**
TUI:         internal/tui/**
Style:       internal/style/**
Entry:       main.go
Tests:       *_test.go
Other:       everything else
```

**Step 0.5 — Validate diff size:**

Count total diff lines. If > 3000 lines, log a warning:
```
WARNING: Large diff ({N} lines). Review may take longer. Consider using --files to scope.
```

**Step 0.6 — Output session header:**
```
REVIEW SESSION
Branch: {BRANCH} | Base: {BASE} | Mode: {report-only | full}
Changes: {N} files ({N} core, {N} tui, {N} test, {N} other) | +{N} -{N} lines
```

---

### Phase 1: Context Loading

Read the following project docs to include as context for all agents:

1. `APOLLO_CONCEPT_BRIEF.md` (project root)
2. `CLAUDE.md` (project root, if it exists)

Read all 5 lens files from `.claude/skills/review-apollo/lenses/`:
- `go-idioms.md`
- `bubbletea-tui.md`
- `architecture.md`
- `database.md`
- `testing.md`

---

### Phase 2: Fan-Out (5 Parallel Agents)

Spawn ALL 5 agents in a SINGLE message using the Task tool. Each agent MUST be launched with:
- `subagent_type: "general-purpose"`
- `model: "sonnet"`

**CRITICAL**: All 5 Task calls MUST be in the SAME message to ensure parallel execution.

Each agent receives a prompt with this structure:

```
You are a specialist code reviewer for Apollo, a Go TUI application for commit review triage built with Bubbletea.
Your review lens: {LENS_NAME}

## Project Context
{Contents of APOLLO_CONCEPT_BRIEF.md — abbreviated to key sections for this lens}
{Contents of CLAUDE.md if it exists — abbreviated to key rules}

## Your Lens Rules
{Contents of the specific lenses/*.md file}

## Changes to Review
{Full git diff output}

## Changed Files
{File list with status from --name-status}

## Instructions

1. Read the diff carefully. For each changed file, use the Read tool to read the FULL current file to understand context around the changes.
2. Apply ONLY your lens criteria — do not review outside your specialty.
3. For each finding, output in this EXACT format:

### FINDING
- **severity**: blocking | important | nit
- **file**: exact/path/to/file.go
- **lines**: start-end (line range in current file, not diff)
- **title**: Short description (max 10 words)
- **problem**: What's wrong and why it matters
- **fix**: Concrete code change or approach to fix
- **lens**: {LENS_NAME}

4. If no issues found for your lens, output:
### NO FINDINGS
Lens: {LENS_NAME} — No issues detected.

5. Do NOT suggest improvements outside your lens scope.
6. Do NOT repeat the project rules back — only report violations.
7. Be specific — include exact file paths, line numbers, and code snippets.
8. For blocking: must be a real bug, data loss risk, or architectural violation that breaks correctness.
9. For important: significant quality issue but won't break production.
10. For nit: style preference, minor improvement, nice-to-have.
```

**Agent-specific context mapping:**

| Lens | Context sections to include |
|------|------------------------|
| Go Idioms & Style | CLAUDE.md code style, concept brief system components |
| Bubbletea & TUI | Concept brief: TUI section, UX Notes |
| Architecture | Concept brief: full system components, domain model |
| Database | Concept brief: domain model, storage section |
| Testing | Concept brief: TDD constraints, domain model (for test expectations) |

---

### Phase 3: Fan-In (Merge & Deduplicate)

After ALL 5 agents complete, collect their findings and process:

**Step 3.1 — Parse findings:**

Extract all `### FINDING` blocks from each agent's response. Parse into structured objects:
```
{ severity, file, lines, title, problem, fix, lens }
```

**Step 3.2 — Group by location:**

Group findings by `(file, overlapping_line_range)`. Two findings overlap if their line ranges intersect or are within 3 lines of each other.

**Step 3.3 — Deduplicate:**

For each group of findings at the same location:
- **Same suggestion** (semantically equivalent fix): Merge into ONE finding, list all contributing lenses
- **Different suggestions** (conflicting fixes): Keep BOTH, mark as CONFLICT

**Step 3.4 — Assign IDs:**

- Regular findings: `R-001`, `R-002`, ... (sorted by severity then file path then line)
- Conflicts: `C-001`, `C-002`, ...

**Step 3.5 — Sort:**

1. Severity: blocking > important > nit
2. File path (alphabetical)
3. Line number (ascending)

**Step 3.6 — Compute stats:**

```
- Total findings
- By severity: blocking N, important N, nit N
- By lens: { lens_name: { blocking: N, important: N, nit: N } }
- Conflicts: N
- Files with most findings
```

---

### Phase 4: Output

**Step 4.1 — Generate timestamp and output path:**

```bash
date '+%Y-%m-%d-%H-%M-%S'
```

Output file: `~/.apollo/reviews/review-{TIMESTAMP}.md`

**Step 4.2 — Create output directory:**

```bash
mkdir -p ~/.apollo/reviews
```

**Step 4.3 — Write report using template:**

Use the Write tool to create the report file with this template:

```markdown
# Code Review — Apollo
Generated: {TIMESTAMP} | Branch: {BRANCH} | Base: {BASE}

## Summary
| Metric | Value |
|--------|-------|
| Files reviewed | {N} |
| Blocking | {N} |
| Important | {N} |
| Nit | {N} |
| Conflicts | {N} |

## Lenses
{LENS_STATUS — pass if lens returned findings or explicit NO FINDINGS, fail if agent failed}

---

## Blocking (must fix)

### R-{ID}: {title}
**Lenses:** {lens1, lens2, ...}
**File:** `{file}:{lines}`
**Problem:** {problem}
**Fix:** {fix}

---

## Important (should fix)

{Same format as blocking}

---

## Nit (nice to have)

{Same format as blocking}

---

## Conflicts (needs human decision)

### C-{ID}: `{file}:{lines}`
- **{Lens A}** suggests: {fix_a}
- **{Lens B}** suggests: {fix_b}
- **Why they conflict:** {explanation}

---

## Stats by Lens
| Lens | Blocking | Important | Nit |
|------|----------|-----------|-----|
| Go Idioms & Style | {N} | {N} | {N} |
| Bubbletea & TUI | {N} | {N} | {N} |
| Architecture | {N} | {N} | {N} |
| Database | {N} | {N} | {N} |
| Testing | {N} | {N} | {N} |
```

If a severity section has no findings, write: `_No {severity} findings._`

**Step 4.4 — Output final summary to user:**

```
REVIEW COMPLETE
Report: ~/.apollo/reviews/review-{TIMESTAMP}.md
Findings: {N} blocking | {N} important | {N} nit | {N} conflicts
Top files: {top 3 files by finding count}
```

---

## ERROR HANDLING

- If an agent fails or times out: log as `{LENS_NAME}: agent failed — {error}` in the lens status, continue with other results
- If git diff fails: abort with clear error message
- If no changes detected: output "No changes to review" and exit
- If all agents fail: output error report explaining what happened

## RULES

- NEVER modify any source files — this is read-only analysis
- NEVER commit or stage changes
- Review ONLY files in the diff — don't review unchanged code
- Each agent reads full files for context but only reports on changed lines
- Be specific: exact file paths, line numbers, code snippets
- Findings must be actionable — no vague "consider improving"
