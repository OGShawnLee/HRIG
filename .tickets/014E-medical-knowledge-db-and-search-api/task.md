# 014E — Medical Knowledge DB Schema, Seed Data, and Search API

| Field | Value |
| :--- | :--- |
| **ID** | `014E` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Go & PostgreSQL Backend Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `db/001_schema.sql`
  - `db/004_medical_manual.sql`
  - `db/queries/medical.sql`
  - `db/README.md`
  - `Taskfile.yml`
  - `server/internal/db/` *(only generated files from `wails3 task db:sqlc`)*
  - `shared/model/telemetry.go`
  - `shared/model/telemetry_test.go`
  - `server/internal/query/service.go`
  - `server/internal/query/service_test.go`
  - `server/internal/query/service_integration_test.go`
  - `server/internal/http/router.go`
  - `server/internal/http/router_test.go`
  - `.tickets/014E-medical-knowledge-db-and-search-api/changelog.md`
- **Must not touch:**
  - `.routing/INPUT.md`
  - `app/` *(reserved for Ticket 015E)*
  - `bot/`, `bot-n/`
  - Existing measurement ingestion semantics (`POST /`, `server/internal/ingestion/`, `server/internal/realtime/`)

## Inputs

- [`.routing/INPUT.md`](../../.routing/INPUT.md) *(NASA Five Hazards & objectives)*
- [`.routing/reference/`](../../.routing/reference/) *(all 5 hazard summaries)*
- [`.routing/SPECIFICATION.md`](../../.routing/SPECIFICATION.md) *(RF04–RF07, RNF01, RNF09)*
- [`db/001_schema.sql`](../../db/001_schema.sql) *(existing tables & enums)*
- [`db/002_seed.sql`](../../db/002_seed.sql) *(exact 23 `metric.code` identifiers)*
- [`sqlc.yaml`](../../sqlc.yaml)
- [`shared/model/telemetry.go`](../../shared/model/telemetry.go)
- [`server/internal/query/service.go`](../../server/internal/query/service.go)
- [`server/internal/http/router.go`](../../server/internal/http/router.go)

## Grounding & Anti-Hallucination Constraints

1. **Five Hazards Only:** Seed `space_hazard` strictly with NASA's 5 hazards from `.routing/reference/*.md`: `space_radiation`, `isolation_confinement`, `distance_from_earth`, `gravity_fields`, `hostile_closed_env`.
2. **Exact Metric Codes:** Every `related_metric_codes` entry in `space_hazard` and `medical_procedure` MUST match a valid `metric.code` in `db/002_seed.sql` (`spo2`, `cabin_o2`, `cabin_ppco2`, `resp_rate`, `cabin_pressure`, `suit_pressure`, `bp_systolic`, `bp_diastolic`, `radiation_rate`, `radiation_dose_cum`, `heart_rate`, `hrv_rmssd`, `core_temp`, `cabin_temp`, `suit_temp`, `sleep_hours`, `mood`, `blood_glucose`, `hydration_intake`, `body_mass`, `rbc_count`, `wbc_count`, `ultrasound`).
3. **No Invented Medication Dosages or Thresholds:** Keep `medical_procedure.steps` and `warnings` strictly non-physician operational checklists (assessment, posture/hydration/shelter/scrub/recheck, equipment prep, CMO / Flight Surgeon escalation) without fabricated prescription mg dosages or unverified numeric clinical cutoffs. Tag every seeded record's `source` honestly (e.g., `'NASA HRP / ISS Operational Checklist (demo - clinical review required)'` or `'NASA HRP Five Hazards'`).
4. **Preserve Existing Integration Tests:** Existing `ingestion` tests execute `001_schema.sql`, `002_seed.sql`, `003_demo_seed.sql`. Adding tables in `001_schema.sql` must keep all existing schema objects, triggers, and tests intact. Also provide `db/004_medical_manual.sql` as an **idempotent upgrade + seed script** (`DO $$ BEGIN CREATE TYPE escalation_level ... EXCEPTION WHEN duplicate_object THEN NULL; END $$;`, `CREATE TABLE IF NOT EXISTS ...`, `INSERT ... ON CONFLICT (code) DO UPDATE`) so running `wails3 task db:seed` updates an already-initialized local `hrig` database safely without data loss.

