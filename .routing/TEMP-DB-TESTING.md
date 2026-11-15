# Plan de pruebas de la base de datos HRIG

Documento de handoff para el agente que ejecute las pruebas cuando PostgreSQL esté disponible. **No se han ejecutado estas pruebas.** Ejecutarlas únicamente en una base descartable de pruebas, nunca en una base con datos que se quieran conservar.

## Preparación

Desde la raíz del repositorio, crear una base vacía de pruebas (por ejemplo, `hrig_test`) y cargar los tres scripts en orden:

```bash
createdb hrig_test
psql -v ON_ERROR_STOP=1 -d hrig_test -f db/001_schema.sql
psql -v ON_ERROR_STOP=1 -d hrig_test -f db/002_seed.sql
psql -v ON_ERROR_STOP=1 -d hrig_test -f db/003_demo_seed.sql
```

Si la base de pruebas ya existe y contiene objetos, detenerse: no borrar ni recrear nada sin confirmar que es descartable. `001_schema.sql` crea el esquema y no es una migración para bases existentes.

## Pruebas de carga y repetición

1. Ejecutar los tres comandos de preparación. Los tres deben terminar con código de salida `0`.
2. Volver a ejecutar `002_seed.sql` y `003_demo_seed.sql`. Deben terminar correctamente y no duplicar los catálogos ni las filas de demostración.
3. Verificar los conteos:

```sql
SELECT count(*) AS astronauts FROM astronaut;                 -- esperado: 3
SELECT count(*) AS metrics FROM metric;                        -- esperado: 23
SELECT count(*) AS threshold_configs FROM metric_threshold;    -- esperado: 23
SELECT count(*) AS demo_measurements FROM measurement;         -- esperado: 15
SELECT count(*) AS demo_values FROM measurement_value;         -- esperado: 117
```

Si se reutiliza la base y ya contiene otros datos, estos conteos absolutos no aplican; verificar los IDs/códigos demo en su lugar.

## Cobertura de rangos

Las 22 métricas numéricas/de escala deben tener configuración numérica, y `ultrasound` debe tener configuración manual:

```sql
SELECT m.code, m.value_type, t.classification_mode
FROM metric m
LEFT JOIN metric_threshold t ON t.metric_id = m.id
ORDER BY m.display_order;
```

Esperado: cero filas con `classification_mode IS NULL`; `ultrasound` en modo `manual`; las restantes en modo `numeric`.

Comprobar que no existen límites invertidos ni filas numéricas sin límites:

```sql
SELECT m.code
FROM metric_threshold t
JOIN metric m ON m.id = t.metric_id
WHERE (t.classification_mode = 'numeric' AND
       t.warning_min IS NULL AND t.warning_max IS NULL AND
       t.critical_min IS NULL AND t.critical_max IS NULL)
   OR (t.warning_min IS NOT NULL AND t.warning_max IS NOT NULL AND t.warning_min > t.warning_max)
   OR (t.critical_min IS NOT NULL AND t.warning_min IS NOT NULL AND t.critical_min >= t.warning_min)
   OR (t.critical_max IS NOT NULL AND t.warning_max IS NOT NULL AND t.critical_max <= t.warning_max);
```

Esperado: cero filas.

## Integridad referencial y triggers

Probar cada caso en una transacción con `ROLLBACK`, usando IDs obtenidos por código y una sesión temporal de prueba. Registrar el error PostgreSQL esperado:

- Un `measurement` sin astronauta y sin dispositivo debe fallar por `measurement_not_orphan`.
- Un valor `cabin_o2` ligado a una sesión que tenga astronauta debe fallar por `trg_measurement_metric_type`.
- Una lectura `heart_rate` con una unidad de masa (`kg`) debe fallar por `trg_measurement_value_dimension`.
- Dos valores de `heart_rate` en la misma sesión deben fallar por la restricción única `(measurement_id, metric_id)`.
- Insertar un valor numérico sin `unit_id` debe fallar por `measurement_value_numeric_needs_unit`.
- Insertar el valor sin `value_numeric`, `value_scale` ni `value_text` debe fallar por `measurement_value_has_value`.
- Una unidad incompatible, un UUID inexistente o un `metric_id` inexistente debe fallar por FK/trigger correspondiente.
  > **Nota (para Ticket `005`):** Probar explícitamente UUIDs inexistentes por separado en `astronaut_id`, `device_id`, `metric_id` y `unit_id` (error FK), además de `media_asset_mime` (aceptar `image/jpeg`/`image/png`/`application/pdf`, rechazar `video/mp4`), `v_astronaut_age`, `trg_astronaut_updated_at` y `ON DELETE RESTRICT` al intentar borrar un `measurement` con valores.

Verificar sincronización temporal: insertar un valor en una sesión temporal con una hora conocida y comprobar que `measurement_value.recorded_at_utc` toma la hora de la sesión; actualizar luego la hora de la sesión y comprobar que la hora del valor se actualiza también.

## Lectura de los límites para el servidor

Comprobar la configuración de frecuencia cardiaca:

```sql
SELECT m.code, u.symbol, t.warning_min, t.warning_max,
       t.critical_min, t.critical_max, t.classification_mode
FROM metric_threshold t
JOIN metric m ON m.id = t.metric_id
LEFT JOIN unit u ON u.id = m.base_unit_id
WHERE m.code = 'heart_rate';
```

Esperado: unidad `/min` (presentada como `bpm` por `display_unit_symbol`), aviso 50–100, crítico inferior 40 y superior 200. Como `measurement_value.status` lo asigna el servidor, probar allí la lógica de clasificación en bordes: 60 normal, 110 anómalo, 200 crítico. Confirmar con el equipo de producto si los límites inclusivos/exclusivos coinciden con la implementación del servidor.

> **Nota (confirmado con `.routing/BD.md` y Ticket `004`/`005`):** El servidor evalúa límites críticos de forma inclusiva (`<= critical_min` o `>= critical_max` → `critical`) antes que el rango de aviso (`< warning_min` o `> warning_max` → `abnormal`; valores en `[50, 100]` → `normal`). Cubrir explícitamente `GetThresholdByMetricCode("heart_rate")` y los valores de borde `40, 41, 50, 60, 100, 101, 110, 199, 200` en `postgres_test.go`.

## Informe de resultados

El agente debe reportar versión de PostgreSQL, comandos ejecutados, resultado de cada caso (PASS/FAIL), errores encontrados y si usó una base nueva o una reutilizada. No modificar rangos ni código durante la prueba sin registrar el cambio por separado.

## Limpieza

Solo si se creó expresamente la base descartable `hrig_test` y ya se guardó el informe:

```bash
dropdb hrig_test
```

Si se usó otro nombre o no se puede confirmar que sea descartable, no ejecutar la limpieza automática.
