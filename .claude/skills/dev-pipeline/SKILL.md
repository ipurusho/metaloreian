---
name: dev-pipeline
description: >
  End-to-end AI-native development pipeline: triage GitHub issues, auto-prioritize, branch, develop the feature,
  run security audit, build a local UAT artifact, and open a PR for manual review. Use this skill whenever the user
  says "triage issues", "pick an issue", "work on an issue", "dev pipeline", "issue to PR", "start working",
  or wants to go from a GitHub issue to a ready-for-review pull request. Also use when the user says "what should
  I work on next?" or asks to prioritize the backlog.
---

# Dev Pipeline

You are orchestrating a full development cycle from GitHub issue to pull request. This is an opinionated, multi-phase workflow that keeps the human in the loop at key decision points while automating the mechanical parts.

## Prerequisites

- `gh` CLI must be authenticated (for issue/PR operations via GitHub API)
- The `security-audit` skill must be available (for pre-PR security gate)
- Docker and docker-compose must be installed (for local UAT artifact)

## Phase 1: Issue Triage & Prioritization

Fetch all open issues from the repository and auto-prioritize them.

### Step 1: Fetch issues

Use `gh issue list` to fetch all open issues. Include labels, assignees, creation date, and comments. Use `gh issue view` for full details.

### Step 2: Auto-prioritize

Score each issue on these dimensions (1-5 scale each):

| Dimension | What to assess |
|-----------|---------------|
| **Severity** | Is this a crash/data-loss bug (5), a functional bug (3-4), a cosmetic issue (2), or a feature request (1)? Look at the issue title, body, and any labels like `bug`, `critical`, `enhancement`. |
| **User Impact** | How many users does this affect? Production-blocking (5), affects core flow (3-4), edge case (1-2). Infer from the description and comments. |
| **Complexity** | Estimate implementation effort. Trivial fix (1), moderate feature (2-3), large cross-cutting change (4-5). Consider which files/systems are involved. |
| **Dependencies** | Does this block or get blocked by other issues? Blocking issues score higher. Check for references like "blocked by #X" or "depends on #Y". |
| **Staleness** | Older issues with no activity get a slight bump — they've been neglected. |

**Priority Score** = (Severity × 2) + (User Impact × 2) + (6 - Complexity) + Dependencies + (Staleness × 0.5)

Higher scores = work on first. The complexity dimension is inverted because quick wins with high impact should be prioritized.

### Step 3: Present to the user

Show a ranked table:

```
#  | Issue | Priority | Severity | Impact | Complexity | Summary
---|-------|----------|----------|--------|------------|--------
1  | #42   | 18.5     | HIGH     | HIGH   | LOW        | Fix CORS...
2  | #37   | 15.0     | MED      | HIGH   | MED        | Add album...
```

Then ask: **"Which issue do you want to work on? (Enter issue number, or 'top' for the highest priority)"**

Wait for the user's choice before proceeding.

## Phase 2: Branch & Plan

### Step 1: Read the issue

Fetch the full issue details (body, comments, linked PRs, labels). Understand exactly what needs to be done.

### Step 2: Create the branch

Derive a branch name from the issue:
- Bug: `fix/<issue-number>-<short-description>`
- Feature: `feat/<issue-number>-<short-description>`
- Refactor: `refactor/<issue-number>-<short-description>`
- Docs: `docs/<issue-number>-<short-description>`

```bash
git checkout main
git pull origin main
git checkout -b <branch-name>
```

### Step 3: Draft an implementation plan

Based on the issue, outline:
1. Which files need to change
2. What the changes are (brief description per file)
3. Any new files needed
4. Potential risks or edge cases

Present the plan to the user: **"Here's my plan. Want me to proceed, or adjust anything?"**

Wait for approval before coding.

## Phase 3: Develop

Implement the changes according to the plan. Follow these principles:

- **Read before writing.** Always read existing code before modifying it.
- **Minimal changes.** Only touch what's needed for the issue. No drive-by refactors.
- **Follow existing patterns.** Match the code style, naming conventions, and architecture already in the codebase.
- **Commit atomically.** Make small, logical commits as you go — not one giant commit at the end. Each commit should have a clear message referencing the issue number.
- **Security audit after every commit.** After each commit, run the `security-audit` skill against the full codebase. This catches vulnerabilities as they're introduced rather than batching them at the end. If the audit finds CRITICAL or HIGH findings introduced by the commit, fix them immediately in a follow-up commit before continuing development. Medium and below can be noted and addressed later.

After implementation is complete, stage and commit all changes.

## Phase 4: Final Security Audit

After all development commits are done, run one final comprehensive `security-audit` as a pre-PR gate. This catches any issues that may have slipped through the per-commit audits or that only emerge from the interaction of multiple changes.

Invoke the security audit — it will check dependencies, static analysis, infrastructure, auth, API security, and secrets.

**Gate check:** If the audit returns any CRITICAL or HIGH findings:
1. Show the findings to the user
2. Ask: **"The security audit found issues that should be fixed before opening the PR. Want me to fix them now?"**
3. If yes, fix them, re-commit, and re-run the audit
4. If no, note them in the PR description as known issues

If the audit passes (no CRITICAL/HIGH), proceed to the next phase.

## Phase 5: Local UAT Artifact

Build and start the full application locally so the user can manually test.

```bash
# Build the Docker image locally
docker compose build

# Start the full stack
docker compose up -d
```

Verify the services are healthy:
```bash
# Check all containers are running
docker compose ps

# Hit the health endpoint
curl -s http://localhost:8080/health
```

Tell the user: **"The app is running locally. You can test at http://localhost:5173 (frontend) or http://localhost:8080 (API). Let me know when you're done testing and ready to open the PR, or if you find issues I should fix."**

Wait for the user's go-ahead before proceeding.

## Phase 6: Squash & Open PR

Every PR ships exactly one commit. This keeps the main branch history clean and makes reverts trivial. The atomic commits during development are useful for per-commit security audits and incremental progress, but they get squashed before the PR is opened.

### Step 1: Squash all commits into one

After UAT passes, squash all commits on the branch into a single commit. Use a soft reset to main and re-commit:

```bash
# Count commits on the branch
git log --oneline main..HEAD

# Soft reset to main (keeps all changes staged)
git reset --soft main

# Create the single squashed commit with a comprehensive message
git commit -m "<message>"
```

The squashed commit message should:
- Start with a concise title (under 70 chars)
- Include a body that summarizes all the work done (features, security fixes, test additions, etc.)
- Reference the issue number with `Closes #<issue-number>`
- End with `Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>`

### Step 2: Push the branch

```bash
git push -u origin <branch-name>
```

### Step 3: Create the PR

Use `gh pr create` to create a pull request with this structure:

**Title:** Short, descriptive (under 70 chars), referencing the issue

**Body:**
```markdown
## Summary
Closes #<issue-number>

<2-3 bullet points describing what changed and why>

## Changes
- `file1.go`: <what changed>
- `file2.tsx`: <what changed>

## Security Audit
<PASS/FAIL with summary of findings>
- Critical: 0
- High: 0
- Medium: [count]
- Low: [count]

[If any medium/low findings, list them briefly]

## Testing
- [ ] Local UAT completed by developer
- [ ] [specific test scenarios relevant to the change]

## Screenshots
[If UI changes, note that screenshots should be added]
```

### Step 4: Confirm

Tell the user: **"PR is open at [URL]. It's ready for manual review."**

## Error Handling

- If `gh` CLI is not authenticated, prompt the user to run `gh auth login`.
- If Docker build fails, show the error and ask the user how to proceed.
- If the security audit can't install a tool (govulncheck, gosec), skip that check and note it in the report.
- If `docker compose up` fails, try `make docker` as fallback, then `make dev` for a non-containerized local run.

## Workflow Diagram

```
Fetch Issues → Auto-Prioritize → [USER PICKS] → Create Branch → Plan → [USER APPROVES]
  → { Develop → Commit → Security Audit } (repeat per commit)
  → Final Security Audit → [GATE CHECK] → Local UAT → [USER TESTS]
  → Squash → Push & Open PR → [READY FOR REVIEW]
```

Human decision points are marked with brackets. Never skip past them without user input.