## Task

1. **Update PostgreSQL Schema (`db/001_schema.sql`):**
   - Add enum:
     ```sql
     CREATE TYPE escalation_level AS ENUM ('self_care', 'cmo', 'flight_surgeon');
     ```
   - Add table `space_hazard`:
     - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
     - `code text NOT NULL UNIQUE`
     - `name text NOT NULL`
     - `summary text NOT NULL`
     - `clinical_impact text NOT NULL`
     - `related_metric_codes text[] NOT NULL DEFAULT '{}'`
     - `source text NOT NULL DEFAULT 'NASA HRP Five Hazards'`
     - `display_order smallint NOT NULL DEFAULT 0`
   - Add table `medical_procedure`:
     - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
     - `code text NOT NULL UNIQUE`
     - `title text NOT NULL`
     - `category text NOT NULL`
     - `body_system text NOT NULL`
     - `summary text NOT NULL`
     - `when_to_use text NOT NULL`
     - `escalation escalation_level NOT NULL DEFAULT 'self_care'`
     - `estimated_minutes smallint NOT NULL DEFAULT 5`
     - `environment environment_mode` *(nullable: `'iva'`, `'eva'`, or `NULL` for both)*
     - `equipment text[] NOT NULL DEFAULT '{}'`
     - `steps text[] NOT NULL DEFAULT '{}'`
     - `warnings text[] NOT NULL DEFAULT '{}'`
     - `related_metric_codes text[] NOT NULL DEFAULT '{}'`
     - `hazard_codes text[] NOT NULL DEFAULT '{}'`
     - `keywords text[] NOT NULL DEFAULT '{}'`
     - `source text NOT NULL DEFAULT 'NASA HRP / demo - clinical review required'`
     - `is_active boolean NOT NULL DEFAULT true`
     - `display_order smallint NOT NULL DEFAULT 0`
     - `CHECK (estimated_minutes > 0)` and `CHECK (array_length(steps, 1) >= 1)`
