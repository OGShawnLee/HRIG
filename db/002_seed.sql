-- ============================================================================
-- HRIG · Astronaut profiles and measurements
-- 002_seed.sql — idempotent catalogs from docs/bd.md
-- Re-running this file never duplicates rows.
-- ============================================================================

BEGIN;

INSERT INTO unit (code, symbol, dimension, decimals) VALUES
    ('%',        '%',      'fraction',      0),
    ('mm[Hg]',   'mmHg',   'pressure',      0),
    ('kPa',      'kPa',    'pressure',      2),
    ('/min',     'rpm',    'rate',          0),
    ('ms',       'ms',     'time',          0),
    ('Cel',      '°C',     'temperature',   1),
    ('uSv/h',    'µSv/h',  'dose_rate',     2),
    ('mSv',      'mSv',    'dose',          2),
    ('mg/dL',    'mg/dL',  'concentration', 0),
    ('kg',       'kg',     'mass',          1),
    ('h',        'h',      'time',          1),
    ('mL',       'mL',     'volume',        0),
    ('10*6/uL',  '10⁶/µL', 'count',         1),
    ('10*3/uL',  '10³/µL', 'count',         1),
    ('{score}',  'score',  'score',         0)
ON CONFLICT (code) DO UPDATE
    SET symbol    = EXCLUDED.symbol,
        dimension = EXCLUDED.dimension,
        decimals  = EXCLUDED.decimals;

INSERT INTO metric (
    code, name, category, tier, metric_type, value_type,
    base_unit_id, display_unit_symbol, scale_min, scale_max, display_order
)
SELECT v.code, v.name, v.category, v.tier::metric_tier, v.metric_type, v.value_type::value_type,
       (SELECT u.id FROM unit u WHERE u.code = v.unit_code),
       v.display_unit_symbol, v.scale_min, v.scale_max, v.display_order
FROM (VALUES
    -- Primary · metric_type 1 = astronaut, 2 = environmental
    ('spo2',               'Blood oxygen saturation', 'oxygenation',    'primary',   1, 'numeric', '%',        '%',     NULL::smallint, NULL::smallint,  1),
    ('cabin_o2',           'Cabin O2',                'oxygenation',    'primary',   2, 'numeric', '%',        '%',     NULL, NULL,  2),
    ('cabin_ppco2',        'Cabin ppCO2',             'oxygenation',    'primary',   2, 'numeric', 'mm[Hg]',   'mmHg',  NULL, NULL,  3),
    ('resp_rate',          'Respiratory rate',        'oxygenation',    'primary',   1, 'numeric', '/min',     'rpm',   NULL, NULL,  4),
    ('cabin_pressure',     'Cabin pressure',          'pressure',       'primary',   2, 'numeric', 'kPa',      'kPa',   NULL, NULL,  5),
    ('suit_pressure',      'Suit pressure',           'pressure',       'primary',   2, 'numeric', 'kPa',      'kPa',   NULL, NULL,  6),
    ('bp_systolic',        'Systolic blood pressure', 'pressure',       'primary',   1, 'numeric', 'mm[Hg]',   'mmHg',  NULL, NULL,  7),
    ('bp_diastolic',       'Diastolic blood pressure','pressure',       'primary',   1, 'numeric', 'mm[Hg]',   'mmHg',  NULL, NULL,  8),
    ('radiation_rate',     'Exposure rate',           'radiation',      'primary',   1, 'numeric', 'uSv/h',    'µSv/h', NULL, NULL,  9),
    ('radiation_dose_cum', 'Cumulative dose',         'radiation',      'primary',   1, 'numeric', 'mSv',      'mSv',   NULL, NULL, 10),
    ('heart_rate',         'Heart rate',              'cardiac',        'primary',   1, 'numeric', '/min',     'bpm',   NULL, NULL, 11),
    ('hrv_rmssd',          'Heart rate variability',  'cardiac',        'primary',   1, 'numeric', 'ms',       'ms',    NULL, NULL, 12),
    ('core_temp',          'Core body temperature',   'temperature',    'primary',   1, 'numeric', 'Cel',      '°C',    NULL, NULL, 13),
    ('cabin_temp',         'Cabin temperature',       'temperature',    'primary',   2, 'numeric', 'Cel',      '°C',    NULL, NULL, 14),
    ('suit_temp',          'Suit temperature',        'temperature',    'primary',   2, 'numeric', 'Cel',      '°C',    NULL, NULL, 15),
    -- Secondary
    ('sleep_hours',        'Sleep hours',             'sleep_mood',     'secondary', 1, 'numeric', 'h',        'h',     NULL, NULL, 16),
    ('mood',               'Mood',                    'sleep_mood',     'secondary', 1, 'scale',   '{score}',  NULL,    1,    5,    17),
    ('blood_glucose',      'Blood glucose',           'metabolic',      'secondary', 1, 'numeric', 'mg/dL',    'mg/dL', NULL, NULL, 18),
    ('hydration_intake',   'Fluid intake',            'hydration_body', 'secondary', 1, 'numeric', 'mL',       'mL',    NULL, NULL, 19),
    ('body_mass',          'Body mass',               'hydration_body', 'secondary', 1, 'numeric', 'kg',       'kg',    NULL, NULL, 20),
    ('rbc_count',          'Red blood cell count',    'lab_imaging',    'secondary', 1, 'numeric', '10*6/uL',  '10⁶/µL',NULL, NULL, 21),
    ('wbc_count',          'White blood cell count',  'lab_imaging',    'secondary', 1, 'numeric', '10*3/uL',  '10³/µL',NULL, NULL, 22),
    ('ultrasound',         'Diagnostic ultrasound',   'lab_imaging',    'secondary', 1, 'image',    NULL,      NULL,    NULL, NULL, 23)
) AS v(code, name, category, tier, metric_type, value_type,
       unit_code, display_unit_symbol, scale_min, scale_max, display_order)
