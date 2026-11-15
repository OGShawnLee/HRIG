// ============================================================================
// HRIG · Astronaut profiles and measurements
// Generated from docs/bd.md — names, types, cardinalities and constraints
// are preserved exactly as specified.
// Target engine: PostgreSQL 18
// ============================================================================

Enum metric_tier {
  primary
  secondary
}

Enum value_type {
  numeric
  scale
  text
  image
}

Enum environment_mode {
  iva
  eva
}

Enum device_kind {
  wearable
  environment_sensor
  manual
  lab_assay
  imaging
}

Enum media_kind {
  ultrasound
  ocular
}

Enum reading_status {
  normal
  abnormal
  critical
}

// ----------------------------------------------------------------------------
// 1. astronaut — crew profile
// ----------------------------------------------------------------------------
Table astronaut {
  id                uuid          [pk, default: `gen_random_uuid()`]
  full_name         text          [not null]
  birth_date        date          [not null]
  nationality_code  char(2)       [not null]
  nationality_label text
  is_active         boolean       [not null, default: true]
  created_at        timestamptz   [not null, default: `now()`]
  updated_at        timestamptz   [not null, default: `now()`]

  checks {
    `length(trim(full_name)) between 1 and 120` [name: 'astronaut_full_name_len']
    `birth_date <= current_date`                [name: 'astronaut_birth_not_future']
    `nationality_code ~ '^[A-Z]{2}$'`           [name: 'astronaut_nationality_iso']
  }

  Note: '''
  Profile: name, age, nationality, date of birth.
  `age` is DERIVED from `birth_date`, never stored (view `v_astronaut_age`).
  `nationality_label` is display-only; the UI falls back to `nationality_code` when NULL.
  `updated_at` is maintained by BEFORE UPDATE trigger `trg_astronaut_updated_at`.
  '''
}

// ----------------------------------------------------------------------------
// 2. unit — unit of measure
// ----------------------------------------------------------------------------
Table unit {
  id        uuid      [pk, default: `gen_random_uuid()`]
  code      text      [not null, unique]
  symbol    text      [not null]
  dimension text      [not null]
  decimals  smallint  [default: 1]
  is_active boolean   [not null, default: true]

  Note: '''
  `code` is a UCUM code: %, mm[Hg], kPa, /min, ms, Cel, uSv/h, mSv, mg/dL, kg, h, mL, 10*6/uL, 10*3/uL, {score}.
  `symbol` is the display text: mmHg, °C, bpm.
  `dimension` groups comparable units: fraction, pressure, temperature, rate, time,
  dose, dose_rate, concentration, mass, volume, count, score.
  Requirement example: "heart rate = 10" is stored as value_numeric = 10 with unit_id = bpm.
  The unit always travels with the value, so changing the unit a device reports never
  rewrites historical readings.
  '''
}

// ----------------------------------------------------------------------------
// 3. metric — measurement type catalog
// ----------------------------------------------------------------------------
Table metric {
  id                  uuid          [pk, default: `gen_random_uuid()`]
  code                text          [not null, unique]
  name                text          [not null]
  short_name          text
  category            text          [not null]
  tier                metric_tier   [not null]
  metric_type         smallint      [not null]
  value_type          value_type    [not null]
  base_unit_id        uuid          [ref: > unit.id]
  display_unit_symbol text
  scale_min           smallint
  scale_max           smallint
  display_order       smallint      [default: 0]
  is_active           boolean       [not null, default: true]

  checks {
    `metric_type in (1, 2)`                                                        [name: 'metric_type_domain']
    `value_type <> 'scale' or (scale_min is not null and scale_max is not null)`   [name: 'metric_scale_needs_range']
    `scale_min is null or scale_max is null or scale_min < scale_max`              [name: 'metric_scale_range_order']
    `value_type not in ('numeric', 'scale') or base_unit_id is not null`           [name: 'metric_numeric_needs_unit']
    `value_type not in ('text', 'image') or base_unit_id is null`                  [name: 'metric_text_no_unit']
  }

  Note: '''
  `metric_type`: 1 = astronaut measurement (physiological), 2 = environmental measurement
  (cabin or suit). measurement.astronaut_id must be NULL when metric_type = 2
  (cross-table rule, enforced by trigger trg_measurement_metric_type).
  `category`: oxygenation, pressure, radiation, cardiac, temperature, sleep_mood,
  metabolic, hydration_body, lab_imaging.
  `value_type` decides which column of measurement_value carries the data:
  numeric -> value_numeric, scale -> value_scale, text -> value_text,
  image -> value_text plus media_asset.
  `base_unit_id` is the canonical unit, present only for numeric and scale metrics.
  `display_unit_symbol` overrides unit.symbol in the HUD when one unit is shared by
  several metrics (/min is bpm for heart_rate but rpm for resp_rate). When NULL,
  unit.symbol is used.
  `scale_min` / `scale_max` apply only to value_type = 'scale' (mood 1-5).
  '''
}

// ----------------------------------------------------------------------------
// 4. device — device that originates the reading
// ----------------------------------------------------------------------------
Table device {
  id        uuid         [pk, default: `gen_random_uuid()`]
  code      text         [not null, unique]
  name      text         [not null]
  kind      device_kind  [not null]
  is_active boolean      [not null, default: true]

  Note: '''
  `code` examples: biomonitor_01, dosimeter_01, suit_bus_02.
  `kind`: wearable, environment_sensor, manual, lab_assay, imaging.
  '''
}

