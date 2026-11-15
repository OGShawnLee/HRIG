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

COMMIT;
