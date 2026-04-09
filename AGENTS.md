# Learning Project — OpenCode Instructions

## Anti-AI Pattern Enforcement (MANDATORY)

**This project must follow the anti-AI guidelines in `.claude/anti-ai-patterns.md` at the workspace root.**

Before writing ANY commit, code, or documentation:

1. **Read the anti-AI patterns guide** — Run `cat ../../.claude/anti-ai-patterns.md` from this directory (2 levels up)
2. **Commit timing must be realistic** — Space out commits naturally, no rapid-fire
3. **Commit messages must be concise** — "add X", "fix Y", not detailed explanations
4. **Leave development scars** — Show iteration, bugs fixed, refactors
5. **README is a learning journal** — First-person with personality, what broke you
6. **No perfect code from the start** — Real code has warts

**Before every commit, verify:**
- [ ] Conventional commits format
- [ ] Atomic commit (one thing, compiles, tests pass)
- [ ] No orphan variables/imports
- [ ] Specific error handling (not generic `catch Exception`)
- [ ] Comments explain WHY, not WHAT
- [ ] Idiomatic names for the language
- [ ] Timing looks plausible

**GGA enforces this** — If GGA rejects a commit, fix the issues and re-commit. NEVER use `--no-verify`.

---

## Rules

- This is a LEARNING project. Built from scratch to understand internals. No shortcuts.
- No frameworks unless the project IS about building one. Raw implementations first.
- This project must have: a README explaining what was learned, working tests, and a Makefile or equivalent task runner.
- Commit messages use conventional commits format. No AI attribution in commits.
- **Commit after every meaningful unit of work** — a function implemented, a test passing, a refactor done, a module wired up. Do NOT batch all work into one commit. Aim for 3-8 commits per session, like a real developer. Each commit should be atomic: it compiles/runs and tests pass.
- Before writing code, read the relevant language skill from `_shared/skills/`.
- GGA (Gentleman Guardian Angel) is installed as a pre-commit hook. It reviews your code against AGENTS.md on every commit. If GGA rejects a commit, fix the issues and re-commit — do NOT bypass with `--no-verify`.
- Never install a library that implements the thing being built — the forbidden dependencies list below is not optional.
- Tests are mandatory. Write tests BEFORE implementation where possible (TDD).
- When stuck, explain the concept first, then write the code. Understanding > working code.

## Development Workflow — SDD Pipeline (MANDATORY)

Every feature, module, or milestone in this project MUST follow the Spec-Driven Development pipeline. Do NOT jump straight to writing code.

**Pipeline for each milestone or feature:**
1. `/sdd-explore` — Investigate the problem space. Research the concepts, read relevant specs/RFCs, understand the domain before proposing anything.
2. `/sdd-new {feature-name}` — Create a change proposal with intent, scope, and approach.
3. `/sdd-ff {feature-name}` — Fast-forward through spec → design → tasks. This produces the full implementation plan.
4. `/sdd-apply {feature-name}` — Implement the tasks from the plan, writing actual code following the specs and design.
5. `/sdd-verify {feature-name}` — Validate that implementation matches specs, design, and tasks.
6. `/sdd-archive {feature-name}` — Archive the completed change.

**When to use SDD:**
- Every milestone listed in this project's Milestones section = one SDD change
- Any feature that touches more than 2 files
- Any architectural decision

**When SDD is overkill:**
- Single-file bug fixes
- README updates
- Config tweaks

## Skills (auto-load based on context)

| Context | Read this file |
|---------|---------------|
| C or header files | `_shared/skills/c-systems/SKILL.md` |
| Go files | `_shared/skills/go-systems/SKILL.md` |
| Python files | `_shared/skills/python-ml/SKILL.md` |
| TypeScript files | `_shared/skills/typescript-web/SKILL.md` |
| Networking/security work | `_shared/skills/networking/SKILL.md` |

## Code Review Rules

### General
- Code must compile/run without errors before committing.
- Every public function must have a clear purpose — no dead code.

### Code Quality
- No TODO or FIXME comments left in committed code unless tracking a known limitation.
- Functions should do ONE thing. If a function is longer than 40 lines, it likely needs splitting.
- Variable and function names must be descriptive — no single-letter names except loop counters and math formulas.
- No hardcoded magic numbers — use named constants.

### Error Handling
- Every error path must be handled. No ignored return values, no empty catch blocks.
- Error messages must include context (what operation failed, with what input).

### Testing
- New functionality must have corresponding tests.
- Tests must be meaningful — not just asserting true == true.
- Test names must describe the scenario being tested.

### Security
- Never commit secrets, API keys, or credentials.
- Never log sensitive data (passwords, tokens, personal information).
- Validate all external inputs before processing.

### Forbidden
- No libraries that implement the core thing being built (defeats the learning purpose).
- No AI attribution in commit messages.
- No `--no-verify` to bypass this hook.

## Engram Session Protocol (MANDATORY)

**On session start — ALWAYS do this first:**
1. Call `mem_search` with the project name to find prior session context.
2. Call `mem_context` to load recent session history.
3. If prior sessions exist, use `mem_get_observation` to read full content.
4. From engram, determine:
   - Which milestone you're working on (1, 2, 3, or 4)
   - Which SDD phase you're in (explore / propose / spec / design / tasks / apply / verify / archive)
   - What was accomplished last session
   - What the next step is
5. Output a brief status: "Resuming [project] — Milestone N, SDD phase: [phase], next step: [what]".
6. Continue from where you left off — do NOT restart from scratch.

**On session end — ALWAYS do this before saying done:**
Call `mem_session_summary` with:
- **Goal**: What milestone/feature you were working on
- **Accomplished**: What was completed this session
- **Current SDD Phase**: Which phase the current milestone is in
- **Next Steps**: Exact next action for the next session to pick up
- **Relevant Files**: Files created or modified

Also call `mem_save` with topic key `learning-projects/{project-name}/progress` containing:
- `milestone`: current milestone number (1-4)
- `sdd_phase`: current SDD phase
- `sdd_change_name`: the `/sdd-new` change name being worked on
- `files_created`: list of source files created so far
- `tests_passing`: true/false
- `blockers`: any issues or open questions

**After completing a milestone:**
Call `mem_save` with topic key `learning-projects/{project-name}/milestone-N` containing:
- What was built
- Key concepts learned
- Decisions made and why
- Gotchas or surprises discovered

This ensures the next session (or the orchestrator) can always pick up exactly where you left off.