// ----------------------------------------------------------------------------
// 5. metric_threshold — configurable classification limits
// ----------------------------------------------------------------------------
Table metric_threshold {
  id           uuid          [pk, default: `gen_random_uuid()`]
  metric_id    uuid          [not null, unique, ref: > metric.id]
  warning_min  numeric(12,4)
  warning_max  numeric(12,4)
  critical_min numeric(12,4)
  critical_max numeric(12,4)
  classification_mode text [not null, default: 'numeric']
  source       text          [not null, default: 'demo']
  updated_at   timestamptz   [not null, default: `now()`]

  Note: '''
  Every catalog metric has a configuration. Boundaries use metric.base_unit_id.
  The server checks critical limits first;
  otherwise values outside warning limits are abnormal, and values within them are
  normal. A NULL boundary means no limit on that side. No row means unconfigured.
  Image metrics use manual mode because they have no numeric limits. Seed values
  are examples that require clinical review, not medical advice. Demo heart_rate
  has critical_max = 200 bpm.
  '''
}

// ----------------------------------------------------------------------------
// 6. measurement — capture event or session
// ----------------------------------------------------------------------------
Table measurement {
  id              uuid              [pk, default: `gen_random_uuid()`]
  astronaut_id    uuid              [ref: > astronaut.id]
  device_id       uuid              [ref: > device.id]
  environment     environment_mode  [not null]
  recorded_at_utc timestamptz       [not null]
  note            text

  indexes {
    (astronaut_id, recorded_at_utc) [note: 'recorded_at_utc DESC']
    (device_id, recorded_at_utc)    [note: 'recorded_at_utc DESC']
  }

  checks {
    `num_nonnulls(astronaut_id, device_id) >= 1` [name: 'measurement_not_orphan']
  }

  Note: '''
  One session groups every value captured together: a blood pressure reading is one
  measurement with two measurement_value rows.
  `astronaut_id` is NULL for cabin/suit readings (O2, pressure, ambient temperature);
  `device_id` is NULL when the data is entered manually.
  `note` holds context such as post-EVA or resting.
  '''
}

// ----------------------------------------------------------------------------
// 7. measurement_value — individual reading value
// ----------------------------------------------------------------------------
Table measurement_value {
  id              uuid          [pk, default: `gen_random_uuid()`]
  measurement_id  uuid          [not null, ref: > measurement.id]
  metric_id       uuid          [not null, ref: > metric.id]
  unit_id         uuid          [ref: > unit.id]
  recorded_at_utc timestamptz   [not null]
  value_numeric   numeric(12,4)
  value_scale     smallint
  value_text      text
  status          reading_status [not null, default: 'normal']
  note            text

  indexes {
    (measurement_id, metric_id)  [unique]
    (metric_id, recorded_at_utc) [note: 'recorded_at_utc DESC — latest value per metric and 24-hour trends']
  }

  checks {
    `num_nonnulls(value_numeric, value_scale, value_text) >= 1` [name: 'measurement_value_has_value']
    `value_numeric is null or unit_id is not null`              [name: 'measurement_value_numeric_needs_unit']
  }

  Note: '''
  `unit_id` is the ACTUAL unit this reading arrived in, and it is required whenever
  value_numeric is present. It stays NULL only for text and image metrics.
  `value_scale` holds subjective scales (mood 1-5); `value_text` holds findings and
  interpretations.
  `status` stores the server classification for this reading. The server consults
  metric_threshold, checking critical limits before warning limits.
  `recorded_at_utc` is the ONLY denormalized field in the schema. It always mirrors
  measurement.recorded_at_utc (trigger trg_measurement_value_timestamp) and exists so
  the HUD can answer "latest value per metric" with a single index scan.
  `unit_id` must share `dimension` with metric.base_unit_id (trigger
  trg_measurement_value_dimension, since a CHECK cannot read other tables).
  ON DELETE RESTRICT: a session with values cannot be deleted.
  '''
}

// ----------------------------------------------------------------------------
// 8. media_asset — ultrasound image
// ----------------------------------------------------------------------------
Table media_asset {
  id              uuid        [pk, default: `gen_random_uuid()`]
  measurement_id  uuid        [not null, ref: > measurement.id]
  kind            media_kind  [not null]
  file_path       text        [not null]
  mime_type       text        [not null]
  findings        text
  captured_at_utc timestamptz

  checks {
    `mime_type in ('image/jpeg', 'image/png', 'application/pdf')` [name: 'media_asset_mime']
  }

  Note: '''
  `file_path` is relative to the project root, e.g.
  media/ultrasound/2026-11-15_crew01.jpg.
  `captured_at_utc` may be NULL; when missing, the UI uses measurement.recorded_at_utc.
  '''
}

// ----------------------------------------------------------------------------
// Explicit relationships (cardinality per docs/bd.md)
// ----------------------------------------------------------------------------
Ref: measurement.astronaut_id          > astronaut.id        // astronaut 1 : N measurement, optional
Ref: measurement.device_id             > device.id           // device 1 : N measurement, optional
Ref: measurement_value.measurement_id  > measurement.id      [delete: restrict] // 1 : N
Ref: measurement_value.metric_id       > metric.id           // metric 1 : N measurement_value
Ref: metric.base_unit_id               > unit.id             // metric N : 1 unit, optional (canonical unit)
Ref: measurement_value.unit_id         > unit.id             // unit 1 : N measurement_value (reading unit)
Ref: media_asset.measurement_id        > measurement.id      // measurement 1 : N media_asset, optional
