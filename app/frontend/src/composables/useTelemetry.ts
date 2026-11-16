import { computed, ref } from "vue"
import { Events } from "@wailsio/runtime"
import "@bindings/events/index.js"
import {
  GetAstronauts,
  GetConnectionStatus,
  GetMetrics,
  GetSnapshot,
  GetTrends,
} from "@bindings/hrig/app/app.js"
import { ConnectionState } from "@bindings/hrig/app/models.js"
import {
  ReadingStatus,
  TrendQuery,
  type AlertEvent,
  type Astronaut,
  type LatestReading,
  type Metric,
  type RealtimeEnvelope,
  type TrendPoint,
} from "@bindings/hrig/shared/model/models.js"
import {
  buildAlertEpisodeKey,
  clearAlertIfResolved,
  dispatchTelemetryStatus,
  lastTriggeredAlert,
  type AlertDispatchPayload,
  type NavReadingStatus,
} from "@/composables/useTelemetryNavigation"

export type TelemetryDisplayStatus = ReadingStatus | "unknown"

export interface FormattedMetric {
  display: string
  status: TelemetryDisplayStatus
}

export interface TelemetrySeriesPoint {
  value: number
  status: TelemetryDisplayStatus
  recordedAtUtc: string
  met: string
}

const MAX_SERIES_POINTS = 40

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
const seriesByKey = ref<Record<string, TelemetrySeriesPoint[]>>({})
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

function extractNumericValue(item: {
  valueNumeric?: number | null
  valueScale?: number | null
}): number | null {
  if (item.valueNumeric !== null && item.valueNumeric !== undefined && Number.isFinite(item.valueNumeric)) {
    return item.valueNumeric
  }
  if (item.valueScale !== null && item.valueScale !== undefined && Number.isFinite(item.valueScale)) {
    return item.valueScale
  }
  return null
}

function comparePointsByTime(a: TelemetrySeriesPoint, b: TelemetrySeriesPoint): number {
  const ta = Date.parse(a.recordedAtUtc)
  const tb = Date.parse(b.recordedAtUtc)
  if (!Number.isNaN(ta) && !Number.isNaN(tb)) {
    return ta - tb
  }
  return a.recordedAtUtc.localeCompare(b.recordedAtUtc)
}

function mergeSeriesPoints(
  existing: TelemetrySeriesPoint[] | undefined,
  incoming: TelemetrySeriesPoint[],
  limit = MAX_SERIES_POINTS,
): TelemetrySeriesPoint[] {
  const cap = Math.max(1, Math.min(limit, MAX_SERIES_POINTS))
  const byTimestamp = new Map<string, TelemetrySeriesPoint>()
  for (const pt of existing ?? []) {
    if (pt.recordedAtUtc) {
      byTimestamp.set(pt.recordedAtUtc, pt)
    }
  }
  for (const pt of incoming) {
    if (pt.recordedAtUtc) {
      byTimestamp.set(pt.recordedAtUtc, pt)
    }
  }
  const merged = Array.from(byTimestamp.values()).sort(comparePointsByTime)
  return merged.length > cap ? merged.slice(-cap) : merged
}

function toSeriesPointFromReading(reading: LatestReading): TelemetrySeriesPoint | null {
  if (!reading.metricCode || reading.isStale) {
    return null
  }
  const value = extractNumericValue(reading)
  if (value === null) {
    return null
  }
  return {
    value,
    status: resolveReadingStatus(reading),
    recordedAtUtc: reading.recordedAtUtc,
    met: reading.met ?? "",
  }
}

function toSeriesPointFromTrend(point: TrendPoint): TelemetrySeriesPoint | null {
  const value = extractNumericValue(point)
  if (value === null) {
    return null
  }
  const status: TelemetryDisplayStatus =
    point.status === ReadingStatus.StatusNormal ||
    point.status === ReadingStatus.StatusAbnormal ||
    point.status === ReadingStatus.StatusCritical ||
    point.status === ReadingStatus.StatusStale
      ? point.status
      : "unknown"
  return {
    value,
    status,
    recordedAtUtc: point.recordedAtUtc,
    met: point.met ?? "",
  }
}

