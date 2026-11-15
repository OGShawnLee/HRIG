# Software Specification

> Refined software project specification for **HRIG**. Maintained incrementally in Spanish until final delivery, when it will be translated to English.
> Key domain terms in bold follow the definitions in [`DOMAIN.md`](DOMAIN.md).

## Requerimientos Funcionales

- **RF01:** El sistema debe mostrar un dashboard principal con los **Datos Primarios** del **Perfil de Tripulante** activo, indicando valor actual, unidad, rango objetivo y estado (**Valor Normal**, **Aviso**, **Valor Anómalo**, **Valor Crítico** o **Dato Obsoleto**).
- **RF02:** El sistema debe registrar de forma cronológica cada medición, imagen diagnóstica, **Sintomatología** y **Alerta** en la **Historia Clínica** del **Perfil de Tripulante** correspondiente.
- **RF03:** El sistema debe ingerir datos de salud mediante **Fuentes de Captura Multimodal**, incluyendo flujos automáticos de **Equipos de Monitoreo de Salud** (IVA o bus cableado EVA), **Entrada Manual** (valores numéricos y escalas de sueño/ánimo) y **Captura de Imágenes** diagnósticas.
- **RF04:** El sistema debe mostrar instrucciones paso a paso en lenguaje sencillo sobre cómo tomar cada medición y utilizar las **Fuentes de Captura Multimodal** correspondientes.
- **RF05:** El sistema debe explicar la importancia clínica de cada indicador y mostrar su relación con los **Peligros del Vuelo Espacial (5 Hazards)** aplicables.
- **RF06:** Cuando un indicador alcance un **Valor Crítico**, el sistema debe mostrar inmediatamente el protocolo de acción de emergencia, los factores causales y el **Nivel de Escalamiento** requerido (**Autocuidado**, **CMO** o **Cirujano de Vuelo en Tierra**).
- **RF07:** Cuando un indicador presente un **Aviso** o **Valor Anómalo**, el sistema debe mostrar recomendaciones accionables para normalizarlo y prevenir su escalamiento a estado crítico.
- **RF08:** Cuando se detecte un **Valor Crítico** en el entorno o en cualquier miembro de la **Tripulación**, el sistema debe emitir una **Alerta** a toda la **Tripulación** indicando el tripulante o zona afectada, la métrica disparadora y la acción inmediata.
- **RF09:** El sistema debe proporcionar acceso a la **Historia Clínica** únicamente a **Sistemas Autorizados** verificados y autenticados, respetando el consentimiento explícito del tripulante para registros personales no críticos.
- **RF10:** El sistema debe mostrar una vista de seguimiento con los **Datos Secundarios** del astronauta y gráficas de tendencia temporal para evaluar su evolución a mediano y largo plazo.
- **RF11:** El sistema debe mostrar una **Figura Anatómica** responsiva que refleje el estado general y por **Región Anatómica** o **Sistema Corporal** mediante codificación redundante (color, ícono y etiqueta).
- **RF12:** El sistema debe permitir seleccionar una **Región Anatómica** (incluyendo extremidades) o **Sistema Corporal** en la **Figura Anatómica** para consultar sus **Procedimientos** paso a paso y los **Peligros del Vuelo Espacial** asociados.
- **RF13:** El sistema debe mostrar la **Sintomatología** asociada o registrada al seleccionar una **Región Anatómica** o **Sistema Corporal** en la **Figura Anatómica**.
- **RF14:** El sistema debe permitir crear, editar y alternar el **Perfil de Tripulante** individual de cada miembro de la **Tripulación**, almacenando su rol, rangos base, preferencias de privacidad y acceso a su **Historia Clínica**.
- **RF15:** El sistema debe permitir cambiar el **Entorno de Misión y Modo Operativo** (**Nave / Estación Espacial / Rover / Base Lunar-Marte** e **IVA / EVA**) para ajustar automáticamente umbrales clínicos, fuentes de sensores y guías operativas.
- **RF16:** El sistema debe generar recordatorios programados de rutina de misión para que el astronauta realice mediciones periódicas o chequeos preventivos de **Datos Secundarios**.

## Requisitos No Funcionales

### Usabilidad y Factores Humanos (UI/UX)

- **RNF01 — Diseño para usuarios sin formación médica:** La interfaz debe presentar instrucciones en lenguaje sencillo y **Procedimientos** paso a paso comprensibles bajo estrés o fatiga.
- **RNF02 — Contraste y legibilidad en cabina:** El sistema debe operar con una paleta oscura (*dark mode first*) manteniendo una relación de contraste ≥ 6:1 para texto general y ≥ 10:1 para valores críticos y alertas.
- **RNF03 — Redundancia visual y codificación de estados:** Ningún estado de salud debe depender exclusivamente del color; cada estado debe incluir un ícono/forma distintiva y una etiqueta de texto explícita.
- **RNF04 — Tipografía inequívoca y alineación numérica:** Los signos vitales, marcas de tiempo y tablas deben utilizar tipografías *sans-serif* con clara diferenciación de caracteres (`l`/`1`, `O`/`0`) y numerales tabulares de ancho fijo.
- **RNF05 — Cero desplazamiento horizontal y navegación mínima:** Los dashboards y tablas no deben requerir desplazamiento horizontal; las tareas activas (instrucción de toma, ingreso y evaluación) deben resolverse en una misma vista para evitar saltos de pantalla.
- **RNF06 — Fricción intencional en anulación de alarmas:** Silenciar o anular una **Alerta** crítica debe exigir una confirmación explícita que advierta las consecuencias operativas, evitando apagados accidentales.
- **RNF07 — Adaptabilidad multidispositivo y modos operativos:** La interfaz debe adaptarse a modo tableta (*touch-first* para uso con guantes), estación de trabajo (vista multipanel para tendencias e imágenes) y HUD compacto EVA (solo **Datos Primarios** y **Alertas** activas, sin desplazamiento).

### Rendimiento y Tiempo Real

- **RNF08 — Latencia de visualización y alerta crítica:** La actualización del dashboard principal y el disparo de **Alertas** visuales y sonoras ante **Valores Críticos** debe ejecutarse en tiempo real con una latencia inferior a 2 segundos tras la recepción del dato.

### Disponibilidad, Tolerancia a Fallos y Operación Offline

- **RNF09 — Operación 100% autónoma (Offline-First):** Todas las funciones de captura, evaluación de umbrales, visualización de tendencias, guías clínicas y almacenamiento de **Historias Clínicas** deben operar sin conexión a Tierra ni dependencias en la nube.
- **RNF10 — Detección e indicación de datos obsoletos (Stale Data):** Si una fuente de monitoreo se desconecta o deja de transmitir, el sistema nunca debe presentar la última lectura como en vivo; debe marcar el indicador inmediatamente como **Dato Obsoleto (`STALE / NO DATA`)** con su última marca de tiempo registrada.

### Seguridad, Privacidad y Trazabilidad

- **RNF11 — Integridad y trazabilidad de registros:** Cada medición almacenada en la **Historia Clínica** debe etiquetarse de forma inmutable con la fuente de origen, el **Perfil de Tripulante** y la marca de tiempo dual (**MET** - *Mission Elapsed Time* + **UTC**).
- **RNF12 — Privacidad y acceso basado en consentimiento:** Los registros personales y conductuales no críticos deben permanecer privados por defecto para el astronauta, habilitando su compartición con el **CMO** o sistemas en tierra únicamente mediante consentimiento explícito o protocolo de emergencia autorizado.
