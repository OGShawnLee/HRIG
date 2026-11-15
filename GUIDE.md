# Guía del servidor receptor de sensores

## Objetivo

El servidor tiene una sola responsabilidad: recibir por HTTP los JSON que envía
el simulador de `bot`, validar que correspondan al catálogo y guardarlos en
PostgreSQL. No calcula gráficas ni reenvía la información a otros servicios.

El servidor debe aceptar `POST` con `Content-Type: application/json`. La URL
actual del bot es de ejemplo (`http://localhost:80/`); puede configurarse con
`SENSOR_ENDPOINT` cuando se defina la ruta real.

La conexión a PostgreSQL también debe poder configurarse. Como ejemplo, el
servidor puede leer `DATABASE_URL` con un valor ilustrativo como
`postgres://usuario:clave@localhost:5432/hrig?sslmode=disable`. No es una
conexión real: se reemplazarán host, base, usuario, contraseña y política TLS
cuando se elija el entorno. Las credenciales reales deben mantenerse fuera del
código y del repositorio.

## Formato recibido

Cada petición representa una sesión de un dispositivo y contiene una o más
lecturas. Este ejemplo corresponde al formato que actualmente produce el bot:

```json
{
  "device_code": "cabin_sensor_01",
  "environment": "iva",
  "recorded_at_utc": "2026-11-15T08:20:00Z",
  "readings": [
    {
      "metric_code": "cabin_pressure",
      "value_numeric": 101.3,
      "unit_code": "kPa"
    }
  ]
}
```

Una lectura personal agrega `astronaut_name`, por ejemplo `"Ana Torres"`.
Las sesiones ambientales de cabina y traje no incluyen ese campo. Una lectura
numérica usa `value_numeric`; una escala como `mood` usa `value_scale`; y la
lectura demostrativa `ultrasound` usa `value_text`. Solo se debe enviar el campo
de valor que corresponde a `metric.value_type`. `unit_code` se omite cuando la
métrica no tiene unidad, como `ultrasound`.

## Validaciones antes de escribir

1. Limitar el tamaño del cuerpo, exigir JSON válido y rechazar campos o valores
   obligatorios ausentes. `readings` debe contener al menos una lectura.
2. Validar que `environment` sea `iva` o `eva` y que `recorded_at_utc` incluya
   zona horaria. Guardar el instante como `timestamptz`.
3. Resolver `device_code` contra `device.code` y comprobar que el dispositivo
   exista y esté activo.
4. Si llega `astronaut_name`, resolverlo contra `astronaut.full_name` y
   comprobar que haya una única coincidencia activa. Si no llega, guardar
   `astronaut_id = NULL`. El esquema exige que al menos `astronaut_id` o
   `device_id` tenga valor; el dispositivo resuelto siempre satisface esto.
5. Por cada lectura, resolver `metric_code` contra `metric.code` y comprobar
   que exista y esté activa. `metric_type = 2` es ambiental y no puede tener
   astronauta asociado; `metric_type = 1` es una medición de astronauta.
6. Verificar que el campo de valor corresponda a `value_type`: `numeric` usa
   `value_numeric`, `scale` usa `value_scale`, y `text` o `image` usan
   `value_text`. Rechazar valores múltiples o ninguno.
7. Para valores numéricos y de escala, resolver `unit_code` contra `unit.code`.
   Confirmar que la dimensión de la unidad coincide con la unidad base de la
   métrica. Para texto e imagen, la métrica no debe tener unidad.
8. Para `scale`, respetar `metric.scale_min` y `metric.scale_max`. Para
   numéricos, aceptar hasta la precisión que admite `numeric(12,4)`.
9. No confiar en `status` recibido: el formato actual no lo manda y la columna
   tiene valor predeterminado `normal`. La clasificación según
   `metric_threshold` puede añadirse aquí si se requiere; no es necesaria para
   insertar las lecturas.

La base también aplica restricciones y triggers para validar unidades, evitar
que una métrica ambiental quede asociada a un astronauta y copiar la fecha de la
sesión a sus lecturas. El servidor debe manejar errores de esas restricciones
sin dejar sesiones parciales.

## Escritura en PostgreSQL

Guardar cada petición en una sola transacción:

1. Resolver el dispositivo y, si aplica, el astronauta.
2. Insertar una fila en `measurement` con `astronaut_id`, `device_id`,
   `environment` y `recorded_at_utc`.