2. **Create Idempotent Medical Knowledge Seed & Upgrade (`db/004_medical_manual.sql`, `Taskfile.yml`, `db/README.md`):**
   - Begin `db/004_medical_manual.sql` with safe idempotent `DO $$` / `CREATE TABLE IF NOT EXISTS` statements matching `db/001_schema.sql` so it runs cleanly on both fresh and pre-existing `hrig` databases.
   - Seed all **5 `space_hazard`** rows (`space_radiation`, `isolation_confinement`, `distance_from_earth`, `gravity_fields`, `hostile_closed_env`) grounded directly in `.routing/reference/*.md`, linked to the relevant metric codes.
   - Seed **20 `medical_procedure`** rows covering all 23 DB metrics across operational/medical categories (`emergency`, `respiratory_cabin`, `cardiovascular`, `radiation`, `temperature`, `metabolic_hydration`, `behavioral_sleep`, `diagnostics_imaging`, `measurement_guide`):
     1. `hypoxia_cabin_o2` — Hypoxia & Low Cabin O2 Response (`cmo`, `cabin_o2`, `spo2`, `resp_rate`)
     2. `hypercapnia_co2_scrub` — Hypercapnia / Elevated Cabin CO2 Protocol (`cmo`, `cabin_ppco2`, `resp_rate`)
     3. `cabin_depressurization` — Cabin Depressurization Emergency (`flight_surgeon`, `cabin_pressure`, `spo2`)
     4. `eva_suit_pressure_anomaly` — EVA Suit Pressure & Thermal Anomaly (`flight_surgeon`, `suit_pressure`, `suit_temp`, `heart_rate`)
     5. `orthostatic_hypotension` — Low Blood Pressure & Orthostatic Recovery (`self_care`, `bp_systolic`, `bp_diastolic`, `hydration_intake`)
     6. `hypertensive_assessment` — Elevated Blood Pressure Assessment (`cmo`, `bp_systolic`, `bp_diastolic`, `heart_rate`)
     7. `tachycardia_assessment` — Elevated Heart Rate & Low HRV Check (`cmo`, `heart_rate`, `hrv_rmssd`)
     8. `bradycardia_syncope` — Severe Bradycardia / Near-Syncope Response (`flight_surgeon`, `heart_rate`, `bp_systolic`, `spo2`)
     9. `acute_radiation_spe` — Solar Particle Event (SPE) Radiation Shelter (`flight_surgeon`, `radiation_rate`, `radiation_dose_cum`)
     10. `cumulative_radiation_review` — Cumulative Radiation Dose Review (`cmo`, `radiation_dose_cum`, `wbc_count`, `rbc_count`)
     11. `hyperthermia_cooling` — Hyperthermia & Heat Stress Cooling (`cmo`, `core_temp`, `cabin_temp`, `suit_temp`, `hydration_intake`)
     12. `hypothermia_rewarming` — Hypothermia & Cold Stress Rewarming (`cmo`, `core_temp`, `cabin_temp`)
     13. `hypoglycemia_recovery` — Low Blood Glucose Rapid Recovery (`cmo`, `blood_glucose`)
     14. `hyperglycemia_monitoring` — Elevated Blood Glucose & Hydration Check (`self_care`, `blood_glucose`, `hydration_intake`)
     15. `dehydration_rehydration` — Fluid Deficit & Rehydration Countermeasure (`self_care`, `hydration_intake`, `body_mass`, `bp_systolic`)
     16. `circadian_sleep_recovery` — Circadian Shift & Sleep Deficit Countermeasure (`self_care`, `sleep_hours`, `mood`, `hrv_rmssd`)
     17. `behavioral_confinement_support` — Isolation, Stress & Mood Check-In (`self_care`, `mood`, `sleep_hours`)
     18. `hematology_immune_follow_up` — Abnormal Blood Panel (WBC / RBC) Follow-Up (`cmo`, `wbc_count`, `rbc_count`, `core_temp`)
     19. `ocular_sans_ultrasound` — Ocular / Optic Nerve Ultrasound & SANS Check (`cmo`, `ultrasound`, `cabin_ppco2`)
     20. `manual_vitals_guide` — How to Measure Manual Crew Vitals (`self_care`, `bp_systolic`, `bp_diastolic`, `blood_glucose`, `body_mass`, `mood`, `sleep_hours`)
   - Ensure each procedure includes rich searchable `keywords` (symptoms/synonyms like `shortness of breath`, `dizziness`, `fainting`, `headache`, `solar storm`, `fever`, `chills`, `insomnia`, `stress`, `eye vision`, `blood sugar`, etc.), clear `steps` (4–6 actionable items), and `warnings` (2–3 escalation triggers).
   - Append `- ./local/pgsql/bin/psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f ./db/004_medical_manual.sql` to `db:seed` in `Taskfile.yml`, and mention `004_medical_manual.sql` in `db/README.md`.
