# Project Context

## Files

- [`CONTEXT.md`](CONTEXT.md) — Project context, ground rules, and index of `.routing` files.
- [`INPUT.md`](INPUT.md) — NASA hackathon challenge prompt, objectives, hazards overview, and radiation monitoring notes.
- [`DOMAIN.md`](DOMAIN.md) — Ubiquitous Language glossary defining core domain concepts, metric tiers, severity states, and clinical roles.
- [`SPECIFICATION.md`](SPECIFICATION.md) — Refined software functional (`RF01`–`RF16`) and non-functional (`RNF01`–`RNF12`) requirements in Spanish until final delivery.
- [`DESIGN.md`](DESIGN.md) — Simplified UI/UX design guidelines, status encoding, body map rules, and human-factors references.
- [`PROGRESS.md`](PROGRESS.md) — Current task status, completed milestones, and upcoming work.
- [`image.png`](image.png) — Reference visual summarizing NASA's five hazards of human spaceflight.
- `reference/` — Minimal reminder notes on NASA's five hazards of human spaceflight:
  - [`space-radiation.md`](reference/space-radiation.md)
  - [`isolation-and-confinement.md`](reference/isolation-and-confinement.md)
  - [`distance-from-earth.md`](reference/distance-from-earth.md)
  - [`gravity-fields.md`](reference/gravity-fields.md)
  - [`hostile-closed-environments.md`](reference/hostile-closed-environments.md)
- External project docs:
  - [`README.md`](../README.md) — Minimal hackathon project overview and repository navigation.
  - [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md) — Cold scope & feasibility review (`venture-killer`) pending team review.
  - [`METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md) — Minimal health metric tiers (Primary, Secondary, Excluded) and NASA references.
  - [`.claude/skills/`](../.claude/skills/) — Project workflow skills (`router`, `ticket`, `venture-killer`).
  - [`.tickets/`](../.tickets/) — Implementation tickets (`task.md`, `spec.md`, `changelog.md`).
  - [`app/`](../app/) — Wails v3 desktop application (Go backend + Vue 3 / UnoCSS / TypeScript frontend).

**HRIG** is a NASA hackathon project to build health monitoring software for astronauts on long-duration space missions (see [`INPUT.md`](INPUT.md)).

## Ground Rules

### Git & Commits
- **Conventional Commits in English**: All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification and be written strictly in English.
  - Format: `<type>[optional scope]: <description>`
  - Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `style`, `test`, `build`, `ci`
  - Keep subject lines clear and imperative; include a minimal body only when helpful context is needed.
