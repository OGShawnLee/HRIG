# Health Metric Prioritization

Minimal metric hierarchy for **HRIG**, organized by urgency so non-physician astronauts can assess their health quickly under **Distance from Earth** communication delays.

## Display Rules
- Show **source + timestamp** for every reading; flag disconnected or overdue sensors as **Stale**.
- Pair color with an **icon + label** (**Nominal**, **Caution**, **Critical**).
- Include **how to measure**, **why it matters**, and **next action** on every metric.

## 1. Primary Metrics (Main HUD — Always Visible)
Indicators of acute life-support or physiological emergencies requiring immediate action:
- **Oxygenation (`SpO2`, Cabin `O2` / `ppCO2`, Resp. Rate)** — Detects hypoxia or `CO2` buildup (*Hostile/Closed Environments*).
- **Pressure (Cabin/Suit Pressure & Blood Pressure)** — Detects decompression risk and fluid-shift/orthostatic drops (*Gravity Fields*).
- **Radiation (Exposure Rate & Cumulative Dose)** — Tracks solar particle events and neutron exposure (*Space Radiation*).
- **Cardiac Rhythm (Heart Rate & HRV)** — Flags extreme exertion, dehydration, or cardiovascular strain (*Gravity Fields*).
- **Temperature (Body Core & Cabin/Suit)** — Detects hyperthermia (cooling failure) or fever/infection (*Hostile/Closed Environments*).

## 2. Secondary Metrics (Trends & Scheduled Check-Ins)
Preventive adaptation and mission-routine indicators:
- **Sleep & Mood** — Daily sleep hours and quick mood check-ins for cognitive/behavioral fatigue (*Isolation and Confinement*).
- **Blood Glucose (`mg/dL`)** — Scheduled/non-invasive metabolic check for EVA exertion and nutrition (*Inadequate Nutrition*).
- **Hydration & Body Mass** — Tracks fluid balance, caloric deficit, and muscle/bone loss (*Gravity Fields*).
- **Blood Panel / Ultrasound (Scheduled)** — Periodic RBC/WBC counts and diagnostic imaging (*Hostile/Closed Environments*, *Gravity Fields*).

## 3. Excluded Metrics (Omitted from Main UI)
Static or specialist-only metrics that add clutter without immediate in-flight actions:
- **Height** — Static preflight baseline (spinal elongation requires no acute action).
- **Thyroid Panel (`TSH`, `T3`, `T4`)** — Slow-changing markers requiring specialist review.

## References
- Challenge Brief & Hazard Notes: [`.routing/INPUT.md`](../.routing/INPUT.md) & [`.routing/reference/`](../.routing/reference/)
- NASA Human Research Program — Hazards: https://www.nasa.gov/hrp/hazards/
- NASA Human Research Program — Risks: https://www.nasa.gov/hrp/risks/
- NASA Human Research Program — Ongoing Research: https://www.nasa.gov/hrp/ongoing-hrp-research/
