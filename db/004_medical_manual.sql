-- ============================================================================
-- HRIG · Medical Knowledge Manual (Hazards & Operational Procedures)
-- 004_medical_manual.sql — idempotent schema upgrade & catalog seed
-- Grounded in NASA HRP Five Hazards (.routing/reference/*.md) and db/002_seed.sql.
-- Re-running this file safely updates existing rows without duplication.
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- Idempotent DDL for pre-existing databases
-- ----------------------------------------------------------------------------
DO $$
BEGIN
    CREATE TYPE escalation_level AS ENUM ('self_care', 'cmo', 'flight_surgeon');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS space_hazard (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 text NOT NULL UNIQUE,
    name                 text NOT NULL,
    summary              text NOT NULL,
    clinical_impact      text NOT NULL,
    related_metric_codes text[] NOT NULL DEFAULT '{}',
    source               text NOT NULL DEFAULT 'NASA HRP Five Hazards',
    display_order        smallint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS medical_procedure (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 text NOT NULL UNIQUE,
    title                text NOT NULL,
    category             text NOT NULL,
    body_system          text NOT NULL,
    summary              text NOT NULL,
    when_to_use          text NOT NULL,
    escalation           escalation_level NOT NULL DEFAULT 'self_care',
    estimated_minutes    smallint NOT NULL DEFAULT 5,
    environment          environment_mode,
    equipment            text[] NOT NULL DEFAULT '{}',
    steps                text[] NOT NULL DEFAULT '{}',
    warnings             text[] NOT NULL DEFAULT '{}',
    related_metric_codes text[] NOT NULL DEFAULT '{}',
    hazard_codes         text[] NOT NULL DEFAULT '{}',
    keywords             text[] NOT NULL DEFAULT '{}',
    source               text NOT NULL DEFAULT 'NASA HRP / demo - clinical review required',
    is_active            boolean NOT NULL DEFAULT true,
    display_order        smallint NOT NULL DEFAULT 0,
    CONSTRAINT medical_procedure_est_minutes CHECK (estimated_minutes > 0),
    CONSTRAINT medical_procedure_has_steps   CHECK (array_length(steps, 1) >= 1)
);

CREATE INDEX IF NOT EXISTS idx_medical_procedure_active
    ON medical_procedure (is_active, display_order, code);

-- ----------------------------------------------------------------------------
-- 1. NASA HRP Five Spaceflight Hazards
-- ----------------------------------------------------------------------------
INSERT INTO space_hazard (
    code,
    name,
    summary,
    clinical_impact,
    related_metric_codes,
    source,
    display_order
) VALUES
    (
        'space_radiation',
        'Space Radiation',
        'Invisible ionizing radiation—including galactic cosmic rays, solar particle events (SPE), and secondary neutron radiation—encountered beyond Earth''s protective atmosphere and magnetic field.',
        'Causes cellular and DNA damage that increases long-term cancer risk and acute biological harm; neutron radiation contributes roughly 30% of total ISS exposure and drives cardiovascular, central nervous system, and immune changes.',
        ARRAY['radiation_rate', 'radiation_dose_cum', 'wbc_count', 'rbc_count', 'heart_rate'],
        'NASA HRP Five Hazards',
        1
    ),
    (
        'isolation_confinement',
        'Isolation and Confinement',
        'Prolonged separation from Earth, family, and familiar social environments while living and working inside a small, enclosed crew habitat.',
        'Triggers behavioral and psychological stress over long missions, disrupts circadian rhythms, sleep duration/quality, and mood, and can impair crew morale, decision-making, and interpersonal cohesion.',
        ARRAY['sleep_hours', 'mood', 'hrv_rmssd', 'heart_rate'],
        'NASA HRP Five Hazards',
        2
    ),
    (
        'distance_from_earth',
        'Distance from Earth',
        'Extreme physical separation from Earth that introduces communication latency, prevents rapid resupply, and eliminates emergency return options.',
        'Flight surgeons cannot provide real-time assistance on deep-space missions, and onboard medications, food, and medical supplies are strictly limited—requiring astronauts to monitor indicators autonomously, assess symptoms, and act early.',
        ARRAY['spo2', 'heart_rate', 'bp_systolic', 'core_temp', 'blood_glucose', 'ultrasound'],
        'NASA HRP Five Hazards',
        3
    ),
    (
        'gravity_fields',
        'Gravity Fields',
        'Exposure to altered gravity environments—microgravity in transit or orbit, partial gravity on the Moon or Mars, and hypergravity/1g transitions on launch and return.',
        'Causes bone density loss and muscle atrophy from reduced mechanical loading, shifts bodily fluids upward, deconditions the cardiovascular system, and affects the vestibular system, vision (SANS), balance, and organ function.',
        ARRAY['bp_systolic', 'bp_diastolic', 'heart_rate', 'body_mass', 'hydration_intake', 'ultrasound'],
        'NASA HRP Five Hazards',
        4
    ),
    (
        'hostile_closed_env',
        'Hostile / Closed Environments',
        'Self-contained spacecraft or habitat ecosystems where air quality, pressure, temperature, lighting, noise, water, and microbes are tightly coupled and recycled.',
        'Confined, recycled systems amplify microbial transmission and alter crew immune responses, while suboptimal temperature, CO2 levels, noise, or lighting degrade sleep and physiological comfort, requiring continuous dual monitoring of habitat and crew.',
        ARRAY['cabin_o2', 'cabin_ppco2', 'cabin_pressure', 'cabin_temp', 'suit_pressure', 'suit_temp', 'resp_rate', 'wbc_count', 'sleep_hours'],
        'NASA HRP Five Hazards',
        5
    )
ON CONFLICT (code) DO UPDATE
    SET name                 = EXCLUDED.name,
        summary              = EXCLUDED.summary,
        clinical_impact      = EXCLUDED.clinical_impact,
        related_metric_codes = EXCLUDED.related_metric_codes,
        source               = EXCLUDED.source,
        display_order        = EXCLUDED.display_order;

-- ----------------------------------------------------------------------------
-- 2. Operational Medical Procedures (20 checklists covering all 23 metrics)
-- ----------------------------------------------------------------------------
INSERT INTO medical_procedure (
    code,
    title,
    category,
    body_system,
    summary,
    when_to_use,
    escalation,
    estimated_minutes,
    environment,
    equipment,
    steps,
    warnings,
    related_metric_codes,
    hazard_codes,
    keywords,
    source,
    is_active,
    display_order
) VALUES
    (
        'hypoxia_cabin_o2',
        'Hypoxia & Low Cabin O2 Response',
        'respiratory_cabin',
        'respiratory',
        'Operational checklist to verify cabin oxygen delivery, stabilize crew oxygenation, and isolate atmospheric supply faults.',
        'When cabin O2 drops into warning or critical alerts, crew SpO2 trends downward, or crew report air hunger, dizziness, or confusion.',
        'cmo'::escalation_level,
        10,
        'iva'::environment_mode,
        ARRAY['Pulse oximeter', 'Portable O2 mask & regulator', 'Cabin environmental console'],
        ARRAY[
            'Cross-check cabin O2 sensor telemetry against a backup atmospheric analyzer and verify crew SpO2 and respiratory rate.',
            'Direct affected crew members to pause strenuous tasks, sit or brace in a neutral posture, and breathe calmly.',
            'Inspect oxygen generation and supply manifold valves, cabin circulation fans, and hatch seals for flow restrictions.',
            'If SpO2 remains depressed or symptoms persist, initiate supplemental breathing oxygen per spacecraft life-support procedures.',
            'Notify the Crew Medical Officer (CMO) and log updated SpO2, respiratory rate, and cabin O2 readings into HRIG.'
        ],
        ARRAY[
            'Escalate immediately to the Flight Surgeon if cabin O2 cannot be restored or crew exhibit cyanosis, confusion, or loss of consciousness.',
            'Never silence an atmospheric O2 alert until a physical sensor cross-check confirms nominal oxygen partial pressure.'
        ],
        ARRAY['cabin_o2', 'spo2', 'resp_rate'],
        ARRAY['hostile_closed_env', 'distance_from_earth'],
        ARRAY['hypoxia', 'low oxygen', 'shortness of breath', 'dyspnea', 'dizziness', 'air hunger', 'cyanosis', 'cabin o2'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        1
    ),
    (
        'hypercapnia_co2_scrub',
        'Hypercapnia / Elevated Cabin CO2 Protocol',
        'respiratory_cabin',
        'respiratory',
        'Checklist for mitigating elevated cabin carbon dioxide partial pressure and monitoring crew respiratory compensation.',
        'When cabin ppCO2 triggers an alert or crew members experience headache, flushed skin, dyspnea, or impaired concentration.',
        'cmo'::escalation_level,
        15,
        'iva'::environment_mode,
        ARRAY['CO2 scrubber status panel', 'Portable CO2 monitor', 'Airflow circulation fan'],
        ARRAY[
            'Verify cabin ppCO2 telemetry with a secondary handheld monitor and assess crew respiratory rate.',
            'Inspect CO2 removal assembly (CDRA / amine scrubber) beds, blower airflow, and inter-module ventilation ducts for blockage.',
            'Increase cabin air mixing with portable fans to eliminate localized CO2 pockets around crew workstations and sleep quarters.',
            'Have symptomatic crew rest in a well-ventilated module and recheck respiratory rate and headache status after scrubbing stabilizes.',
            'Report persistent elevation or neurological symptoms to the CMO and record environmental and crew observations.'
        ],
        ARRAY[
            'Escalate to the Flight Surgeon if elevated ppCO2 is accompanied by severe lethargy, disorientation, or visual disturbances.',
            'Avoid heavy physical exercise in stagnant cabin zones while CO2 scrubbing capacity is degraded.'
        ],
        ARRAY['cabin_ppco2', 'resp_rate'],
        ARRAY['hostile_closed_env'],
        ARRAY['hypercapnia', 'co2', 'carbon dioxide', 'headache', 'scrubber', 'stagnant air', 'shortness of breath', 'lethargy'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        2
    ),
    (
        'cabin_depressurization',
        'Cabin Depressurization Emergency',
        'emergency',
        'respiratory',
        'Immediate life-support and crew isolation response for suspected or confirmed habitat pressure loss.',
        'When cabin pressure drops rapidly or triggers a critical depressurization alarm.',
        'flight_surgeon'::escalation_level,
        5,
        'iva'::environment_mode,
        ARRAY['Emergency breathing mask', 'IVA pressure garment', 'Hatch isolation controls', 'Pulse oximeter'],
        ARRAY[
            'Don emergency oxygen masks immediately and initiate spacecraft depressurization isolation callouts across all crew.',
            'Close and verify module hatches sequentially according to cabin pressure telemetry to isolate the leaking compartment.',
            'Confirm all crew are accounted for in a pressurized safe haven and monitor SpO2 continuously.',
            'Establish immediate communication with Mission Control / Flight Surgeon (accounting for deep-space latency) and report pressure trend.',
            'Monitor crew for decompression symptoms (joint pain, visual changes, shortness of breath) once pressure is stabilized.'
        ],
        ARRAY[
            'Do not reopen an isolated module hatch until structural integrity and pressure stability are verified by flight procedures.',
            'Treat any post-event joint pain, neurological deficit, or persistent hypoxemia as a Flight Surgeon emergency.'
        ],
        ARRAY['cabin_pressure', 'spo2'],
        ARRAY['hostile_closed_env', 'distance_from_earth'],
        ARRAY['depressurization', 'pressure leak', 'vacuum', 'decompression', 'hatch', 'emergency', 'hypoxia', 'bends'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        3
    ),
    (
        'eva_suit_pressure_anomaly',
        'EVA Suit Pressure & Thermal Anomaly',
        'emergency',
        'multi_system',
        'Extravehicular activity abort and airlock recovery steps for suit pressure or liquid cooling garment anomalies.',
        'During EVA when suit pressure or suit temperature leaves nominal bounds or heart rate spikes unexpectedly.',
        'flight_surgeon'::escalation_level,
        10,
        'eva'::environment_mode,
        ARRAY['EVA suit display & cuff controls', 'Airlock repressurization system', 'Wearable biomonitor'],
        ARRAY[
            'Halt EVA translation and anchor safely while verifying suit pressure, suit temperature, and heart rate telemetry with the IVA crewmate.',
            'Check suit cooling loop flow setting and secondary oxygen pack status per EVA emergency card.',
            'Initiate immediate return to the airlock if suit pressure trend indicates a leak or thermal control cannot stabilize.',
            'Complete airlock repressurization and doff the helmet/suit with IVA crew assistance while monitoring heart rate and symptom status.',
            'Brief the CMO and Flight Surgeon with recorded suit telemetry and perform a post-EVA physical check.'
        ],
        ARRAY[
            'Never continue an EVA task with an unverified or downward-trending suit pressure reading.',
            'Watch closely for heat exhaustion, cold injury, or decompression symptoms immediately after suit doffing.'
        ],
        ARRAY['suit_pressure', 'suit_temp', 'heart_rate'],
        ARRAY['hostile_closed_env', 'distance_from_earth'],
        ARRAY['eva', 'spacewalk', 'suit leak', 'suit pressure', 'cooling loop', 'airlock', 'overheating', 'palpitations'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        4
    ),
    (
        'orthostatic_hypotension',
        'Low Blood Pressure & Orthostatic Recovery',
        'cardiovascular',
        'cardiovascular',
        'Self-care countermeasure for gravity-transition lightheadedness and reduced blood pressure.',
        'When systolic or diastolic blood pressure reads low or a crew member feels dizzy during posture changes or post-landing/partial-gravity transitions.',
        'self_care'::escalation_level,
        15,
        NULL,
        ARRAY['Blood pressure cuff', 'Oral hydration pouch / electrolyte drink', 'Compression garment (if assigned)'],
        ARRAY[
            'Have the crew member sit or lie supine immediately and elevate legs slightly if in a gravity field to restore central venous return.',
            'Measure and log repeat systolic and diastolic blood pressure alongside pulse rate after 5 minutes of rest.',
            'Encourage oral fluid and electrolyte intake per standard hydration countermeasure guidance.',
            'Perform gentle lower-extremity muscle tensing before standing slowly in staged increments.',
            'Log fluid intake and follow-up blood pressure values in HRIG and inform the CMO if symptoms recur.'
        ],
        ARRAY[
            'Escalate to the CMO or Flight Surgeon if fainting (syncope), chest discomfort, or persistent hypotension occurs.',
            'Do not allow solo operation of hatches or heavy equipment while orthostatic dizziness persists.'
        ],
        ARRAY['bp_systolic', 'bp_diastolic', 'hydration_intake'],
        ARRAY['gravity_fields', 'distance_from_earth'],
        ARRAY['low blood pressure', 'hypotension', 'dizziness', 'lightheaded', 'orthostatic', 'fainting', 'standing up', 'fluid shift'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        5
    ),
    (
        'hypertensive_assessment',
        'Elevated Blood Pressure Assessment',
        'cardiovascular',
        'cardiovascular',
        'Standardized rest-and-recheck workflow to confirm elevated blood pressure readings and screen for red-flag symptoms.',
        'When blood pressure cuff measurements exceed warning or critical thresholds during routine or symptomatic checks.',
        'cmo'::escalation_level,
        15,
        NULL,
        ARRAY['Blood pressure cuff', 'Wearable biomonitor', 'Timer / stopwatch'],
        ARRAY[
            'Verify proper cuff size and arm positioning at heart level after the crew member rests quietly for at least 5 minutes.',
            'Repeat systolic and diastolic blood pressure measurements twice, spaced several minutes apart, and record heart rate.',
            'Screen for associated symptoms including severe headache, vision changes, chest tightness, or shortness of breath.',
            'Review recent exercise, caffeine intake, stress load, and cabin CO2 levels as potential transient contributors.',
            'Submit verified readings to the CMO for trend review and schedule serial blood pressure follow-up.'
        ],
        ARRAY[
            'Escalate immediately to the Flight Surgeon if elevated blood pressure is paired with chest pain, neurological deficits, or acute visual changes.',
            'Do not base clinical escalation on a single unverified cuff reading taken immediately after resistive exercise.'
        ],
        ARRAY['bp_systolic', 'bp_diastolic', 'heart_rate'],
        ARRAY['gravity_fields', 'isolation_confinement'],
        ARRAY['high blood pressure', 'hypertension', 'headache', 'chest tightness', 'blood pressure cuff', 'stress'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        6
    ),
    (
        'tachycardia_assessment',
        'Elevated Heart Rate & Low HRV Check',
        'cardiovascular',
        'cardiovascular',
        'Assessment protocol for sustained resting tachycardia or depressed heart rate variability.',
        'When wearable heart rate stays elevated at rest or HRV (RMSSD) drops below personal baseline.',
        'cmo'::escalation_level,
        10,
        NULL,
        ARRAY['Wearable biomonitor', 'Pulse oximeter', 'Oral thermometer', 'Blood pressure cuff'],
        ARRAY[
            'Confirm whether the crew member recently completed exercise or an EVA; if not, have them rest quietly for 5 minutes.',
            'Check wearable heart rate and HRV alongside manual radial pulse, SpO2, blood pressure, and core temperature.',
            'Assess hydration status, recent sleep duration, and potential cabin stressors (elevated temperature or ppCO2).',
            'Encourage rest and oral hydration if dehydrated or thermally stressed, and re-evaluate heart rate in 15 minutes.',
            'Notify the CMO if resting heart rate remains elevated or if palpitations or dyspnea are present.'
        ],
        ARRAY[
            'Escalate to the Flight Surgeon if tachycardia is accompanied by chest pain, syncope, fever, or hypoxemia.',
            'Review ECG/waveform telemetry if available on the onboard medical workstation.'
        ],
        ARRAY['heart_rate', 'hrv_rmssd'],
        ARRAY['gravity_fields', 'isolation_confinement', 'hostile_closed_env'],
        ARRAY['tachycardia', 'fast heart rate', 'palpitations', 'racing heart', 'low hrv', 'stress', 'autonomic'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        7
    ),
    (
        'bradycardia_syncope',
        'Severe Bradycardia / Near-Syncope Response',
        'emergency',
        'cardiovascular',
        'Urgent crew stabilization and medical escalation checklist for abnormally slow heart rate or near-fainting.',
        'When heart rate triggers a low critical/warning alert or a crew member experiences near-syncope, weakness, or presyncope.',
        'flight_surgeon'::escalation_level,
        10,
        NULL,
        ARRAY['Wearable biomonitor', 'Blood pressure cuff', 'Pulse oximeter', 'Onboard AED / cardiac monitor'],
        ARRAY[
            'Position the crew member safely in a supported horizontal or seated restraint posture and ensure a clear airway.',
            'Check responsiveness, manual pulse, heart rate telemetry, blood pressure, and SpO2 immediately.',
            'Keep the onboard cardiac monitor / AED kit staged nearby without delaying communication.',
            'Contact the CMO and transmit an urgent status report to the Flight Surgeon with full vitals and symptom timeline.',
            'Maintain continuous observation and log serial readings every few minutes until cleared by medical command.'
        ],
        ARRAY[
            'Never leave a symptomatic bradycardic or near-syncopal crew member unattended.',
            'Initiate emergency resuscitation protocols immediately if the crew member becomes unresponsive or pulseless.'
        ],
        ARRAY['heart_rate', 'bp_systolic', 'spo2'],
        ARRAY['gravity_fields', 'distance_from_earth'],
        ARRAY['bradycardia', 'slow heart rate', 'syncope', 'fainting', 'collapse', 'weakness', 'presyncope', 'unresponsive'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        8
    ),
    (
        'acute_radiation_spe',
        'Solar Particle Event (SPE) Radiation Shelter',
        'radiation',
        'radiation_hematologic',
        'Emergency sheltering and dosimetry protocol during elevated solar particle radiation flux.',
        'When radiation exposure rate alerts trigger or Mission Control issues a Solar Particle Event warning.',
        'flight_surgeon'::escalation_level,
        15,
        NULL,
        ARRAY['Personal radiation dosimeter', 'Area radiation monitor', 'Water/polyethylene shelter zone'],
        ARRAY[
            'Verify elevated exposure rate on both personal dosimeters and fixed spacecraft area monitors.',
            'Direct all crew to suspend non-essential tasks (and abort any active EVA immediately) and relocate to the designated shielded shelter volume.',
            'Confirm every crew member is wearing an active personal dosimeter and log shelter entry time in HRIG.',
            'Monitor cumulative dose and exposure rate updates while maintaining communications with Flight Control and Flight Surgeon.',
            'Remain in shelter until radiation rate returns to nominal background levels and Flight Control issues an all-clear.'
        ],
        ARRAY[
            'Do not exit the shielded storm shelter early during transient dips in solar particle flux.',
            'Report any nausea, fatigue, or acute prodromal symptoms after high-rate exposure directly to the Flight Surgeon.'
        ],
        ARRAY['radiation_rate', 'radiation_dose_cum'],
        ARRAY['space_radiation', 'distance_from_earth'],
        ARRAY['solar storm', 'solar particle event', 'spe', 'radiation alert', 'shelter', 'dosimeter', 'radiation rate'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        9
    ),
    (
        'cumulative_radiation_review',
        'Cumulative Radiation Dose Review',
        'radiation',
        'radiation_hematologic',
        'Periodic exposure accounting and hematologic correlation checklist for cumulative mission radiation dose.',
        'When cumulative radiation dose approaches warning limits or during scheduled crew radiation safety reviews.',
        'cmo'::escalation_level,
        20,
        'iva'::environment_mode,
        ARRAY['Personal radiation dosimeter', 'Blood panel analyzer (lab_assay_01)', 'Dosimetry log'],
        ARRAY[
            'Audit individual cumulative radiation dose records across all assigned wearable dosimeters for consistency.',
            'Review scheduled EVA assignments and high-exposure habitat work zones with the Mission Commander and CMO.',
            'Conduct a follow-up complete blood count panel (WBC and RBC counts) per mission medical schedule.',
            'Assess crew for fatigue, recurrent infections, or healing delays and document findings in HRIG.',
            'Submit cumulative exposure summary to the Flight Surgeon for mission-level dose management planning.'
        ],
        ARRAY[
            'Escalate to the Flight Surgeon before scheduling additional high-exposure EVAs if cumulative dose alerts are active.',
            'Report any unexpected dosimeter drift or missing badge data promptly.'
        ],
        ARRAY['radiation_dose_cum', 'wbc_count', 'rbc_count'],
        ARRAY['space_radiation', 'distance_from_earth'],
        ARRAY['cumulative radiation', 'msv', 'dosimetry', 'cancer risk', 'gcr', 'neutron', 'blood panel', 'wbc', 'rbc'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        10
    ),
    (
        'hyperthermia_cooling',
        'Hyperthermia & Heat Stress Cooling',
        'temperature',
        'thermoregulatory',
        'Active cooling and hydration protocol for elevated core, cabin, or EVA suit temperatures.',
        'When core body temperature, cabin temperature, or suit temperature triggers warning/critical high alerts or crew report heat exhaustion.',
        'cmo'::escalation_level,
        20,
        NULL,
        ARRAY['Core temperature thermometer', 'Cooling vest / cold packs', 'Oral electrolyte hydration', 'Airflow fan'],
        ARRAY[
            'Stop physical exertion immediately, move to the coolest ventilated cabin zone (or maximize EVA suit cooling), and loosen restrictive layers.',
            'Measure core body temperature and verify cabin or suit temperature telemetry to distinguish environmental load from fever.',
            'Apply active surface cooling (wet towels, cooling vest, or cold packs to neck, axillae, and groin) while circulating air.',
            'Provide cool oral fluids in steady sips if the crew member is alert and not nauseated; log hydration intake.',
            'Recheck core temperature and mental status every 15 minutes and notify the CMO.'
        ],
        ARRAY[
            'Escalate immediately to the Flight Surgeon if core temperature continues rising or if confusion, cessation of sweating, or collapse occurs.',
            'If cabin temperature remains normal while core temperature is elevated, evaluate for infectious fever rather than environmental heat stress.'
        ],
        ARRAY['core_temp', 'cabin_temp', 'suit_temp', 'hydration_intake'],
        ARRAY['hostile_closed_env', 'distance_from_earth'],
        ARRAY['fever', 'hyperthermia', 'heat stress', 'heat exhaustion', 'overheating', 'high temperature', 'sweating', 'cooling'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        11
    ),
    (
        'hypothermia_rewarming',
        'Hypothermia & Cold Stress Rewarming',
        'temperature',
        'thermoregulatory',
        'Passive and active external rewarming protocol for low core body temperature or cold cabin exposure.',
        'When core temperature or cabin temperature drops below warning thresholds or a crew member exhibits persistent shivering or sluggishness.',
        'cmo'::escalation_level,
        25,
        'iva'::environment_mode,
        ARRAY['Core temperature thermometer', 'Thermal blankets / sleep bag', 'Warm oral fluids', 'Cabin thermal control panel'],
        ARRAY[
            'Move the crew member out of cold drafts, remove any damp clothing, and insulate them in dry thermal layers or a crew sleep restraint.',
            'Verify cabin thermal control settings and restore nominal cabin heating airflow.',
            'Measure core temperature, heart rate, and blood pressure gently without abrupt vigorous movement.',
            'Offer warm (not scalding) sweet oral fluids if the crew member is fully alert and able to swallow safely.',
            'Monitor core temperature trend until normothermia is restored and update the CMO.'
        ],
        ARRAY[
            'Escalate immediately to the Flight Surgeon if shivering stops despite low core temperature, or if slurred speech and confusion appear.',
            'Handle severely hypothermic crew gently and avoid rapid hot-pack application directly to bare skin.'
        ],
        ARRAY['core_temp', 'cabin_temp'],
        ARRAY['hostile_closed_env', 'distance_from_earth'],
        ARRAY['hypothermia', 'chills', 'shivering', 'cold stress', 'low temperature', 'rewarming', 'cabin cold'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        12
    ),
    (
        'hypoglycemia_recovery',
        'Low Blood Glucose Rapid Recovery',
        'metabolic_hydration',
        'metabolic',
        'Rapid oral carbohydrate recovery and re-test checklist for low blood glucose readings.',
        'When blood glucose reads low or a crew member develops shakiness, diaphoresis, irritability, or confusion.',
        'cmo'::escalation_level,
        15,
        NULL,
        ARRAY['Blood glucose meter & test strips', 'Fast-acting oral carbohydrate gel/drink', 'Complex snack'],
        ARRAY[
            'Confirm the low reading with a clean fingerstick blood glucose check and assess alertness and ability to swallow.',
            'If alert and able to swallow safely, provide a fast-acting oral carbohydrate source from the crew medical/nutrition kit per kit instructions.',
            'Keep the crew member seated and resting while monitoring for symptom improvement.',
            'Repeat the blood glucose measurement after 15 minutes and log the updated value in HRIG.',
            'Once normalized, provide a balanced meal or snack and notify the CMO to review nutrition and workload.'
        ],
        ARRAY[
            'Never give oral fluids or food to a crew member who is drowsy, disoriented, or unable to protect their airway—escalate immediately to the Flight Surgeon.',
            'Escalate to the CMO/Flight Surgeon if blood glucose fails to recover after repeat oral carbohydrate intake.'
        ],
        ARRAY['blood_glucose'],
        ARRAY['distance_from_earth'],
        ARRAY['hypoglycemia', 'low blood sugar', 'shakiness', 'sweating', 'glucose', 'confusion', 'fatigue'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        13
    ),
    (
        'hyperglycemia_monitoring',
        'Elevated Blood Glucose & Hydration Check',
        'metabolic_hydration',
        'metabolic',
        'Verification, hydration, and symptom screening checklist for elevated blood glucose values.',
        'When blood glucose reads above warning range on routine metabolic monitoring or post-meal spot checks.',
        'self_care'::escalation_level,
        15,
        NULL,
        ARRAY['Blood glucose meter & test strips', 'Oral hydration pouch'],
        ARRAY[
            'Verify whether the measurement was fasting or post-prandial and repeat a clean fingerstick check after handwashing to rule out food residue.',
            'Encourage steady water intake to maintain hydration and log fluid volume in HRIG.',
            'Screen for symptoms such as excessive thirst, frequent urination, nausea, abdominal discomfort, or fruity breath.',
            'Schedule a follow-up fasting blood glucose check prior to the next main meal period.',
            'Share repeated elevated readings with the CMO for dietary and metabolic review.'
        ],
        ARRAY[
            'Escalate promptly to the CMO or Flight Surgeon if elevated glucose is accompanied by persistent vomiting, lethargy, or rapid breathing.',
            'Ensure test strips are stored within temperature specifications and within expiration dates.'
        ],
        ARRAY['blood_glucose', 'hydration_intake'],
        ARRAY['distance_from_earth'],
        ARRAY['hyperglycemia', 'high blood sugar', 'blood glucose', 'thirst', 'metabolic', 'fingerstick'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        14
    ),
    (
        'dehydration_rehydration',
        'Fluid Deficit & Rehydration Countermeasure',
        'metabolic_hydration',
        'metabolic_renal',
        'Structured oral rehydration and body mass tracking checklist for fluid deficit.',
        'When daily hydration intake drops below target, body mass trends down unexpectedly, or urine/orthostatic signs suggest dehydration.',
        'self_care'::escalation_level,
        20,
        NULL,
        ARRAY['Graduated hydration pouch', 'Electrolyte packets', 'Mass measurement device', 'Blood pressure cuff'],
        ARRAY[
            'Review the past 24 hours of logged fluid intake, EVA sweat loss, and recent body mass measurements.',
            'Initiate paced oral rehydration using water and onboard electrolyte packets rather than large rapid boluses.',
            'Monitor systolic blood pressure and heart rate during rest to check for orthostatic or volume-depletion strain.',
            'Record all consumed fluid volumes in HRIG and recheck body mass at the next scheduled morning weigh-in.',
            'Report persistent weight loss, dark/absent urine output, or inability to keep fluids down to the CMO.'
        ],
        ARRAY[
            'Escalate to the CMO or Flight Surgeon if dehydration is accompanied by persistent vomiting, fever, or orthostatic presyncope.',
            'Account for EVA cooling garment water consumption separately from crew oral fluid intake.'
        ],
        ARRAY['hydration_intake', 'body_mass', 'bp_systolic'],
        ARRAY['gravity_fields', 'hostile_closed_env', 'distance_from_earth'],
        ARRAY['dehydration', 'fluid deficit', 'thirst', 'weight loss', 'body mass', 'rehydration', 'electrolytes', 'water intake'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        15
    ),
    (
        'circadian_sleep_recovery',
        'Circadian Shift & Sleep Deficit Countermeasure',
        'behavioral_sleep',
        'neurobehavioral',
        'Sleep hygiene, lighting, and workload recovery checklist for circadian misalignment or short sleep.',
        'When logged sleep hours drop repeatedly, mission schedule shifts occur, or HRV and mood indicate accumulated fatigue.',
        'self_care'::escalation_level,
        15,
        'iva'::environment_mode,
        ARRAY['Crew sleep restraint', 'Programmable cabin/sleep lighting', 'Earplugs / eye mask', 'Wearable biomonitor'],
        ARRAY[
            'Audit the past 72 hours of sleep duration, HRV (RMSSD), and mood check-in scores in HRIG.',
            'Align crew quarter lighting—bright blue-enriched light during scheduled wake periods and dim warm lighting prior to sleep.',
            'Verify sleep station ventilation, noise attenuation (earplugs), and cabin temperature comfort.',
            'Coordinate with the Mission Commander to protect an uninterrupted recovery sleep block and defer non-critical late-shift tasks.',
            'Log next-day sleep hours and mood score; consult the CMO if sleep deficit persists across consecutive mission days.'
        ],
        ARRAY[
            'Avoid scheduling high-risk EVA or complex robotics operations for crew members with severe cumulative sleep deficit.',
            'Escalate to the CMO if insomnia is accompanied by severe anxiety, persistent low mood, or daytime microsleeps.'
        ],
        ARRAY['sleep_hours', 'mood', 'hrv_rmssd'],
        ARRAY['isolation_confinement', 'hostile_closed_env'],
        ARRAY['insomnia', 'sleep deficit', 'circadian', 'fatigue', 'tired', 'jet lag', 'lighting', 'sleep quality'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        16
    ),
    (
        'behavioral_confinement_support',
        'Isolation, Stress & Mood Check-In',
        'behavioral_sleep',
        'neurobehavioral',
        'Structured peer support and psychological check-in workflow for prolonged confinement stress.',
        'When mood scores drop to warning/critical levels or crew members notice irritability, withdrawal, or interpersonal strain.',
        'self_care'::escalation_level,
        20,
        'iva'::environment_mode,
        ARRAY['Private crew communication terminal', 'Mood & sleep log'],
        ARRAY[
            'Conduct a private, non-judgmental check-in with the crew member in a quiet compartment away from shared work areas.',
            'Review recent workload intensity, sleep hours, and mood check-in trends to identify acute stressors.',
            'Offer scheduled private family/ground support communication (accounting for deep-space latency via async video/audio messages).',
            'Encourage structured physical exercise and brief daily recovery breaks to support mood and cohesion.',
            'Document the check-in respectfully and arrange a private CMO or behavioral health Flight Surgeon consultation if low mood persists.'
        ],
        ARRAY[
            'Escalate immediately to the CMO and Flight Surgeon if any crew member expresses hopelessness, severe panic, or self-harm concerns.',
            'Protect crew psychological privacy when logging behavioral notes.'
        ],
        ARRAY['mood', 'sleep_hours'],
        ARRAY['isolation_confinement', 'distance_from_earth'],
        ARRAY['stress', 'anxiety', 'mood', 'depression', 'isolation', 'confinement', 'morale', 'conflict', 'mental health'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        17
    ),
    (
        'hematology_immune_follow_up',
        'Abnormal Blood Panel (WBC / RBC) Follow-Up',
        'diagnostics_imaging',
        'hematologic_immune',
        'Lab verification, infection screening, and immune follow-up checklist for abnormal blood counts.',
        'When lab assay results show out-of-range WBC or RBC counts on routine or symptomatic blood panels.',
        'cmo'::escalation_level,
        25,
        'iva'::environment_mode,
        ARRAY['Blood panel analyzer (lab_assay_01)', 'Sample collection kit', 'Oral thermometer', 'Surface disinfectant wipes'],
        ARRAY[
            'Verify analyzer calibration status and inspect sample handling log to rule out pre-analytical collection artifacts.',
            'Check core body temperature and screen the crew member for localized infection signs (respiratory, skin, urinary, or dental).',
            'Review cumulative radiation dose and recent closed-environment air/water microbial monitoring logs.',
            'Repeat the blood panel on a fresh sample if requested by the CMO to confirm the trend.',
            'Transmit confirmed WBC/RBC results, temperature, and symptom notes to the Flight Surgeon for longitudinal evaluation.'
        ],
        ARRAY[
            'Escalate promptly to the Flight Surgeon if abnormal WBC count is accompanied by fever, chills, or acute systemic illness.',
            'Follow strict biohazard containment and surface sanitation when handling blood samples in microgravity.'
        ],
        ARRAY['wbc_count', 'rbc_count', 'core_temp'],
        ARRAY['space_radiation', 'hostile_closed_env', 'distance_from_earth'],
        ARRAY['wbc', 'rbc', 'white blood cell', 'red blood cell', 'anemia', 'infection', 'immune', 'blood panel', 'lab'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        18
    ),
    (
        'ocular_sans_ultrasound',
        'Ocular / Optic Nerve Ultrasound & SANS Check',
        'diagnostics_imaging',
        'ocular_neuro',
        'Ultrasound imaging workflow to monitor Spaceflight Associated Neuro-ocular Syndrome (SANS) and fluid shifts.',
        'During scheduled ocular checks or when a crew member reports blurred near vision, eye pressure, or persistent headache with elevated CO2.',
        'cmo'::escalation_level,
        30,
        'iva'::environment_mode,
        ARRAY['Diagnostic ultrasound (ultrasound_01)', 'Sterile acoustic gel', 'Eye shield / probe cover', 'Visual acuity chart'],
        ARRAY[
            'Prepare the diagnostic ultrasound unit, verify sterile probe cover and acoustic gel, and enter crew member ID.',
            'Perform bilateral ocular and optic nerve sheath ultrasound sweeps gently through closed eyelids without applying globe pressure.',
            'Verify image clarity and capture still frames/clips per onboard imaging guide, recording findings text.',
            'Log recent cabin ppCO2 levels and conduct a brief visual acuity / Amsler grid self-check alongside the scan.',
            'Upload ultrasound media and findings into HRIG and transmit to the CMO and Flight Surgeon for comparison with baseline.'
        ],
        ARRAY[
            'Never press firmly on the closed eye with the ultrasound probe; float the probe on an adequate layer of acoustic gel.',
            'Escalate immediately to the Flight Surgeon for sudden vision loss, flashes/floaters, or severe eye pain.'
        ],
        ARRAY['ultrasound', 'cabin_ppco2'],
        ARRAY['gravity_fields', 'hostile_closed_env', 'distance_from_earth'],
        ARRAY['eye vision', 'ocular', 'sans', 'optic nerve', 'ultrasound', 'blurred vision', 'headache', 'fluid shift'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        19
    ),
    (
        'manual_vitals_guide',
        'How to Measure Manual Crew Vitals',
        'measurement_guide',
        'general',
        'Step-by-step operational guide for capturing accurate manual check-in metrics in microgravity or partial gravity.',
        'When performing daily manual check-ins, calibrating wearable readings, or logging spot measurements into HRIG.',
        'self_care'::escalation_level,
        15,
        'iva'::environment_mode,
        ARRAY['Blood pressure cuff', 'Blood glucose meter', 'Mass measurement device', 'HRIG manual check-in panel'],
        ARRAY[
            'Blood Pressure: Rest quietly for 5 minutes, wrap the cuff snugly on the bare upper arm at heart level, and record systolic/diastolic values.',
            'Blood Glucose: Clean and dry the fingertip thoroughly before lancing, apply blood cleanly to the test strip, and record fasting vs post-meal context in the note.',
            'Body Mass: Zero the microgravity mass measurement device (or partial-gravity scale) and remain still during the measurement cycle.',
            'Sleep & Mood: Log total restorative sleep hours over the past 24 hours and rate subjective mood honestly on the 1–5 scale.',
            'Review all entered values and units in HRIG before submitting so trend charts and alerts remain accurate.'
        ],
        ARRAY[
            'Do not take blood pressure immediately after resistive exercise or caffeine—wait at least 30 minutes when possible.',
            'Report malfunctioning cuffs, expired glucose strips, or inconsistent sensor readings to the CMO.'
        ],
        ARRAY['bp_systolic', 'bp_diastolic', 'blood_glucose', 'body_mass', 'mood', 'sleep_hours'],
        ARRAY['distance_from_earth', 'gravity_fields'],
        ARRAY['manual vitals', 'how to measure', 'blood pressure cuff', 'glucose strip', 'body mass', 'sleep log', 'mood scale', 'check-in'],
        'NASA HRP / ISS Operational Checklist (demo - clinical review required)',
        true,
        20
    )
ON CONFLICT (code) DO UPDATE
    SET title                = EXCLUDED.title,
        category             = EXCLUDED.category,
        body_system          = EXCLUDED.body_system,
        summary              = EXCLUDED.summary,
        when_to_use          = EXCLUDED.when_to_use,
        escalation           = EXCLUDED.escalation,
        estimated_minutes    = EXCLUDED.estimated_minutes,
        environment          = EXCLUDED.environment,
        equipment            = EXCLUDED.equipment,
        steps                = EXCLUDED.steps,
        warnings             = EXCLUDED.warnings,
        related_metric_codes = EXCLUDED.related_metric_codes,
        hazard_codes         = EXCLUDED.hazard_codes,
        keywords             = EXCLUDED.keywords,
        source               = EXCLUDED.source,
        is_active            = EXCLUDED.is_active,
        display_order        = EXCLUDED.display_order;

COMMIT;
