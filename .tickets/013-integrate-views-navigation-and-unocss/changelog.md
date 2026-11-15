# 013 — Changelog

## Files Created
- `app/frontend/public/crew/sam-sonia.jpg` (restored from `git show seth:app/frontend/public/crew/sam-sonia.jpg`)
- `app/frontend/public/crew/patrick-cuckburn.jpg` (restored from `git show seth:app/frontend/public/crew/patrick-cuckburn.jpg`)
- `app/frontend/src/composables/useTelemetryNavigation.ts`
- `app/frontend/src/components/ViewNavBar.vue`
- `app/frontend/src/components/DemoAlertStrip.vue`
- `app/frontend/src/views/support-alert.vue`
- `.tickets/013-integrate-views-navigation-and-unocss/changelog.md`

## Files Modified
- `app/frontend/src/composables/useTelemetry.ts`
- `app/frontend/src/components/CrewCard.vue`
- `app/frontend/src/App.vue`
- `app/frontend/src/router/index.ts`
- `app/frontend/src/views/dashboard-principal.vue`
- `app/frontend/src/views/warning-alert.vue`

## Untouched Protected Files Verified
- `app/frontend/uno.config.ts`
- `app/frontend/src/views/HomeView.vue`
- `app/frontend/src/components/WindowBar.vue`
- `app/frontend/src/components/EnvironmentCard.vue`
- `app/frontend/src/components/MedicalManualButton.vue`
- `server/`
- `app/*.go` (`app/app.go`, `app/client.go`, `app/client_test.go`, `app/main.go`)

## Verification Results

1. **Type-check (`bun run --cwd app/frontend type-check`)**
   - Exit code: `0`
   - Output: `$ vue-tsc --noEmit` (clean)

2. **Production build (`bun run --cwd app/frontend build`)**
   - Exit code: `0`
   - Output:
     - `1795 modules transformed.`
     - `dist/index.html 0.50 kB`
     - `dist/assets/index-BUaMStGW.css 16.89 kB`
     - `dist/assets/index-DhToN8WC.js 167.14 kB`
     - `built in 1.38s`
   - Confirmed `app/frontend/dist/index.html` and `app/frontend/dist/crew/sam-sonia.jpg`, `app/frontend/dist/crew/patrick-cuckburn.jpg` exist.

3. **Typo check (`rg -n "inmediately" app/frontend/src/`)**
   - Empty (no matches); `"immediately"` and `<!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->` preserved in `app/frontend/src/views/support-alert.vue`.

4. **Transport isolation check (`rg -n "new WebSocket|fetch\(" app/frontend/src/`)**
   - Empty (no direct `new WebSocket` or `fetch(` calls in `app/frontend/src/`).

5. **Protected diff check (`git diff -- app/*.go app/frontend/uno.config.ts app/frontend/src/views/HomeView.vue app/frontend/src/components/WindowBar.vue app/frontend/src/components/EnvironmentCard.vue app/frontend/src/components/MedicalManualButton.vue server/`)**
   - Empty (`0` diff lines).

## Notes
- No git commits or pushes were performed.
- No `.go` files were modified (`go test ./...` was not run per ticket instructions because system `webkitgtk-6.0` is unavailable).