function appendReadingToSeriesMap(
  target: Record<string, TelemetrySeriesPoint[]>,
  reading: LatestReading | null | undefined,
): boolean {
  if (!reading) {
    return false
  }
  const pt = toSeriesPointFromReading(reading)
  if (!pt) {
    return false
  }
  const key = readingKey(reading.astronautId, reading.metricCode)
  target[key] = mergeSeriesPoints(target[key], [pt])
  return true
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
  const nextSeries = { ...seriesByKey.value }
  let seriesChanged = false

  if (Array.isArray(readings)) {
    for (const reading of readings) {
      if (!reading || !reading.metricCode) {
        continue
      }
      nextMap[readingKey(reading.astronautId, reading.metricCode)] = reading
      if (appendReadingToSeriesMap(nextSeries, reading)) {
        seriesChanged = true
      }
    }
  }
  readingsByKey.value = nextMap
  if (seriesChanged) {
    seriesByKey.value = nextSeries
  }
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

function toNavStatus(status: TelemetryDisplayStatus): NavReadingStatus | null {
  switch (status) {
    case ReadingStatus.StatusNormal:
      return "normal"
    case ReadingStatus.StatusAbnormal:
      return "abnormal"
    case ReadingStatus.StatusCritical:
      return "critical"
    case ReadingStatus.StatusStale:
      return "stale"
    default:
      return null
  }
}

function resolveAstronautName(astronautId: string | null | undefined): string | undefined {
  if (!astronautId) {
    return undefined
  }
  return astronauts.value.find((item) => item.id === astronautId)?.fullName
}

function readingToAlertPayload(
  reading: LatestReading,
  status: "abnormal" | "critical",
): AlertDispatchPayload {
  const formatted = formatMetricValue(reading, metricsByCode.value[reading.metricCode])
  return {
    status,
    astronautId: reading.astronautId ?? undefined,
    astronautName: resolveAstronautName(reading.astronautId),
    metricCode: reading.metricCode,
    source: reading.source,
    deviceCode: reading.deviceCode,
    valueDisplay: formatted.display,
    met: reading.met,
    timestampUtc: reading.recordedAtUtc,
  }
}

function alertToDispatchPayload(
  alert: AlertEvent,
  status: "abnormal" | "critical",
): AlertDispatchPayload {
  return {
    status,
    astronautId: alert.astronautId ?? undefined,
    astronautName: resolveAstronautName(alert.astronautId),
    metricCode: alert.metricCode,
    source: alert.source,
    deviceCode: alert.deviceCode,
    message: alert.message,
    met: alert.met,
    timestampUtc: alert.triggeredAtUtc,
  }
}

function isAlertSupersededByNormalReading(alert: AlertEvent): boolean {
  const live = readingsByKey.value[readingKey(alert.astronautId, alert.metricCode)]
  if (!live || resolveReadingStatus(live) !== ReadingStatus.StatusNormal) {
    return false
  }
  if (!live.recordedAtUtc || !alert.triggeredAtUtc) {
    return true
  }
  const liveMs = Date.parse(live.recordedAtUtc)
  const alertMs = Date.parse(alert.triggeredAtUtc)
  if (Number.isNaN(liveMs) || Number.isNaN(alertMs)) {
    return true
  }
  return liveMs >= alertMs
}

function reconcileActiveAlerts(): void {
  const activePayloads: AlertDispatchPayload[] = []
  const activeEpisodeKeys = new Set<string>()

  for (const reading of Object.values(readingsByKey.value)) {
    const navStatus = toNavStatus(resolveReadingStatus(reading))
    if (navStatus === "abnormal" || navStatus === "critical") {
      const payload = readingToAlertPayload(reading, navStatus)
      activePayloads.push(payload)
      activeEpisodeKeys.add(buildAlertEpisodeKey(payload))
    }
  }

  for (const alert of alerts.value) {
    if (alert.acknowledged) {
      continue
    }
    const navStatus = toNavStatus(alert.status)
    if (navStatus !== "abnormal" && navStatus !== "critical") {
      continue
    }
    if (isAlertSupersededByNormalReading(alert)) {
      continue
    }
    const payload = alertToDispatchPayload(alert, navStatus)
    activePayloads.push(payload)
    activeEpisodeKeys.add(buildAlertEpisodeKey(payload))
  }

  let bestActive: AlertDispatchPayload | null = null
  for (const candidate of activePayloads) {
    if (!bestActive) {
      bestActive = candidate
      continue
    }
    const candRank = candidate.status === "critical" ? 2 : 1
    const bestRank = bestActive.status === "critical" ? 2 : 1
    if (candRank > bestRank) {
      bestActive = candidate
      continue
    }
    if (candRank === bestRank) {
      const candTime = candidate.timestampUtc ? Date.parse(candidate.timestampUtc) : 0
      const bestTime = bestActive.timestampUtc ? Date.parse(bestActive.timestampUtc) : 0
      if (candTime >= bestTime) {
        bestActive = candidate
      }
    }
  }

  if (bestActive === null && lastTriggeredAlert.value?.source === "mockup-demo") {
    return
  }

  clearAlertIfResolved(bestActive, activeEpisodeKeys)
}

function handleEnvelope(envelope: RealtimeEnvelope | null | undefined): void {
  if (!envelope) {
    return
  }
  switch (envelope.type) {
    case "snapshot":
      applySnapshot(envelope.readings, envelope.alerts, envelope.sentAtUtc)
      reconcileActiveAlerts()
      error.value = null
      break
    case "reading":
    case "stale": {
      const incomingReadings: LatestReading[] = []
      if (envelope.reading) {
        incomingReadings.push(envelope.reading)
      }
      if (Array.isArray(envelope.readings)) {
        for (const item of envelope.readings) {
          if (item) {
            incomingReadings.push(item)
          }
        }
      }

      const nextSeries = envelope.type === "reading" ? { ...seriesByKey.value } : null
      let seriesChanged = false
      for (const item of incomingReadings) {
        upsertReading(item)
        if (nextSeries && appendReadingToSeriesMap(nextSeries, item)) {
          seriesChanged = true
        }
      }
      if (nextSeries && seriesChanged) {
        seriesByKey.value = nextSeries
      }

      if (envelope.type === "reading") {
        for (const item of incomingReadings) {
          const navStatus = toNavStatus(resolveReadingStatus(item))
          if (navStatus === "abnormal" || navStatus === "critical") {
            dispatchTelemetryStatus(readingToAlertPayload(item, navStatus))
          }
        }
      }

      reconcileActiveAlerts()

      lastUpdatedUtc.value =
        envelope.sentAtUtc && envelope.sentAtUtc !== ""
          ? envelope.sentAtUtc
          : new Date().toISOString()
      break
    }
    case "alert":
    case "alert_ack":
      upsertAlert(envelope.alert)
      if (envelope.type === "alert" && envelope.alert && !envelope.alert.acknowledged) {
        const navStatus = toNavStatus(envelope.alert.status)
        if (navStatus === "abnormal" || navStatus === "critical") {
          dispatchTelemetryStatus(alertToDispatchPayload(envelope.alert, navStatus))
        }
      }
      reconcileActiveAlerts()
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
        reconcileActiveAlerts()
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

export function getCrewSeries(astronautId: string, metricCode: string): TelemetrySeriesPoint[] {
  return seriesByKey.value[readingKey(astronautId, metricCode)] ?? []
}

export function getEnvSeries(metricCode: string): TelemetrySeriesPoint[] {
  return seriesByKey.value[readingKey(null, metricCode)] ?? []
}

export async function loadMetricTrends(
  astronautId: string,
  metricCode: string,
  limit = MAX_SERIES_POINTS,
): Promise<TelemetrySeriesPoint[]> {
  const cleanMetric = metricCode.trim()
  if (!cleanMetric) {
    return []
  }
  const cleanAstro = astronautId.trim()
  const key = readingKey(cleanAstro || null, cleanMetric)
  try {
    const points = await GetTrends(
      new TrendQuery({
        astronautId: cleanAstro || undefined,
        metricCode: cleanMetric,
        limit,
      }),
    )
    const parsed: TelemetrySeriesPoint[] = []
    for (const pt of points ?? []) {
      const converted = toSeriesPointFromTrend(pt)
      if (converted) {
        parsed.push(converted)
      }
    }
    const merged = mergeSeriesPoints(seriesByKey.value[key], parsed, limit)
    seriesByKey.value = {
      ...seriesByKey.value,
      [key]: merged,
    }
    return merged
  } catch {
    return seriesByKey.value[key] ?? []
  }
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
    seriesByKey,
    alerts,
    connectionState,
    loading,
    error,
    lastUpdatedUtc,
    refresh,
    loadMetricTrends,
    getCrewReading,
    getEnvReading,
    getCrewSeries,
    getEnvSeries,
    resolveReadingStatus,
    formatMetricValue,
    formatBloodPressure,
  }
}
