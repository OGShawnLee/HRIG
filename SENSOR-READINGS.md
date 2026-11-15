[
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:00:00Z",
    "astronaut": { "full_name": "Alex Morgan" },
    "readings": [
      { "metric_code": "heart_rate", "value_numeric": 72, "unit_code": "/min" },
      { "metric_code": "body_temperature", "value_numeric": 36.8, "unit_code": "Cel" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:05:00Z",
    "astronaut": { "full_name": "Alex Morgan" },
    "readings": [
      { "metric_code": "spo2", "value_numeric": 98, "unit_code": "%" },
      { "metric_code": "resp_rate", "value_numeric": 16, "unit_code": "/min" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:10:00Z",
    "astronaut": { "full_name": "Alex Morgan" },
    "readings": [
      { "metric_code": "systolic_bp", "value_numeric": 118, "unit_code": "mm[Hg]" },
      { "metric_code": "diastolic_bp", "value_numeric": 76, "unit_code": "mm[Hg]" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:15:00Z",
    "astronaut": { "full_name": "Alex Morgan" },
    "readings": [
      { "metric_code": "heart_rate", "value_numeric": 104, "unit_code": "/min" },
      { "metric_code": "resp_rate", "value_numeric": 21, "unit_code": "/min" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:20:00Z",
    "astronaut": { "full_name": "Alex Morgan" },
    "readings": [
      { "metric_code": "mood", "value_scale": 4, "unit_code": "{score}" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:00:00Z",
    "astronaut": { "full_name": "Samira Patel" },
    "readings": [
      { "metric_code": "heart_rate", "value_numeric": 64, "unit_code": "/min" },
      { "metric_code": "body_temperature", "value_numeric": 36.5, "unit_code": "Cel" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:05:00Z",
    "astronaut": { "full_name": "Samira Patel" },
    "readings": [
      { "metric_code": "spo2", "value_numeric": 97, "unit_code": "%" },
      { "metric_code": "resp_rate", "value_numeric": 14, "unit_code": "/min" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:10:00Z",
    "astronaut": { "full_name": "Samira Patel" },
    "readings": [
      { "metric_code": "systolic_bp", "value_numeric": 112, "unit_code": "mm[Hg]" },
      { "metric_code": "diastolic_bp", "value_numeric": 72, "unit_code": "mm[Hg]" }
    ]
  },
  {
    "device_code": "biomonitor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:15:00Z",
    "astronaut": { "full_name": "Samira Patel" },
    "readings": [
      { "metric_code": "heart_rate", "value_numeric": 88, "unit_code": "/min" },
      { "metric_code": "hydration", "value_numeric": 2.1, "unit_code": "L" }
    ]
  },
  {
    "device_code": "cabin_sensor_01",
    "environment": "iva",
    "recorded_at_utc": "2026-11-15T08:20:00Z",
    "readings": [
      { "metric_code": "cabin_o2", "value_numeric": 21.0, "unit_code": "%" },
      { "metric_code": "cabin_pressure", "value_numeric": 101.3, "unit_code": "kPa" },
      { "metric_code": "cabin_temperature", "value_numeric": 22.4, "unit_code": "Cel" }
    ]
  }
]

Notas para que encajen con el esquema actual:

- Los códigos de dispositivos, métricas y unidades deben existir en los catálogos de la base, cargados desde 002_seed.sql.
- Alex Morgan y Samira Patel son nombres de ejemplo: sus perfiles deben existir primero en astronaut. El servidor debe buscar su id y usarlo como astronaut_id.
- En la métrica ambiental de la nave, astronaut se omite. Así se cumple la regla de que métricas ambientales (metric_type = 2) no estén asociadas a astronautas.
- El esquema guarda un valor por métrica dentro de cada sesión; por eso cada metric_code aparece una sola vez en cada objeto de sesión.
- status puede omitirse: la base lo establece como normal. Para clasificar lecturas según límites, el servidor tendría que calcularlo antes de guardar.
- Los nombres de métricas y unidades de estos ejemplos deben verificarse con el catálogo real de 002_seed.sql antes de probarlos; el JSON es ilustrativo y el esquema SQL no recibe JSON directamente.