3. **Add SQLc Queries (`db/queries/medical.sql`) and Regenerate `server/internal/db/`:**
   - Add queries:
     - `-- name: ListSpaceHazards :many` ordered by `display_order ASC, code ASC`
     - `-- name: ListMedicalProcedures :many` ordered by `display_order ASC, code ASC` (filtering `is_active = true`)
     - `-- name: SearchMedicalProcedures :many` with optional `q text` (`ILIKE '%' || $1 || '%'` across `title`, `summary`, `when_to_use`, `category`, `body_system`, `array_to_string(keywords, ' ')`, `array_to_string(related_metric_codes, ' ')`), optional `category text`, optional `metric_code text` (`$3 = '' OR $3 = ANY(related_metric_codes)`), ordered by `display_order ASC, code ASC`.
     - `-- name: SearchSpaceHazards :many` with `q text` (`$1 = '' OR name ILIKE ... OR summary ILIKE ... OR clinical_impact ILIKE ... OR array_to_string(related_metric_codes, ' ') ILIKE ...`).
   - Run `wails3 task db:sqlc` and ensure `server/internal/db/` compiles cleanly.
4. **Update Shared Contracts (`shared/model/telemetry.go` & `telemetry_test.go`):**
   - Add `EscalationLevel` string type and constants:
     - `EscalationSelfCare EscalationLevel = "self_care"`
     - `EscalationCMO EscalationLevel = "cmo"`
     - `EscalationFlightSurgeon EscalationLevel = "flight_surgeon"`
   - Add `SpaceHazard` struct (`id`, `code`, `name`, `summary`, `clinicalImpact`, `relatedMetricCodes`, `source`, `displayOrder`).
   - Add `MedicalProcedure` struct (`id`, `code`, `title`, `category`, `bodySystem`, `summary`, `whenToUse`, `escalation`, `estimatedMinutes`, `environment` `*EnvironmentMode`, `equipment`, `steps`, `warnings`, `relatedMetricCodes`, `hazardCodes`, `keywords`, `source`, `displayOrder`).
   - Add `SearchResults` struct (`query`, `tab`, `astronauts []Astronaut`, `procedures []MedicalProcedure`, `hazards []SpaceHazard`, `metrics []Metric`).
   - Always normalize slices on JSON structs to non-nil empty slices where appropriate and add JSON round-trip unit tests in `shared/model/telemetry_test.go`.
5. **Implement Query Methods & HTTP Endpoints (`server/internal/query/` & `server/internal/http/`):**
   - Extend `Reader` interface and `Service` in `server/internal/query/service.go`:
     - `ListHazards(ctx context.Context, query string) ([]model.SpaceHazard, error)`
     - `ListProcedures(ctx context.Context, query string, category string, metricCode string) ([]model.MedicalProcedure, error)`
     - `Search(ctx context.Context, q string, tab string) (model.SearchResults, error)`:
       - Validate `tab` in `{"", "all", "crew", "procedures", "hazards_metrics"}` (return `ErrBadRequest` on unknown `tab`).
       - Filter astronauts (name, nationality code, nationality label), procedures, hazards, and metrics (`code`, `name`, `shortName`, `category`) cleanly; when `tab` targets a single category, return empty slices for the excluded categories while keeping valid normalized results.
   - Update `server/internal/http/router.go`:
     - Mount `GET /api/hazards` (optional `q`), `GET /api/procedures` (optional `q`, `category`, `metric_code`), and `GET /api/search` (optional `q`, `tab`), returning `503` if `Reader == nil` and `400` on `query.ErrBadRequest`.
   - Add unit tests in `server/internal/query/service_test.go` and `server/internal/http/router_test.go` (do not forget updating any existing `mockReader` structs cleanly), plus integration checks in `server/internal/query/service_integration_test.go` loading `004_medical_manual.sql` and asserting 5 hazards, 20 procedures, and `/api/search` filtering.

## Verification

Run every check in [spec.md](spec.md). Also apply `004_medical_manual.sql` to the local `hrig` database (via `wails3 task db:start` and `wails3 task db:seed`), run the PostgreSQL integration suite on disposable `hrig_test`, then drop `hrig_test` and stop PostgreSQL if started solely for testing.

## Change log

Write `.tickets/014E-medical-knowledge-db-and-search-api/changelog.md` listing every file touched, every verification command and its outcome, and any assumptions or edge cases encountered.
