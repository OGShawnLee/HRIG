# UI/UX Design Guidelines

Minimal UI/UX rules for **HRIG**, aligned with [`INPUT.md`](INPUT.md), [`SPECIFICATION.md`](SPECIFICATION.md), and [`METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md).

## 1. Core Principles
- **Glanceable First:** Surface overall health and acute anomalies immediately without deep menu navigation.
- **Calm, Clinical, Dark-Mode:** High-contrast, low-glare interface designed for spacecraft cabins—functional, zero decorative clutter.
- **Non-Physician Friendly:** Clear language and guided actions for stressed or fatigued crewmembers.

## 2. Responsive Modes
- **Tablet (Primary):** Touch-first, large tap targets (gloved/turbulence friendly), portrait & landscape.
- **Desktop / Laptop:** Multi-pane layout for trend charts, diagnostic imaging, and medical records.
- **Compact HUD / EVA:** High-contrast view showing **Primary Metrics** and active alerts only—no scrolling.

## 3. Layout & Context
- **Metric Tiers:** Always show **Tier 1 (Primary)** on the main HUD; place **Tier 2 (Secondary)** in trend/check-in tabs; omit **Tier 3 (Excluded)**.
- **Zero Horizontal Scroll:** Keep active measurement guides, input, and evaluation together in one view/drawer.
- **Environment Switcher:** Toggle **Spacecraft / Station / Rover / Lunar-Mars Base** to adapt thresholds and guidance.

## 4. Typography & Readability
- **Contrast:** Minimum `6:1` text contrast (`10:1` preferred for alerts).
- **Numerics:** Use `tabular-nums` (or monospace) with aligned decimals, explicit units (`mg/dL`, `mmHg`, `bpm`, `mSv`, `kPa`), target ranges, and trend deltas—no mental math.

## 5. Status Colors & Redundancy
Never rely on color alone—pair every status with an icon and text label:
- **Nominal** (`Cyan / Teal`, Check) — Within safe range.
- **Advisory** (`Blue`, Info) — Early drift or scheduled check-in reminder.
- **Caution** (`Yellow / Amber`, Triangle) — Approaching threshold; self-care or CMO check.
- **Critical / Warning** (`Red`, Alert Pulse) — Threshold breach; immediate action or crew-wide alert.
- **Stale / Unknown** (`Gray / Dashed`, Disconnected) — Sensor offline or overdue; never show old data as live.

## 6. Alerts & Guidance
- **Separate Priorities:** Only **Caution/Critical** states trigger interrupting alerts; advisories stay passive.
- **Filter Spikes & Confirm Overrides:** Smooth brief wearable noise; require explicit confirmation before silencing critical alarms.
- **Explainable Actions:** Show *why* an alert fired and route to **Self-Care**, **Crew Medical Officer (CMO)**, or **Flight Surgeon Downlink**.
- **Micro-Guides:** Embed *"How to measure"* and *"Why it matters"* inside every metric card.

## 7. Interactive Body Map (`RF11`–`RF13`)
- Clickable anatomical figure reflecting overall and limb/system status.
- Selecting a limb/system displays **symptomatology**, **linked spaceflight hazards**, and **step-by-step procedure instructions**.

## 8. Data, Input & Privacy
- **Multi-Modal Input:** Quick taps/sliders (mood, sleep), numeric entry (glucose, BP), sensor streams (IVA wearables / EVA wired suit bus), and image uploads (ultrasound/exams).
- **Source & Time Tagging:** Show source device and timestamp (MET + UTC) on every reading.
- **Offline-First & Consent-Gated:** All monitoring, storage, and guidance run autonomously offline; non-emergency records stay private unless shared with consent or authorized access (`RF09`).

## References
- **NASA-STD-3001 (Vol. 1 & Vol. 2) — OCHMO Standards Portal:** https://www.nasa.gov/ochmo/human-spaceflight-and-aviation-standards/
- **Human Integration Design Handbook (HIDH, NASA/SP-2010-3407 Rev 1):** https://www.nasa.gov/sites/default/files/atoms/files/human_integration_design_handbook_revision_1.pdf
- **FDA — *Applying Human Factors and Usability Engineering to Medical Devices*:** https://www.fda.gov/regulatory-information/search-fda-guidance-documents/applying-human-factors-and-usability-engineering-medical-devices
- **xEMU Wired Heart Rate Monitor & In-Suit Safety (`ICES-2022-429`):** https://hdl.handle.net/2346/89871
- **14 CFR Part 1212 (NASA Privacy Act Regulations):** https://www.ecfr.gov/current/title-14/chapter-V/part-1212?toc=1