ON CONFLICT (code) DO UPDATE
    SET name                = EXCLUDED.name,
        category            = EXCLUDED.category,
        tier                = EXCLUDED.tier,
        metric_type         = EXCLUDED.metric_type,
        value_type          = EXCLUDED.value_type,
        base_unit_id        = EXCLUDED.base_unit_id,
        display_unit_symbol = EXCLUDED.display_unit_symbol,
        scale_min           = EXCLUDED.scale_min,
        scale_max           = EXCLUDED.scale_max,
        display_order       = EXCLUDED.display_order;

INSERT INTO device (code, name, kind) VALUES
    ('biomonitor_01',   'IVA biomonitor',           'wearable'),
    ('baumanometro_01', 'Blood pressure cuff',      'wearable'),
    ('dosimeter_01',    'Radiation dosimeter',      'wearable'),
    ('cabin_sensor_01', 'Cabin environment sensor', 'environment_sensor'),
    ('lab_assay_01',    'Blood panel analyzer',     'lab_assay'),
    ('ultrasound_01',   'Diagnostic ultrasound',    'imaging')
ON CONFLICT (code) DO UPDATE
    SET name = EXCLUDED.name,
        kind = EXCLUDED.kind;

-- Mockup limits in each metric's canonical unit, informed by common adult
-- resting reference ranges where applicable. Context-specific spacecraft,
-- EVA, and individual limits are illustrative and require expert review.
INSERT INTO metric_threshold (metric_id, warning_min, warning_max, critical_min, critical_max, source)
SELECT m.id, v.warning_min, v.warning_max, v.critical_min, v.critical_max, 'demo - clinical review required'
FROM (VALUES
    ('heart_rate',        50::numeric, 100::numeric, 40::numeric, 200::numeric),
    ('spo2',              94,          NULL,         90,          NULL),
    ('resp_rate',         10,          20,           6,           30),
    ('bp_systolic',       90,          120,          70,          180),
    ('bp_diastolic',      60,          80,           40,          120),
    ('core_temp',         36.5,        37.3,         35,          40),
    ('cabin_o2',          19.5,        23.5,         17,          25),
    ('cabin_ppco2',       NULL,        3,            NULL,        5),
    ('cabin_pressure',    97,          104,          90,          110),
    ('cabin_temp',        18,          27,           10,          35),
    ('suit_pressure',     28,          30,           25,          33),
    ('suit_temp',         15,          35,           5,           45),
    ('hrv_rmssd',         20,          NULL,         10,          NULL),
    ('radiation_rate',    NULL,        0.5,          NULL,        2),
    ('radiation_dose_cum',NULL,        20,           NULL,        50),
    ('sleep_hours',       7,           9,            4,           14),
    ('mood',              2,           4,            1,           5),
    ('blood_glucose',     70,          140,          54,          250),
    ('hydration_intake',  1500,        3500,         500,         5000),
    ('body_mass',         45,          110,          35,          140),
    ('rbc_count',         4,           6,            2.5,         8),
    ('wbc_count',         4,           11,           1,           30)
) AS v(code, warning_min, warning_max, critical_min, critical_max)
JOIN metric m ON m.code = v.code
ON CONFLICT (metric_id) DO UPDATE SET
    warning_min = EXCLUDED.warning_min,
    warning_max = EXCLUDED.warning_max,
    critical_min = EXCLUDED.critical_min,
    critical_max = EXCLUDED.critical_max,
    classification_mode = 'numeric',
    source = EXCLUDED.source,
    updated_at = now();

-- Image studies are reviewed by a person; they do not have numeric limits.
INSERT INTO metric_threshold (metric_id, classification_mode, source)
SELECT id, 'manual', 'manual image review - demo configuration'
FROM metric WHERE code = 'ultrasound'
ON CONFLICT (metric_id) DO UPDATE SET
    warning_min = NULL, warning_max = NULL, critical_min = NULL, critical_max = NULL,
    classification_mode = 'manual', source = EXCLUDED.source, updated_at = now();

COMMIT;