3. Insertar una fila por lectura en `measurement_value`, con el ID de sesión,
   IDs de métrica y unidad, y exactamente una columna de valor.
4. Confirmar la transacción solo si todas las lecturas se insertaron. Ante
   cualquier error, hacer rollback.

La aplicación no debe enviar UUID generados por el cliente: PostgreSQL genera
los IDs de `measurement` y `measurement_value`. Las consultas deben usar
parámetros, no concatenar valores del JSON dentro de SQL.

Flujo ilustrativo en pseudocódigo (los marcadores `:valor` son parámetros del
driver PostgreSQL):

```text
db = conectar(DATABASE_URL)  # configuración de ejemplo, sustituible
tx = db.begin()

device_id = SELECT id FROM device
            WHERE code = :device_code AND is_active = true
astronaut_id = NULL, o SELECT id FROM astronaut
                WHERE full_name = :astronaut_name AND is_active = true

measurement_id = INSERT INTO measurement
  (astronaut_id, device_id, environment, recorded_at_utc)
  VALUES (:astronaut_id, :device_id, :environment, :recorded_at_utc)
  RETURNING id

por cada lectura:
  buscar metric.id y unit.id por sus códigos
  INSERT INTO measurement_value
    (measurement_id, metric_id, unit_id, recorded_at_utc,
     value_numeric, value_scale, value_text)
    VALUES (:measurement_id, :metric_id, :unit_id, :recorded_at_utc,
            :value_numeric, :value_scale, :value_text)

tx.commit()  # solo si todas las búsquedas e inserciones funcionaron
```

En el servidor real, `DATABASE_URL` se lee del entorno al iniciar, se reutiliza
un pool de conexiones y las búsquedas e inserciones de cada petición ocurren en
la misma transacción. Ante cualquier error se revierte la transacción.

Esquema y catálogos a cargar antes de recibir mensajes:

```sh
psql -d hrig -f db/001_schema.sql
psql -d hrig -f db/002_seed.sql
psql -d hrig -f db/003_demo_seed.sql
```

`003_demo_seed.sql` crea los tres astronautas y agrega `suit_sensor_01`, que
usa el simulador. `cabin_sensor_01`, los otros dispositivos, métricas y unidades
se declaran en `002_seed.sql`.

## Respuesta HTTP

- `201 Created`: sesión y lecturas guardadas.
- `400 Bad Request`: JSON inválido o estructura/campos incorrectos.
- `404 Not Found`: código de dispositivo, astronauta, métrica o unidad sin
  correspondencia en los catálogos.
- `409 Conflict`: lectura repetida u otra colisión de integridad.
- `500 Internal Server Error`: fallo interno o de base de datos. No devolver
  detalles SQL ni credenciales al emisor.

Responder solo después de confirmar la transacción. El bot considera cualquier
código HTTP fuera de `2xx` como fallo de envío. Como el formato actual no incluye
un identificador único de mensaje, una retransmisión tras un timeout podría
crear mediciones duplicadas; para reintentos seguros conviene añadir
`message_id` al contrato y una restricción única persistente.

## Alcance y limitaciones actuales

- Las lecturas se envían cada dos segundos; el endpoint debe soportar ese ritmo
  de solicitudes y responder pronto.
- El nombre del astronauta es adecuado para estos datos demo, pero no es una
  clave estable. Para producción, conviene que el emisor mande un código o UUID
  inequívoco.
- El bot envía un texto demostrativo para `ultrasound`. El esquema también
  dispone de `media_asset`, pero actualmente el JSON no transporta archivo,
  ruta ni MIME type; por lo tanto, el receptor no puede guardar una imagen real
  a partir de ese mensaje.
- La ruta configurada en el bot usa HTTP local de ejemplo. Si el servidor se
  expone fuera de la máquina de desarrollo, se debe configurar HTTPS y
  autenticación entre emisor y receptor.

## Criterios para considerar listo el receptor

- Un JSON válido crea una fila en `measurement` y una por cada lectura en
  `measurement_value`.
- Un paquete de cabina crea la sesión sin astronauta.
- Una métrica o unidad desconocida se rechaza sin guardar ninguna parte del
  paquete.
- Las métricas personales quedan asociadas al astronauta resuelto.
- Cada error revierte toda la transacción y devuelve un código HTTP acorde.
