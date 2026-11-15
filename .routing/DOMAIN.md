# Glosario de Dominio (Ubiquitous Language)

Definiciones oficiales del dominio para **HRIG**, alineadas con [`METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md), [`DESIGN.md`](DESIGN.md) y los 5 peligros de [`INPUT.md`](INPUT.md).

## Métricas y Fuentes de Datos

- **Datos Primarios:** Indicadores de supervivencia inmediata y soporte vital visibles en el dashboard principal:
  - Oxigenación (`SpO2`, `O2` / `ppCO2` ambiental y frecuencia respiratoria).
  - Presión (arterial y barométrica de cabina/traje).
  - Radiación (tasa de exposición y dosis acumulada).
  - Ritmo cardíaco (frecuencia cardíaca y HRV).
  - Temperatura (corporal y ambiental/traje).
- **Datos Secundarios:** Indicadores preventivos y de adaptación evaluados mediante tendencias temporales y chequeos programados:
  - Calidad del sueño, niveles de cortisol y estado de ánimo.
  - Glucosa en sangre (`mg/dL`).
  - Hidratación, consumo calórico, masa corporal y salud musculoesquelética.
  - Conteo de glóbulos rojos/blancos e imágenes diagnósticas (ultrasonido / ocular).
- **Métricas Excluidas:** Datos estáticos o de especialidad sin acción autónoma en vuelo (altura, perfil tiroideo `TSH`/`T3`/`T4`), omitidos de la interfaz principal para evitar sobrecarga cognitiva.
- **Fuentes de Captura Multimodal:** Mecanismos de ingreso de datos al sistema:
  - **Equipos de Monitoreo de Salud:** Wearables IVA (actigrafía, biomonitor), bus cableado de traje EVA, baumanómetro, pulsioxímetro, dosímetro de radiación y equipos diagnósticos.
  - **Entrada Manual:** Formularios rápidos (teclado, clics o deslizadores para ánimo, sueño, glucosa o presión).
  - **Captura de Imágenes:** Carga o captura mediante cámara/ultrasonido para autoevaluación visual.

## Estados de Valor y Alertamiento

- **Valor Normal (Nominal):** Lectura dentro del rango esperado para el tripulante según su **Entorno de Misión**.
- **Aviso (Advisory):** Desviación temprana o recordatorio de chequeo programado; se muestra de forma pasiva sin interrumpir la operación.
- **Valor Anómalo (Caution / Precaución):** Lectura fuera del rango óptimo que requiere acciones de normalización, **Autocuidado** o revisión del **CMO**, sin riesgo vital inmediato.
- **Valor Crítico (Warning / Crítico):** Umbral superado con riesgo de pérdida de consciencia, deterioro cognitivo/motor o amenaza vital; exige acción inmediata.
- **Dato Obsoleto (Stale Data):** Lectura de un sensor desconectado o chequeo vencido. Debe marcarse explícitamente como `STALE / NO DATA` con su última marca de tiempo y nunca mostrarse como en vivo.
- **Alerta:** Notificación visual y sonora (siempre acompañada de ícono y etiqueta de texto) dirigida al astronauta o a toda la **Tripulación** ante un **Valor Crítico** en el entorno o en algún tripulante.

## Actores, Registros y Contexto Operativo

- **Tripulación:** Conjunto de astronautas que participan en una misión espacial y comparten hábitat, nave o estación.
- **Perfil de Tripulante:** Ficha individual de cada miembro de la **Tripulación** con su rol (tripulante u **Oficial Médico de la Tripulación [CMO]**), rangos base personales, preferencias de privacidad e historial propio.
- **Historia Clínica:** Registro cronológico e inmutable de mediciones, imágenes, **Sintomatología**, **Alertas** y **Procedimientos** aplicados, etiquetado con dispositivo de origen, perfil y marca de tiempo dual (`MET` + `UTC`).
- **Sistemas Autorizados:** Actores o sistemas verificados y autenticados con permiso para consultar **Historias Clínicas**: el **CMO** durante la misión (con control de consentimiento para datos no críticos) y el **Cirujano de Vuelo en Tierra** mediante enlace diferido.
- **Entorno de Misión y Modo Operativo:** Contexto físico (**Nave / Estación Espacial / Rover / Base Lunar-Marte**) y modo (**IVA** en cabina vs. **EVA** en traje espacial), que adaptan umbrales clínicos, fuentes activas y guías.
- **Peligros del Vuelo Espacial (5 Hazards):** Los cinco riesgos mapeados por el HRP de la NASA — radiación espacial, aislamiento y confinamiento, distancia de la Tierra, campos gravitatorios y entornos hostiles/cerrados — vinculados a métricas y alertas.

## Anatomía y Respuesta Clínica

- **Figura Anatómica:** Modelo visual interactivo y responsivo del cuerpo humano organizado por **Regiones Anatómicas y Sistemas Corporales** (extremidades y sistemas cardiovascular, respiratorio, musculoesquelético, inmune y neuro-ocular) que refleja el estado del astronauta.
- **Sintomatología:** Conjunto de síntomas o señales clínicas registradas o asociadas a una **Región Anatómica** o **Sistema Corporal**.
- **Procedimientos y Niveles de Escalamiento:** Instrucciones paso a paso en lenguaje sencillo para tomar mediciones, normalizar anomalías o atender emergencias, escalando según gravedad entre **Autocuidado**, **CMO** y **Cirujano de Vuelo en Tierra**.
