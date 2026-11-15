# Spec — 025 Convert Dashboard `px` Tracks & Micro-Text to `rem`

## Functional Requirements

1. Every grid track width and typography size in `dashboard-principal.vue` scales uniformly with root `font-size` (`rem`), preventing truncation (`9...`, `1...`, `0....`, `3...`) and preventing the astronaut silhouette from overflowing its `<aside>` card at any viewport scale (`0.7x`–`2.25x`).
2. `<strong>{{ row.value }}</strong>` renders full metric values (`98`, `112/74`, `0.245`, `36.9`) without ellipsis truncation.
3. `CrewCard.vue` portrait badge micro positioning (`top-[1.8125rem]`, `text-[0.625rem]`) scales proportionally with `rem`.

## Acceptance Criteria

- `rg -n "grid-cols-.*px|text-\[[0-9]+px\]" app/frontend/src/views/dashboard-principal.vue` returns zero matches.
- `rg -n "11\.5rem|13rem" app/frontend/src/views/dashboard-principal.vue` matches both scaled grid tracks.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
