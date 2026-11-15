import { computed, ref } from "vue"
import { Events } from "@wailsio/runtime"
import "@bindings/events/index.js"
import {
  GetAstronauts,
  GetConnectionStatus,
  GetMetrics,
  GetSnapshot,
} from "@bindings/hrig/app/app.js"
import { ConnectionState } from "@bindings/hrig/app/models.js"
import {
  ReadingStatus,
  type AlertEvent,
  type Astronaut,
  type LatestReading,
  type Metric,
  type RealtimeEnvelope,
} from "@bindings/hrig/shared/model/models.js"

export type TelemetryDisplayStatus = ReadingStatus | "unknown"

export interface FormattedMetric {
  display: string
  status: TelemetryDisplayStatus
}

const STATUS_SEVERITY: Record<TelemetryDisplayStatus, number> = {
  unknown: 0,
  [ReadingStatus.$zero]: 0,
  [ReadingStatus.StatusNormal]: 1,
  [ReadingStatus.StatusAbnormal]: 2,
  [ReadingStatus.StatusStale]: 3,
  [ReadingStatus.StatusCritical]: 4,
}

const astronauts = ref<Astronaut[]>([])
const metricsByCode = ref<Record<string, Metric>>({})
const readingsByKey = ref<Record<string, LatestReading>>({})
const alerts = ref<AlertEvent[]>([])
const connectionState = ref<ConnectionState>(new ConnectionState())
const loading = ref<boolean>(false)
const error = ref<string | null>(null)
const lastUpdatedUtc = ref<string | null>(null)

let initialized = false
let listenersRegistered = false
let activeRefreshPromise: Promise<void> | null = null

function readingKey(astronautId: string | null | undefined, metricCode: string): string {
  const scope = astronautId && astronautId.trim() !== "" ? astronautId : "env"
  return `${scope}:${metricCode}`
}

function upsertReading(reading: LatestReading | null | undefined): void {
  if (!reading || !reading.metricCode) {
    return
  }
  const key = readingKey(reading.astronautId, reading.metricCode)
  readingsByKey.value = {
    ...readingsByKey.value,
    [key]: reading,
  }
}

function applySnapshot(
  readings: LatestReading[] | undefined,
  nextAlerts: AlertEvent[] | undefined,
  timestamp?: string,
): void {
  const nextMap: Record<string, LatestReading> = {}
  if (Array.isArray(readings)) {
    for (const reading of readings) {
      if (!reading || !reading.metricCode) {
        continue
      }
      nextMap[readingKey(reading.astronautId, reading.metricCode)] = reading
    }
  }
  readingsByKey.value = nextMap
  if (Array.isArray(nextAlerts)) {
    alerts.value = nextAlerts
  }
  lastUpdatedUtc.value = timestamp && timestamp !== "" ? timestamp : new Date().toISOString()
}

function upsertAlert(alert: AlertEvent | null | undefined): void {
  if (!alert || !alert.id) {
    return
  }
  const index = alerts.value.findIndex((item) => item.id === alert.id)
  if (index >= 0) {
    const updated = [...alerts.value]
    updated[index] = alert
    alerts.value = updated
  } else {
    alerts.value = [alert, ...alerts.value]
  }
}

function handleEnvelope(envelope: RealtimeEnvelope | null | undefined): void {
  if (!envelope) {
    return
  }
  switch (envelope.type) {
    case "snapshot":
      applySnapshot(envelope.readings, envelope.alerts, envelope.sentAtUtc)
      error.value = null
      break
    case "reading":
    case "stale":
      if (envelope.reading) {
        upsertReading(envelope.reading)
      }
      if (Array.isArray(envelope.readings)) {
        for (const item of envelope.readings) {
          upsertReading(item)
        }
      }
      lastUpdatedUtc.value =
        envelope.sentAtUtc && envelope.sentAtUtc !== ""
          ? envelope.sentAtUtc
          : new Date().toISOString()
      break
    case "alert":
    case "alert_ack":
      upsertAlert(envelope.alert)
      if (envelope.sentAtUtc) {
        lastUpdatedUtc.value = envelope.sentAtUtc
      }
      break
  }
}

function registerListeners(): void {
  if (listenersRegistered) {
    return
  }
  listenersRegistered = true

  Events.On("telemetry:envelope", (ev) => {
    handleEnvelope(ev.data)
  })

  Events.On("telemetry:connection", (ev) => {
    if (!ev.data) {
      return
    }
    const wasConnected = connectionState.value.connected
    connectionState.value = ev.data
    if (!wasConnected && ev.data.connected && (astronauts.value.length === 0 || error.value !== null)) {
      void refresh()
    }
  })
}

