# Project Scope & Feasibility Review (`venture-killer`)

Audit of [`.routing/SPECIFICATION.md`](.routing/SPECIFICATION.md) and [`.routing/DOMAIN.md`](.routing/DOMAIN.md) against [`.routing/INPUT.md`](.routing/INPUT.md), [`docs/METRICS_PRIORIZATION.md`](docs/METRICS_PRIORIZATION.md), and hackathon feasibility constraints.

## 1. Assumption Autopsy (Claim vs. Reality)

- **Naive assumption:** A hackathon prototype needs full hospital EHR infrastructure—verified external system authentication (`RF09`), consent-permission toggles (`RNF12`), immutable clinical ledgers (`RNF11`), three separate device UI layouts (`RNF07`), full crew-profile CRUD (`RF14`), four environment rule engines (`RF15`), background routine schedulers (`RF16`), and lab assays like cortisol and RBC/WBC counts ([`.routing/DOMAIN.md`](.routing/DOMAIN.md)).
- **Operational / competition reality:** The app must run **100% offline** (`RNF09`), demonstrate the **Collect → Evaluate → Act** loop to judges in under 2 minutes, and stay usable by non-physician astronauts. Auth servers, multi-device layouts, and lab hematology workflows burn build time without improving the core demo.

## 2. Fatal Failure Points (What Judges Will Challenge)

- **Auth & External System Dead-End (`RF09`, `RNF12`):** Claiming "verified, authenticated, authorized external systems" in an offline hackathon app invites scrutiny or wastes time building fake login screens.
- **Unverified Numeric Claims (`RNF02`, `RNF08`):** `6:1`/`10:1` contrast and `< 2s` alert latency have no cited source in [`.routing/INPUT.md`](.routing/INPUT.md) or [`docs/METRICS_PRIORIZATION.md`](docs/METRICS_PRIORIZATION.md) (`[UNVERIFIED]`), and verified clinical alert thresholds for vitals are not yet locked in.
- **Simulated vs. Lab Metrics Bloat ([`.routing/DOMAIN.md`](.routing/DOMAIN.md)):** Listing cortisol, caloric intake, and RBC/WBC blood counts alongside simulated wearables blurs what an astronaut can actually measure and act on live in a 2-minute demo.
- **Navigation vs. HUD Clutter:** `RF01`–`RF16` spread across multiple views will create deep menus unless unified by the persistent **Command Palette (`Ctrl+K` / Dynamic Island)** and strict Tier 1 / Tier 2 separation.

## 3. Verdict & Scope Cut

### Cut / Out of Scope
- **Drop `RF09` & `RNF12` (Auth / Verified External Systems / Consent Engine):** Replace with a simple local **Downlink Summary Export (JSON/View)**—no login, auth, or permission plumbing.
- **Shrink `RF14` (Full Profile CRUD):** Keep **2–3 preloaded crew presets + quick switcher** instead of profile creation/editing forms.
- **Shrink `RF15` (4 Environments):** Keep **IVA (Cabin/Base) vs. EVA (Suited)** mode toggle (plus scenario presets such as Nominal, Solar Storm, EVA Hypoxia) instead of 4 separate clinical threshold engines.
- **Cut `RF16` (Background Scheduler) or fold into a lightweight Check-In badge:** No cron/notification engine; show overdue/pending check-ins directly on Tier 2 cards.
- **Trim `RNF07` (3 Device Modes):** Target **1 clean responsive dark-mode HUD** (Tablet/Desktop) + **1 compact EVA view**, not 3 bespoke UIs.
- **Trim [`.routing/DOMAIN.md`](.routing/DOMAIN.md) Secondary Metrics & Terms:** Drop cortisol and RBC/WBC lab counts from MVP scope (keep **Sleep/Mood**, **Blood Glucose `mg/dL`**, **Hydration/Mass**, and optional **Image/Ultrasound Spot Check**); simplify *Sistemas Autorizados*, *Métricas Excluidas*, and immutability (`RNF11`) wording.

### Pragmatic MVP Core
1. **Tier 1 Live HUD + Scenario Simulator (`RF01`, `RF08`):** Primary vitals & cabin/suit metrics with redundant severity encoding (**Nominal / Precaución / Crítico / Stale**) + crew-wide alert banner.
2. **Interactive Anatomical Figure (`RF11`–`RF13`):** Clickable body regions/systems showing status, symptoms, NASA hazard link, and plain-language procedures.
3. **Fast Capture & Guided Evaluation (`RF02`–`RF07`):** Simulated sensor streams + quick manual check-in (`mg/dL`, sleep, mood, spot image) with inline *"How to measure"*, *"Why it matters (5 Hazards)"*, and *"What to do"* actions.
4. **Tier 2 Trends & Local History (`RF10`, `RF02`):** Simple trend charts + timestamped crew log (`MET` / `UTC`).
5. **Persistent Command Palette / Dynamic Island (`Ctrl+K` + Touch):** Always-visible pill for jumping straight to any crewmember, metric, body region, procedure, or hazard without deep menus.
