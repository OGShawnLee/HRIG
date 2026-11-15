# Project Progress

## Current Task
- [ ] **Review [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)** *(Pending — Team Review)*
  - Review the `venture-killer` scope & feasibility audit and confirm which features to cut from [`DOMAIN.md`](DOMAIN.md) and [`SPECIFICATION.md`](SPECIFICATION.md).
- [ ] **Working on the UI & UI Flow** *(In Progress)*
  - General UI flow defined; designing the core interactive visualization prototype (primary/secondary dashboards, responsive anatomical figure, alerts, and guidance views).
  - Adding a persistent **Command Palette / Universal Search** ("Dynamic Island" / Omnibox) rendered at all times—accessible via touch/click or `Ctrl+K`—for fast system-wide search and navigation.
- [ ] **Scope Reduction, Domain & Specification Refinement** *(Pending)*
  - Apply approved cuts from [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md) to [`DOMAIN.md`](DOMAIN.md) and [`SPECIFICATION.md`](SPECIFICATION.md) so the glossary and requirements stay lean for hackathon delivery.

## Completed
- [x] Documented challenge brief in [`INPUT.md`](INPUT.md)
- [x] Created and updated [`README.md`](../README.md) with repository navigation
- [x] Added [`CONTEXT.md`](CONTEXT.md) with Conventional Commits (English) ground rule
- [x] Added NASA Five Hazards reminder notes in [`reference/`](reference/)
- [x] Defined minimal health metric prioritization in [`docs/METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md)
- [x] Added simplified UI/UX guidelines in [`DESIGN.md`](DESIGN.md)
- [x] Refined domain glossary ([`DOMAIN.md`](DOMAIN.md)) and initial full specification (`RF01`–`RF16`, `RNF01`–`RNF12` in [`SPECIFICATION.md`](SPECIFICATION.md))
- [x] Outlined general UI flow and persistent **Command Palette / Universal Search** (`Ctrl+K` / Dynamic Island) concept
- [x] Adapted workflow skills (`router`, `ticket`, `venture-killer`) in [`.claude/skills/`](../.claude/skills/)
- [x] Generated cold scope & feasibility audit in [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)
- [x] Completed [`.tickets/001-scaffold-wails3-app/`](../.tickets/001-scaffold-wails3-app/) — bare-bones Wails v3 + Go + Vue 3 + Vite + Bun + UnoCSS scaffold verified via `wails3 task client:check`

## Upcoming
- [ ] Draft & execute `.tickets/002-3d-anatomical-figure/` (interactive 3D low-poly anatomical figure + left floating procedures/diagnosis modal)
- [ ] Trim scope and refine [`DOMAIN.md`](DOMAIN.md) and [`SPECIFICATION.md`](SPECIFICATION.md) based on [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)
- [ ] Health indicator input, medical records, and risk assessment logic
- [ ] Translate [`SPECIFICATION.md`](SPECIFICATION.md) and [`DOMAIN.md`](DOMAIN.md) to English and finalize delivery docs