async function refresh(): Promise<void> {
  if (activeRefreshPromise) {
    return activeRefreshPromise
  }

  loading.value = true
  error.value = null

  activeRefreshPromise = (async () => {
    try {
      const [connResult, astroResult, metricsResult, snapshotResult] = await Promise.allSettled([
        GetConnectionStatus(),
        GetAstronauts(),
        GetMetrics(),
        GetSnapshot(),
      ])

      if (connResult.status === "fulfilled" && connResult.value) {
        connectionState.value = connResult.value
      }

      const errors: string[] = []

      if (astroResult.status === "fulfilled") {
        astronauts.value = Array.isArray(astroResult.value) ? astroResult.value : []
      } else {
        errors.push(
          astroResult.reason instanceof Error
            ? astroResult.reason.message
            : String(astroResult.reason),
        )
      }

      if (metricsResult.status === "fulfilled") {
        const nextMetrics: Record<string, Metric> = {}
        for (const metric of metricsResult.value ?? []) {
          if (metric && metric.code) {
            nextMetrics[metric.code] = metric
          }
        }
        metricsByCode.value = nextMetrics
      } else {
        errors.push(
          metricsResult.reason instanceof Error
            ? metricsResult.reason.message
            : String(metricsResult.reason),
        )
      }

      if (snapshotResult.status === "fulfilled" && snapshotResult.value) {
        applySnapshot(
          snapshotResult.value.readings,
          snapshotResult.value.alerts,
          snapshotResult.value.generatedAtUtc,
        )
      } else if (snapshotResult.status === "rejected") {
        errors.push(
          snapshotResult.reason instanceof Error
            ? snapshotResult.reason.message
            : String(snapshotResult.reason),
        )
      }

      error.value = errors.length > 0 ? errors[0] : null
    } finally {
      loading.value = false
      activeRefreshPromise = null
    }
  })()

  return activeRefreshPromise
}

function ensureInitialized(): void {
  if (initialized) {
    return
  }
  initialized = true
  registerListeners()
  void refresh()
}

export function resolveReadingStatus(reading: LatestReading | undefined): TelemetryDisplayStatus {
  if (!reading) {
    return "unknown"
  }
  if (reading.isStale || reading.status === ReadingStatus.StatusStale) {
    return ReadingStatus.StatusStale
  }
  if (
    reading.status === ReadingStatus.StatusNormal ||
    reading.status === ReadingStatus.StatusAbnormal ||
    reading.status === ReadingStatus.StatusCritical
  ) {
    return reading.status
  }
  return "unknown"
}

export function getCrewReading(astronautId: string, metricCode: string): LatestReading | undefined {
  return readingsByKey.value[readingKey(astronautId, metricCode)]
}

export function getEnvReading(metricCode: string): LatestReading | undefined {
  return readingsByKey.value[readingKey(null, metricCode)]
}

export function formatMetricValue(
  reading: LatestReading | undefined,
  metric: Metric | undefined,
): FormattedMetric {
  if (!reading) {
    return { display: "—", status: "unknown" }
  }

  const status = resolveReadingStatus(reading)
  let rawValue: string | null = null

  if (reading.valueNumeric !== null && reading.valueNumeric !== undefined && Number.isFinite(reading.valueNumeric)) {
    const decimals = metric?.baseUnitDecimals ?? 0
    rawValue = reading.valueNumeric.toFixed(decimals)
  } else if (reading.valueScale !== null && reading.valueScale !== undefined) {
    rawValue = String(reading.valueScale)
  } else if (reading.valueText !== null && reading.valueText !== undefined && reading.valueText.trim() !== "") {
    rawValue = reading.valueText.trim()
  }

  if (rawValue === null) {
    return { display: "—", status }
  }

  const unit = metric?.displayUnitSymbol ?? metric?.baseUnitSymbol ?? reading.unitCode ?? ""
  if (!unit) {
    return { display: rawValue, status }
  }
  if (unit === "%") {
    return { display: `${rawValue}%`, status }
  }
  return { display: `${rawValue} ${unit}`, status }
}

function pickWorstStatus(a: TelemetryDisplayStatus, b: TelemetryDisplayStatus): TelemetryDisplayStatus {
  return STATUS_SEVERITY[a] >= STATUS_SEVERITY[b] ? a : b
}

export function formatBloodPressure(astronautId: string): FormattedMetric {
  const sysReading = getCrewReading(astronautId, "bp_systolic")
  const diaReading = getCrewReading(astronautId, "bp_diastolic")

  if (!sysReading && !diaReading) {
    return { display: "—", status: "unknown" }
  }

  const sysMetric = metricsByCode.value["bp_systolic"]
  const diaMetric = metricsByCode.value["bp_diastolic"]

  const sysDecimals = sysMetric?.baseUnitDecimals ?? 0
  const diaDecimals = diaMetric?.baseUnitDecimals ?? 0

  const sysText =
    sysReading?.valueNumeric !== null &&
    sysReading?.valueNumeric !== undefined &&
    Number.isFinite(sysReading.valueNumeric)
      ? sysReading.valueNumeric.toFixed(sysDecimals)
      : "—"

  const diaText =
    diaReading?.valueNumeric !== null &&
    diaReading?.valueNumeric !== undefined &&
    Number.isFinite(diaReading.valueNumeric)
      ? diaReading.valueNumeric.toFixed(diaDecimals)
      : "—"

  const unit =
    sysMetric?.displayUnitSymbol ??
    sysMetric?.baseUnitSymbol ??
    diaMetric?.displayUnitSymbol ??
    diaMetric?.baseUnitSymbol ??
    sysReading?.unitCode ??
    diaReading?.unitCode ??
    "mmHg"

  const sysStatus = resolveReadingStatus(sysReading)
  const diaStatus = resolveReadingStatus(diaReading)

  return {
    display: `${sysText}/${diaText} ${unit}`,
    status: pickWorstStatus(sysStatus, diaStatus),
  }
}

export function useTelemetry() {
  ensureInitialized()

  const activeAstronauts = computed(() =>
    astronauts.value.filter((astronaut) => astronaut.isActive),
  )

  return {
    astronauts,
    activeAstronauts,
    metricsByCode,
    readingsByKey,
    alerts,
    connectionState,
    loading,
    error,
    lastUpdatedUtc,
    refresh,
    getCrewReading,
    getEnvReading,
    resolveReadingStatus,
    formatMetricValue,
    formatBloodPressure,
  }
}
