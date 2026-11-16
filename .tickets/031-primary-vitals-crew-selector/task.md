# 031 — Add Crew Member Selector to Primary Vitals Dashboard

| Field | Value |
| :--- | :--- |
| **ID** | `031-primary-vitals-crew-selector` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

`dashboard-secundario.vue` (`OTHER STATS`) already includes a `<select>` dropdown in its `<header>` to switch between active crew members (`crewPool`) without returning to `HomeView.vue`. `dashboard-principal.vue` (`PRIMARY VITALS`) lacked this `<select>`, forcing the user to return to `Home` to view another astronaut's primary vitals.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/031-primary-vitals-crew-selector/`.
- Preserve all existing features in `dashboard-principal.vue` (`Other Stats` button, `?` Medical Manual links, alert colors/blinking, anatomical silhouette alerts, disconnected sensor / no-signal chart state, and `rem` scaling).
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Add `crewPool` computed (`activeAstronauts.value.length > 0 ? activeAstronauts.value : astronauts.value`) and reuse it inside `activeAstronaut`.
   - Add `onAstronautSelect(event: Event)` matching `dashboard-secundario.vue`:
     - Read the selected astronaut `id` from `HTMLSelectElement`.
     - Look up the astronaut in `crewPool.value`, call `setSelectedAstronaut(nextId, found?.fullName)`, and call `void router.replace({ name: "dashboard-principal", query: { ...route.query, astronaut: nextId } })`.
   - In `<header>` right after `<h1>PRIMARY VITALS</h1>`, render the crew `<select id="primary-crew-select">` (`v-if="crewPool.length > 0"`, `:value="resolvedAstronautId"`, `aria-label="Select crew member"`, `class="field h-10 px-3 text-xs sm:text-sm min-w-40 cursor-pointer"`, `@change="onAstronautSelect"`) with `<option v-for="astro in crewPool" :key="astro.id" :value="astro.id" class="bg-ground text-main">{{ astro.fullName }}</option>`, identical in style and behavior to `dashboard-secundario.vue`.
2. Run `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check`, `bun run --cwd /home/sacred/bd/HRIG/app/frontend build`, and `rg` checks, mark `task.md` `completed`, and write `changelog.md`.
