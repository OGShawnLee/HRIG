---
name: ticket
description: Create a RIG ticket (task.md and spec.md) in .tickets/. Use when a request must be handed to a worker subagent. The ticket must be so clear, grounded, and specific that a worker agent can execute it without guessing or hallucinating.
---

# Ticket — HRIG (RIG)

A ticket is a folder in `.tickets/` with two files: `task.md` (what to do) and `spec.md` (binary pass/fail verification checks). Workers hallucinate or drift when requirements, file paths, or clinical thresholds are left vague. Your job is to write comprehensive tickets that remove every ambiguity before handing out work.

## Before Writing: Ask the User

If anything is unclear, **ask the user before writing the ticket**. Do not guess, and do not invent requirements or clinical thresholds. Ask when:
- The goal, output format, or acceptance criteria are not exact.
- You do not know which files may be created/edited or which must remain untouched.
- A metric threshold, formula, or NASA rule has not been verified in `docs/METRICS_PRIORIZATION.md`, `.routing/INPUT.md`, or `.routing/reference/`.
- The request could be interpreted two ways.
- The scope is too broad for one ticket.
- The request conflicts with [CONTEXT.md](../../../.routing/CONTEXT.md), [DOMAIN.md](../../../.routing/DOMAIN.md), [SPECIFICATION.md](../../../.routing/SPECIFICATION.md), or [DESIGN.md](../../../.routing/DESIGN.md).

Ask short, specific questions and wait for answers. Write the ticket only when every step can be stated without "maybe", "probably", or "etc.".

## Folder and Naming

- Location: `.tickets/<id>-<slug>/`
- `<id>`: next free three-digit number (`001`, `002`, ...). Check `.tickets/` first.
- `<slug>`: lowercase words joined by hyphens naming the single outcome.
- Initial status: `draft`.
- Examples: `.tickets/001-lock-mvp-scope/` and `.tickets/002-scaffold-hrig-app/`.

## `task.md`

Use this exact structure:

1. **Title:** `# <id> — <short outcome>`
2. **Metadata table** with rows: `ID`, `Status`, `Worker` (`self` / `inherit`, or user-specified target), `Role`, `Spec` (relative link to `spec.md`).
3. **Scope**
   - `May create or edit:` exact file paths.
   - `Must not touch:` exact paths or folders (always protect `.routing/INPUT.md` and unrelated modules).
4. **Inputs:** relative links to only the exact files (and sections) the worker must read (e.g., `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `docs/METRICS_PRIORIZATION.md`).
5. **Grounding & Anti-Hallucination Constraints:** list exact verified values/formulas to use, and instruct the worker to **stop and record in `changelog.md`** instead of inventing any missing threshold or standard.
6. **Task:** numbered, single-action steps in execution order.
7. **Verification:** one line pointing to `spec.md`.
8. **Change log:** instruction to write `changelog.md` inside the ticket folder listing files touched, checks run, and any blockers/assumptions avoided.

## `spec.md`

- A comprehensive checklist of binary yes/no checks.
- Each check names an exact file, string/pattern, count, or command with expected exit code/output.
- Prefer runnable checks: file exists, build/test command exits with `0`, grep for forbidden/unverified terms returns empty.
- Include a scope check confirming no files outside `May create or edit` were modified.
- Include a ground-rules check (English for code, `.tickets/`, `docs/`, and commits; Conventional Commits in English with a minimal body if a commit was requested).

## Writing Rules for the Worker's Sake

- One ticket, one outcome. Split large tasks.
- Use exact paths, exact symbol names, and exact numbers. Never write "relevant files" or "appropriate values".
- Write steps as imperative commands: "Create", "Update", "Run". One action per step.
- Explicitly state what **not** to do (e.g., "Do not add extra metrics outside `docs/METRICS_PRIORIZATION.md`").
- Ban vague words: "appropriate", "clean", "nice", "as needed", "etc.", "if possible".
- Include a concrete example whenever a format or rule could be misread.
- Keep the ticket concise enough to read in one pass.

## After Writing

1. Re-read `task.md` and `spec.md` as if you had zero prior conversation context. Fix any gap.
2. Share the ticket folder path and any open question with the user.
3. Wait for user confirmation before dispatching the worker.
