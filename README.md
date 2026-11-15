# HRIG — Resource Integration Gear

**Resource Integration Gear (HRIG)** is a NASA hackathon project for building visual, self-assessment health monitoring software for astronauts on long-duration space missions.

Inspired by the **RIG** concept from *Dead Space*—an at-a-glance, body-integrated health and status system—this project adapts that core idea into a pragmatic, calm, clinical software prototype achievable within a 2-day hackathon using current technology.

## Challenge & Objective

Long-duration missions expose crews to NASA's five hazards of human spaceflight (**Space Radiation**, **Isolation and Confinement**, **Distance from Earth**, **Gravity Fields**, and **Hostile/Closed Environments**).

Resource Integration Gear provides an offline-first interface that gathers key health indicators, visualizes status clearly (including an interactive anatomical figure), and gives non-physician astronauts actionable guidance to prevent and respond to health risks.

## Repository Navigation

### Core Project Docs
- [`PROJECT-REVIEW.md`](PROJECT-REVIEW.md) — Scope & feasibility audit (`venture-killer`) with proposed MVP cuts *(pending team review)*.
- [`docs/METRICS_PRIORIZATION.md`](docs/METRICS_PRIORIZATION.md) — Health metric hierarchy (Primary HUD, Secondary Trends, Excluded) and NASA references.
- [`docs/STACK_SETUP.md`](docs/STACK_SETUP.md) — Full-stack installation & development guide (Wails v3 desktop/UI, Go API, PostgreSQL 17+).
- [`db/README.md`](db/README.md) — PostgreSQL schema (`001_schema.sql`), catalog seeds (`002_seed.sql`), and 3-day demo dataset (`003_demo_seed.sql`).

### Planning & Ground Rules (`.routing/`)
- [`.routing/PROGRESS.md`](.routing/PROGRESS.md) — Current tasks, completed milestones, and upcoming work.
- [`.routing/CONTEXT.md`](.routing/CONTEXT.md) — Ground rules (Conventional Commits in English) and file map.
- [`.routing/SPECIFICATION.md`](.routing/SPECIFICATION.md) — Functional (`RF`) and non-functional (`RNF`) requirements *(maintained in Spanish until final delivery)*.
- [`.routing/DOMAIN.md`](.routing/DOMAIN.md) — Ubiquitous Language glossary for metrics, severity states, crew roles, and anatomy.
- [`.routing/DESIGN.md`](.routing/DESIGN.md) — UI/UX guidelines, status color/icon redundancy, and human-factors references.
- [`.routing/BD.md`](.routing/BD.md) — PostgreSQL database design specification (tables, constraints, triggers, seed units/metrics).
- [`.routing/BDML.md`](.routing/BDML.md) — DBML schema representation generated from `.routing/BD.md`.
- [`.routing/INPUT.md`](.routing/INPUT.md) — Original NASA hackathon challenge brief and radiation monitoring notes.
- [`.routing/reference/`](.routing/reference/) — Quick reference notes on each of NASA's five hazards of human spaceflight.

### Agent Workflow Skills (`.claude/skills/`)
- [`router`](.claude/skills/router/SKILL.md) — Project coordination, grounding, and ticket dispatch workflow.
- [`ticket`](.claude/skills/ticket/SKILL.md) — Specification format for `.tickets/` tasks and binary verification checks.
- [`venture-killer`](.claude/skills/venture-killer/SKILL.md) — Scope creep and feasibility audit rules.
