<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import { AlertCircle, AlertTriangle, HelpCircle, Home, Search } from "lucide-vue-next"
import { ReadingStatus, type AlertEvent } from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type TelemetryDisplayStatus,
  type TelemetrySeriesPoint,
} from "@/composables/useTelemetry"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const TRACKED_VITALS_METRICS = [
  "heart_rate",
  "spo2",
  "bp_systolic",
  "bp_diastolic",
  "radiation_rate",
  "core_temp",
] as const

const SVG_WIDTH = 300
const SVG_HEIGHT = 64
const PAD_Y = 8

const route = useRoute()
const router = useRouter()
const {
  astronauts,
  activeAstronauts,
  metricsByCode,
  alerts,
  loadMetricTrends,
  getCrewReading,
  getEnvReading,
  getCrewSeries,
  resolveReadingStatus,
  formatMetricValue,
  formatBloodPressure,
} = useTelemetry()
const {
  selectedAstronautId,
  selectedAstronautFallbackName,
  lastTriggeredAlert,
  setSelectedAstronaut,
  goToHome,
} = useTelemetryNavigation()

const metricFilter = ref("")

function openMedicalManual(manualQuery: string) {
  void router.push({
    name: "search",
    query: { tab: "procedures", q: manualQuery },
  })
}

watch(
  () => route.query.astronaut,
  (astronaut) => {
    if (typeof astronaut === "string" && astronaut.trim() !== "") {
      setSelectedAstronaut(astronaut.trim())
    }
  },
  { immediate: true },
)

const activeAstronaut = computed(() => {
  const targetId = selectedAstronautId.value.trim()
  const pool = activeAstronauts.value.length > 0 ? activeAstronauts.value : astronauts.value
  if (targetId) {
    const exact = pool.find((a) => a.id === targetId)
    if (exact) {
      return exact
    }
    const lower = targetId.toLowerCase()
    const byName = pool.find((a) => a.fullName.toLowerCase().includes(lower))
    if (byName) {
      return byName
    }
  }
  return pool[0] ?? null
})

const resolvedAstronautId = computed(() => activeAstronaut.value?.id ?? selectedAstronautId.value)
const selectedAstronautName = computed(
  () => activeAstronaut.value?.fullName ?? selectedAstronautFallbackName.value,
)

watch(
  resolvedAstronautId,
  (astronautId) => {
    const cleanId = astronautId.trim()
    if (!cleanId) {
      return
    }
    for (const metricCode of TRACKED_VITALS_METRICS) {
      void loadMetricTrends(cleanId, metricCode, 40)
    }
  },
  { immediate: true },
)

type CardStatus = TelemetryDisplayStatus | "fallback"

const CARD_STATUS_SEVERITY: Record<CardStatus, number> = {
  unknown: 0,
  [ReadingStatus.$zero]: 0,
  fallback: 1,
  [ReadingStatus.StatusNormal]: 2,
  [ReadingStatus.StatusStale]: 3,
  [ReadingStatus.StatusAbnormal]: 4,
  [ReadingStatus.StatusCritical]: 5,
}

function pickWorstCardStatus(a: CardStatus, b: CardStatus): CardStatus {
  return CARD_STATUS_SEVERITY[a] >= CARD_STATUS_SEVERITY[b] ? a : b
}

function normalizeAlertStatus(status: string | null | undefined): CardStatus {
  switch (status) {
    case ReadingStatus.StatusCritical:
      return ReadingStatus.StatusCritical
    case ReadingStatus.StatusAbnormal:
      return ReadingStatus.StatusAbnormal
    case ReadingStatus.StatusStale:
      return ReadingStatus.StatusStale
    case ReadingStatus.StatusNormal:
      return ReadingStatus.StatusNormal
    default:
      return "unknown"
  }
}

function matchesAstronautScope(
  alertAstronautId: string | null | undefined,
  selectedId: string,
): boolean {
  const cleanAlertAstro = alertAstronautId?.trim() ?? ""
  if (!cleanAlertAstro) {
    return true
  }
  return cleanAlertAstro === selectedId.trim()
}

function isAlertSupersededByNormalReading(alert: AlertEvent): boolean {
  const live =
    alert.astronautId && alert.astronautId.trim() !== ""
      ? getCrewReading(alert.astronautId.trim(), alert.metricCode)
      : getEnvReading(alert.metricCode)
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

/**
 * Computes the worst severity for a vital card across its primary reading,
 * paired crew/environmental subline metrics, active unacknowledged alerts,
 * and the active triggered alert banner/route state.
 */
function resolveCardStatus(
  astronautId: string,
  primaryStatus: CardStatus,
  crewMetrics: readonly string[],
  envMetrics: readonly string[] = [],
): CardStatus {
  let worst: CardStatus = primaryStatus

  if (astronautId) {
    for (const code of crewMetrics) {
      worst = pickWorstCardStatus(worst, resolveReadingStatus(getCrewReading(astronautId, code)))
    }
  }

  for (const code of envMetrics) {
    worst = pickWorstCardStatus(worst, resolveReadingStatus(getEnvReading(code)))
  }

  const metricSet = new Set([...crewMetrics, ...envMetrics])

  for (const alert of alerts.value) {
    if (alert.acknowledged || !metricSet.has(alert.metricCode)) {
      continue
    }
    if (!matchesAstronautScope(alert.astronautId, astronautId)) {
      continue
    }
    if (isAlertSupersededByNormalReading(alert)) {
      continue
    }
    worst = pickWorstCardStatus(worst, normalizeAlertStatus(alert.status))
  }

  const triggered = lastTriggeredAlert.value
  if (
    triggered &&
    triggered.metricCode &&
    metricSet.has(triggered.metricCode) &&
    matchesAstronautScope(triggered.astronautId, astronautId)
  ) {
    worst = pickWorstCardStatus(worst, normalizeAlertStatus(triggered.status))
  }

  return worst
}

function statusTextClass(status: CardStatus): string {
  switch (status) {
    case ReadingStatus.StatusNormal:
    case "fallback":
      return "text-ok"
    case ReadingStatus.StatusAbnormal:
      return "text-warn"
    case ReadingStatus.StatusCritical:
      return "text-crit vital-text-crit"
    case ReadingStatus.StatusStale:
      return "text-stale"
    default:
      return "text-side"
  }
}

function cardSecondaryTextClass(status: CardStatus, nominalClass: string): string {
  switch (status) {
    case ReadingStatus.StatusAbnormal:
      return "text-warn"
    case ReadingStatus.StatusCritical:
      return "text-crit vital-text-crit"
    default:
      return nominalClass
  }
}

function cardFrameClass(status: CardStatus): string {
  switch (status) {
    case ReadingStatus.StatusAbnormal:
      return "border-2 border-warn vital-frame-warn"
    case ReadingStatus.StatusCritical:
      return "border-2 border-crit vital-frame-crit"
    default:
      return ""
  }
}

function vitalChartStyles(
  nominalStroke: string,
  nominalFill: string,
  status: CardStatus,
): {
  stroke: string
  fill: string
} {
  switch (status) {
    case ReadingStatus.StatusAbnormal:
      return { stroke: "stroke-warn", fill: "fill-warn/20" }
    case ReadingStatus.StatusCritical:
      return { stroke: "stroke-crit", fill: "fill-crit/20" }
    case ReadingStatus.StatusStale:
      return { stroke: "stroke-stale", fill: "fill-stale/15" }
    case ReadingStatus.StatusNormal:
    case "fallback":
    default:
      return { stroke: nominalStroke, fill: nominalFill }
  }
}

interface ChartPointCoord {
  x: number
  y: number
}

interface ChartWaveData {
  wavePath: string
  fillPath?: string
  minVal: number | null
  maxVal: number | null
}

function roundCoord(value: number): number {
  return Math.round(value * 10) / 10
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function computeSeriesPhase(series: TelemetrySeriesPoint[]): number {
  if (series.length === 0) {
    return 0
  }
  const latest = series[series.length - 1]
  const parsedMs = Date.parse(latest.recordedAtUtc)
  if (Number.isFinite(parsedMs)) {
    return (parsedMs / 1000) * 18
  }
  let hash = series.length * 19
  for (let i = 0; i < latest.recordedAtUtc.length; i += 1) {
    hash = (hash * 31 + latest.recordedAtUtc.charCodeAt(i)) % 100000
  }
  return hash
}

function sampleSeriesAt(values: number[], tNorm: number): number {
  if (values.length === 0) {
    return 0
  }
  if (values.length === 1) {
    return values[0]
  }
  const pos = clamp(tNorm, 0, 1) * (values.length - 1)
  const idx = Math.floor(pos)
  const frac = pos - idx
  const next = Math.min(idx + 1, values.length - 1)
  return values[idx] * (1 - frac) + values[next] * frac
}

function computeSeriesRange(
  primarySeries: TelemetrySeriesPoint[],
  secondarySeries: TelemetrySeriesPoint[] = [],
): { minVal: number | null; maxVal: number | null } {
  const allValues = [
    ...primarySeries.map((pt) => pt.value),
    ...secondarySeries.map((pt) => pt.value),
  ]
  if (allValues.length === 0) {
    return { minVal: null, maxVal: null }
  }
  return {
    minVal: Math.min(...allValues),
    maxVal: Math.max(...allValues),
  }
}

function clipPolylineToXRange(
  pts: ChartPointCoord[],
  minX: number,
  maxX: number,
): ChartPointCoord[] {
  const out: ChartPointCoord[] = []
  const pushPoint = (pt: ChartPointCoord) => {
    const prev = out[out.length - 1]
    if (!prev || prev.x !== pt.x || prev.y !== pt.y) {
      out.push(pt)
    }
  }

  for (let i = 0; i < pts.length - 1; i += 1) {
    const a = pts[i]
    const b = pts[i + 1]
    if (a.x < minX && b.x < minX) {
      continue
    }
    if (a.x > maxX && b.x > maxX) {
      continue
    }
    if (a.x === b.x) {
      if (a.x >= minX && a.x <= maxX) {
        pushPoint({ x: roundCoord(a.x), y: roundCoord(a.y) })
        pushPoint({ x: roundCoord(b.x), y: roundCoord(b.y) })
      }
      continue
    }
    const dx = b.x - a.x
    let t0 = 0
    let t1 = 1
    if (a.x < minX) {
      t0 = Math.max(t0, (minX - a.x) / dx)
    }
    if (b.x > maxX) {
      t1 = Math.min(t1, (maxX - a.x) / dx)
    }
    if (t0 <= t1) {
      pushPoint({
        x: roundCoord(a.x + t0 * dx),
        y: roundCoord(a.y + t0 * (b.y - a.y)),
      })
      pushPoint({
        x: roundCoord(a.x + t1 * dx),
        y: roundCoord(a.y + t1 * (b.y - a.y)),
      })
    }
  }

  return out
}

function polylineToPath(pts: ChartPointCoord[]): string {
  if (pts.length === 0) {
    return `M0 ${SVG_HEIGHT / 2}H${SVG_WIDTH}`
  }
  return `M${pts[0].x} ${pts[0].y} ` + pts.slice(1).map((p) => `L${p.x} ${p.y}`).join(" ")
}

function smoothPathFromPoints(pts: ChartPointCoord[]): string {
  if (pts.length === 0) {
    return `M0 ${SVG_HEIGHT / 2}H${SVG_WIDTH}`
  }
  if (pts.length === 1) {
    return `M0 ${pts[0].y}H${SVG_WIDTH}`
  }
  let d = `M${roundCoord(pts[0].x)} ${roundCoord(pts[0].y)}`
  for (let i = 0; i < pts.length - 1; i += 1) {
    const p0 = pts[Math.max(0, i - 1)]
    const p1 = pts[i]
    const p2 = pts[i + 1]
    const p3 = pts[Math.min(pts.length - 1, i + 2)]
    const cp1x = roundCoord(p1.x + (p2.x - p0.x) / 6)
    const cp1y = roundCoord(clamp(p1.y + (p2.y - p0.y) / 6, 4, SVG_HEIGHT - 4))
    const cp2x = roundCoord(p2.x - (p3.x - p1.x) / 6)
    const cp2y = roundCoord(clamp(p2.y - (p3.y - p1.y) / 6, 4, SVG_HEIGHT - 4))
    d += ` C${cp1x} ${cp1y} ${cp2x} ${cp2y} ${roundCoord(p2.x)} ${roundCoord(p2.y)}`
  }
  return d
}

function buildCardiacChart(
  series: TelemetrySeriesPoint[],
  fallbackBpm: number,
): ChartWaveData {
  const { minVal, maxVal } = computeSeriesRange(series)
  const values = series.length > 0 ? series.map((pt) => pt.value) : [fallbackBpm]
  const latestBpm = values[values.length - 1]
  const mean = values.reduce((acc, v) => acc + v, 0) / values.length
  const span = Math.max(...values) - Math.min(...values)

  const beatPeriod = clamp((60 / clamp(latestBpm, 35, 200)) * 88, 34, 150)
  const phasePx = computeSeriesPhase(series) % beatPeriod

  const landmarks: ChartPointCoord[] = []
  for (let beatStart = -phasePx - beatPeriod; beatStart < SVG_WIDTH; beatStart += beatPeriod) {
    const beatNorm = clamp((beatStart + beatPeriod * 0.5) / SVG_WIDTH, 0, 1)
    const sampleBpm = sampleSeriesAt(values, beatNorm)
    const devNorm = span > 0 ? (sampleBpm - mean) / span : 0
    const rAmp = clamp(24 + clamp((sampleBpm - 45) / 115, 0, 1) * 10 + devNorm * 5, 18, 36)
    const yBase = clamp(40 - devNorm * 3, 36, 44)

    const beatPoints: [number, number][] = [
      [0.0, yBase],
      [0.1, yBase],
      [0.15, yBase - 3.5],
      [0.2, yBase],
      [0.27, yBase],
      [0.3, yBase + 6],
      [0.34, clamp(yBase - rAmp, 5, 24)],
      [0.38, clamp(yBase + 8, 42, 54)],
      [0.42, yBase],
      [0.5, yBase - 1],
      [0.57, yBase - 6.5],
      [0.65, yBase],
      [0.82, clamp(yBase + devNorm * 1.5, 36, 44)],
    ]
    for (const [f, y] of beatPoints) {
      landmarks.push({ x: beatStart + f * beatPeriod, y })
    }
  }

  const clipped = clipPolylineToXRange(landmarks, 0, SVG_WIDTH)
  return {
    wavePath: polylineToPath(clipped),
    minVal,
    maxVal,
  }
}

function periodicDist(u: number, center: number): number {
  return ((u - center + 1.5) % 1) - 0.5
}

function buildOxygenationChart(
  series: TelemetrySeriesPoint[],
  fallbackSpo2: number,
): ChartWaveData {
  const { minVal, maxVal } = computeSeriesRange(series)
  const values = series.length > 0 ? series.map((pt) => pt.value) : [fallbackSpo2]
  const seriesMin = Math.min(...values)
  const seriesSpan = Math.max(...values) - seriesMin

  const cycleWidth = 60
  const phasePx = computeSeriesPhase(series) % cycleWidth
  const pts: ChartPointCoord[] = []

  for (let x = 0; x <= SVG_WIDTH; x += 4) {
    const tNorm = x / SVG_WIDTH
    const localSpo2 = sampleSeriesAt(values, tNorm)
    const spo2Norm = clamp((localSpo2 - 80) / 20, 0.15, 1.0)
    const u = ((x + phasePx) % cycleWidth) / cycleWidth

    const systolic = Math.exp(-Math.pow(periodicDist(u, 0.28) / 0.11, 2))
    const dicrotic = 0.52 * Math.exp(-Math.pow(periodicDist(u, 0.62) / 0.13, 2))
    const shape = clamp((systolic + dicrotic) / 1.02, 0, 1)

    const troughY = 54 - spo2Norm * 4
    const relMod = seriesSpan > 0 ? ((localSpo2 - seriesMin) / seriesSpan - 0.5) * 6 : 0
    const amplitude = clamp(18 + spo2Norm * 24 + relMod, 14, 44)
    const y = clamp(troughY - shape * amplitude, 6, 58)
    pts.push({ x, y })
  }

  return {
    wavePath: smoothPathFromPoints(pts),
    minVal,
    maxVal,
  }
}

function buildPressureChart(
  sysSeries: TelemetrySeriesPoint[],
  diaSeries: TelemetrySeriesPoint[],
  fallbackSys: number,
  fallbackDia: number,
): ChartWaveData {
  const { minVal, maxVal } = computeSeriesRange(sysSeries, diaSeries)
  const sysValues = sysSeries.length > 0 ? sysSeries.map((pt) => pt.value) : [fallbackSys]
  const diaValues = diaSeries.length > 0 ? diaSeries.map((pt) => pt.value) : [fallbackDia]
  const allValues = [...sysValues, ...diaValues]

  const minScale = Math.min(55, ...allValues)
  const maxScale = Math.max(150, ...allValues)
  const scaleSpan = Math.max(20, maxScale - minScale)
  const usableHeight = SVG_HEIGHT - 2 * PAD_Y
  const toY = (val: number): number =>
    clamp(SVG_HEIGHT - PAD_Y - ((val - minScale) / scaleSpan) * usableHeight, 6, 58)

  const cycleWidth = 52
  const phasePx = (computeSeriesPhase(sysSeries) + computeSeriesPhase(diaSeries) * 0.5) % cycleWidth
  const pts: ChartPointCoord[] = []

  for (let x = 0; x <= SVG_WIDTH; x += 3) {
    const tNorm = x / SVG_WIDTH
    const localSys = sampleSeriesAt(sysValues, tNorm)
    const localDia = sampleSeriesAt(diaValues, tNorm)
    const yPeak = toY(Math.max(localSys, localDia + 5))
    const yTrough = toY(Math.min(localDia, localSys - 5))

    const u = ((x + phasePx) % cycleWidth) / cycleWidth
    let shape = 0
    if (u < 0.14) {
      shape = 0.5 - 0.5 * Math.cos((u / 0.14) * Math.PI)
    } else if (u < 0.34) {
      shape = 1.0 - 0.58 * (0.5 - 0.5 * Math.cos(((u - 0.14) / 0.2) * Math.PI))
    } else if (u < 0.46) {
      shape = 0.42 + 0.16 * Math.sin(((u - 0.34) / 0.12) * Math.PI)
    } else {
      shape = 0.42 * Math.pow(1 - (u - 0.46) / 0.54, 1.35)
    }

    const y = clamp(yTrough - shape * (yTrough - yPeak), 6, 58)
    pts.push({ x, y })
  }

  return {
    wavePath: smoothPathFromPoints(pts),
    minVal,
    maxVal,
  }
}

function buildRadiationChart(
  series: TelemetrySeriesPoint[],
  fallbackRad: number,
): ChartWaveData {
  const { minVal, maxVal } = computeSeriesRange(series)
  const values = series.length > 0 ? series.map((pt) => pt.value) : [fallbackRad]
  const vMin = Math.min(...values)
  const vMax = Math.max(...values)
  const span = vMax - vMin
  const ceiling = Math.max(45, vMax)

  const period = 55
  const rampWidth = 29
  const yBase = 60
  const phasePx = computeSeriesPhase(series) % period

  const landmarks: ChartPointCoord[] = []
  for (let tStart = -phasePx - period; tStart < SVG_WIDTH; tStart += period) {
    const tNorm = clamp((tStart + rampWidth * 0.5) / SVG_WIDTH, 0, 1)
    const sampleVal = sampleSeriesAt(values, tNorm)
    const absRatio = clamp(sampleVal / ceiling, 0.15, 1.0)
    const relRatio = span > 0 ? (sampleVal - vMin) / span : 0.65
    const heightFactor = clamp(0.45 * absRatio + 0.55 * (0.35 + 0.65 * relRatio), 0.2, 1.0)
    const toothHeight = heightFactor * 48
    const yPeak = yBase - toothHeight
    const yMid = yBase - toothHeight * 0.48

    landmarks.push(
      { x: tStart, y: yBase },
      { x: tStart, y: yPeak },
      { x: tStart + rampWidth, y: yMid },
      { x: tStart + rampWidth, y: yBase },
      { x: tStart + period, y: yBase },
    )
  }

  const clipped = clipPolylineToXRange(landmarks, 0, SVG_WIDTH)
  const wavePath = polylineToPath(clipped)
  return {
    wavePath,
    fillPath: `${wavePath} V${SVG_HEIGHT} H0 Z`,
    minVal,
    maxVal,
  }
}

function buildTemperatureChart(
  series: TelemetrySeriesPoint[],
  fallbackTemp: number,
): ChartWaveData {
  const { minVal, maxVal } = computeSeriesRange(series)
  const values = series.length > 0 ? series.map((pt) => pt.value) : [fallbackTemp]

  const tempToY = (tempC: number): number => {
    if (tempC <= 36.5) {
      return clamp(56 + (36.5 - tempC) * 1.2, 54, 58)
    }
    if (tempC <= 37.2) {
      return 56 - ((tempC - 36.5) / 0.7) * 2.5
    }
    return clamp(53.5 - ((tempC - 37.2) / 3.3) * 45.5, 8, 53.5)
  }

  const pts: ChartPointCoord[] = []
  for (let x = 0; x <= SVG_WIDTH; x += 12) {
    const tNorm = x / SVG_WIDTH
    const val = sampleSeriesAt(values, tNorm)
    pts.push({ x, y: tempToY(val) })
  }

  return {
    wavePath: smoothPathFromPoints(pts),
    minVal,
    maxVal,
  }
}

function liveReadingNumeric(astronautId: string, metricCode: string, fallback: number): number {
  const val = astronautId ? getCrewReading(astronautId, metricCode)?.valueNumeric : undefined
  return val !== null && val !== undefined && Number.isFinite(val) ? val : fallback
}

function formatRangeLabel(
  minVal: number | null,
  maxVal: number | null,
  metricCode: string,
  unit: string,
): string {
  if (minVal === null || maxVal === null) {
    return "NO DATA"
  }
  const decimals = metricsByCode.value[metricCode]?.baseUnitDecimals ?? 0
  const minText = minVal.toFixed(decimals)
  const maxText = maxVal.toFixed(decimals)
  const suffix = unit === "%" ? "%" : unit ? ` ${unit}` : ""
  return minText === maxText ? `${minText}${suffix}` : `${minText}–${maxText}${suffix}`
}

function formatNumericPart(
  astronautId: string,
  metricCode: string,
  fallbackValue: string,
  fallbackUnit: string,
) {
  const reading = astronautId ? getCrewReading(astronautId, metricCode) : undefined
  const metric = metricsByCode.value[metricCode]
  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === "unknown" || formatted.display === "—") {
    return {
      value: fallbackValue,
      unit: fallbackUnit,
      status: "fallback" as const,
      isLive: false,
    }
  }
  const unit = metric?.displayUnitSymbol ?? metric?.baseUnitSymbol ?? reading?.unitCode ?? fallbackUnit
  const numericOnly =
    reading?.valueNumeric !== null &&
    reading?.valueNumeric !== undefined &&
    Number.isFinite(reading.valueNumeric)
      ? reading.valueNumeric.toFixed(metric?.baseUnitDecimals ?? 0)
      : formatted.display

  return {
    value: unit === "%" ? `${numericOnly}%` : numericOnly,
    unit: unit === "%" ? fallbackUnit : unit,
    status: formatted.status,
    isLive: true,
  }
}

function formatEnvLine(
  metricCode: string,
  prefix: string,
  fallbackDisplay: string,
  spaceBeforeValue = true,
) {
  const reading = getEnvReading(metricCode)
  const metric = metricsByCode.value[metricCode]
  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === "unknown" || formatted.display === "—") {
    return `${prefix}${spaceBeforeValue ? " " : ""}${fallbackDisplay}`
  }
  return `${prefix}${spaceBeforeValue ? " " : ""}${formatted.display}`
}

function formatCrewSubLine(
  astronautId: string,
  metricCode: string,
  prefix: string,
  fallbackDisplay: string,
) {
  const reading = astronautId ? getCrewReading(astronautId, metricCode) : undefined
  const metric = metricsByCode.value[metricCode]
  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === "unknown" || formatted.display === "—") {
    return `${prefix} ${fallbackDisplay}`
  }
  return `${prefix} ${formatted.display}`
}

interface VitalRow {
  id: string
  title: string
  subtitle?: string
  manualQuery: string
  value: string
  unit: string
  subline: string
  status: CardStatus
  rangeLabel: string
  chart: ChartWaveData
  strokeClass: string
  fillClass: string
}

const vitalRows = computed<VitalRow[]>(() => {
  const id = resolvedAstronautId.value

  const hr = formatNumericPart(id, "heart_rate", "72", "bpm")
  const hrCardStatus = resolveCardStatus(id, hr.status, ["heart_rate", "hrv_rmssd"])
  const hrvSub = formatCrewSubLine(id, "hrv_rmssd", "HRV", "25 ms")
  const hrSeries = id ? getCrewSeries(id, "heart_rate") : []
  const hrChart = buildCardiacChart(hrSeries, liveReadingNumeric(id, "heart_rate", 72))
  const hrStyles = vitalChartStyles("stroke-ok", "fill-none", hrCardStatus)

  const spo2 = formatNumericPart(id, "spo2", "98%", "SpO2")
  const spo2CardStatus = resolveCardStatus(id, spo2.status, ["spo2"], ["cabin_o2"])
  const cabinO2Sub = formatEnvLine("cabin_o2", "O2:", "20.9%")
  const spo2Series = id ? getCrewSeries(id, "spo2") : []
  const spo2Chart = buildOxygenationChart(spo2Series, liveReadingNumeric(id, "spo2", 98))
  const spo2Styles = vitalChartStyles("stroke-ok", "fill-none", spo2CardStatus)

  const bpFormatted = id ? formatBloodPressure(id) : { display: "—", status: "unknown" as const }
  const bpLive = bpFormatted.status !== "unknown" && bpFormatted.display !== "—"
  const bpValue = bpLive ? bpFormatted.display.replace(/\s*mmHg$/i, "") : "118/77"
  const bpPrimaryStatus: CardStatus = bpLive ? bpFormatted.status : "fallback"
  const bpCardStatus = resolveCardStatus(
    id,
    bpPrimaryStatus,
    ["bp_systolic", "bp_diastolic"],
    ["cabin_pressure"],
  )
  const cabinPressureSub = formatEnvLine("cabin_pressure", "CABIN:", "101.3 kPa")
  const sysSeries = id ? getCrewSeries(id, "bp_systolic") : []
  const diaSeries = id ? getCrewSeries(id, "bp_diastolic") : []
  const bpChart = buildPressureChart(
    sysSeries,
    diaSeries,
    liveReadingNumeric(id, "bp_systolic", 118),
    liveReadingNumeric(id, "bp_diastolic", 77),
  )
  const bpStyles = vitalChartStyles("stroke-main", "fill-none", bpCardStatus)

  const radRate = formatNumericPart(id, "radiation_rate", "32", "µSv/h")
  const radCardStatus = resolveCardStatus(id, radRate.status, [
    "radiation_rate",
    "radiation_dose_cum",
  ])
  const radCumSub = formatCrewSubLine(id, "radiation_dose_cum", "TOTAL", "1.2 mSv")
  const radSeries = id ? getCrewSeries(id, "radiation_rate") : []
  const radChart = buildRadiationChart(radSeries, liveReadingNumeric(id, "radiation_rate", 32))
  const radStyles = vitalChartStyles("stroke-warn", "fill-warn/20", radCardStatus)

  const coreTemp = formatNumericPart(id, "core_temp", "36.6", "°C")
  const tempCardStatus = resolveCardStatus(id, coreTemp.status, ["core_temp"], ["cabin_temp"])
  const cabinTempSub = formatEnvLine("cabin_temp", "CABIN", "22.0°C")
  const tempSeries = id ? getCrewSeries(id, "core_temp") : []
  const tempChart = buildTemperatureChart(tempSeries, liveReadingNumeric(id, "core_temp", 36.6))
  const tempStyles = vitalChartStyles("stroke-main", "fill-none", tempCardStatus)

  const all: VitalRow[] = [
    {
      id: "cardiac",
      title: "CARDIAC RHYTHM",
      manualQuery: "heart_rate",
      value: hr.value,
      unit: hr.unit,
      subline: hrvSub,
      status: hrCardStatus,
      rangeLabel: formatRangeLabel(hrChart.minVal, hrChart.maxVal, "heart_rate", hr.unit),
      chart: hrChart,
      strokeClass: hrStyles.stroke,
      fillClass: hrStyles.fill,
    },
    {
      id: "oxygenation",
      title: "OXYGENATION",
      manualQuery: "spo2",
      value: spo2.value,
      unit: spo2.unit,
      subline: cabinO2Sub,
      status: spo2CardStatus,
      rangeLabel: formatRangeLabel(spo2Chart.minVal, spo2Chart.maxVal, "spo2", "%"),
      chart: spo2Chart,
      strokeClass: spo2Styles.stroke,
      fillClass: spo2Styles.fill,
    },
    {
      id: "pressure",
      title: "PRESSURE",
      subtitle: "BP",
      manualQuery: "pressure",
      value: bpValue,
      unit: "mmHg",
      subline: cabinPressureSub,
      status: bpCardStatus,
      rangeLabel: formatRangeLabel(bpChart.minVal, bpChart.maxVal, "bp_systolic", "mmHg"),
      chart: bpChart,
      strokeClass: bpStyles.stroke,
      fillClass: bpStyles.fill,
    },
    {
      id: "radiation",
      title: "RADIATION",
      subtitle: "SENSOR",
      manualQuery: "radiation",
      value: radRate.value,
      unit: radRate.unit,
      subline: radCumSub,
      status: radCardStatus,
      rangeLabel: formatRangeLabel(radChart.minVal, radChart.maxVal, "radiation_rate", radRate.unit),
      chart: radChart,
      strokeClass: radStyles.stroke,
      fillClass: radStyles.fill,
    },
    {
      id: "temperature",
      title: "TEMPERATURE",
      subtitle: "BODY",
      manualQuery: "core_temp",
      value: coreTemp.value,
      unit: coreTemp.unit,
      subline: cabinTempSub,
      status: tempCardStatus,
      rangeLabel: formatRangeLabel(tempChart.minVal, tempChart.maxVal, "core_temp", coreTemp.unit),
      chart: tempChart,
      strokeClass: tempStyles.stroke,
      fillClass: tempStyles.fill,
    },
  ]

  const q = metricFilter.value.trim().toLowerCase()
  if (!q) {
    return all
  }
  return all.filter(
    (row) =>
      row.title.toLowerCase().includes(q) ||
      (row.subtitle?.toLowerCase().includes(q) ?? false) ||
      row.subline.toLowerCase().includes(q),
  )
})
</script>

<template>
  <main
    class="flex-1 w-full bg-ground text-main px-4 sm:px-8 py-6 flex flex-col items-center gap-6"
    aria-label="Primary Vitals Dashboard"
  >
    <header class="w-full max-w-4xl flex flex-wrap items-center gap-3">
      <button
        type="button"
        class="button-side h-10 px-3 text-xs gap-1.5 shrink-0"
        aria-label="Return to Home"
        @click="goToHome"
      >
        <Home class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Home</span>
      </button>

      <h1 class="text-lg sm:text-xl font-semibold tracking-wide text-main shrink-0">
        PRIMARY VITALS
      </h1>

      <div class="relative flex-1 min-w-44 flex items-center">
        <Search class="w-4 h-4 text-side absolute left-3 pointer-events-none" aria-hidden="true" />
        <input
          v-model="metricFilter"
          type="text"
          aria-label="Filter vital signs"
          placeholder="Search metrics"
          class="field w-full h-10 pl-9 pr-3 text-sm"
        />
      </div>
    </header>

    <div class="w-full max-w-4xl grid grid-cols-1 md:grid-cols-[1fr_200px] gap-6 items-start">
      <section class="flex flex-col gap-3 min-w-0" aria-label="Primary vital signs">
        <article
          v-for="row in vitalRows"
          :key="row.id"
          class="hud-card overflow-hidden grid grid-cols-[minmax(0,1fr)_140px_36px] sm:grid-cols-[minmax(0,1fr)_168px_44px] min-h-22"
          :class="cardFrameClass(row.status)"
        >
          <div class="bg-surface border-r border-solid border-lineDim flex flex-col justify-between min-w-0">
            <div class="px-3 pt-2 flex items-center justify-between gap-2 text-[11px] font-bold tracking-wider">
              <div class="flex items-center gap-2 min-w-0">
                <h2 :class="statusTextClass(row.status)" class="truncate">{{ row.title }}</h2>
                <span
                  v-if="row.subtitle"
                  class="shrink-0"
                  :class="cardSecondaryTextClass(row.status, 'text-main')"
                >
                  {{ row.subtitle }}
                </span>
              </div>
              <span
                class="text-[10px] font-medium hud-value tabular-nums shrink-0"
                :class="cardSecondaryTextClass(row.status, 'text-dull')"
              >
                {{ row.rangeLabel }}
              </span>
            </div>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-14 px-1" aria-hidden="true">
              <path
                v-if="row.chart.fillPath"
                :d="row.chart.fillPath"
                class="stroke-none"
                :class="row.fillClass"
              />
              <path
                :d="row.chart.wavePath"
                fill="none"
                stroke-width="2"
                stroke-linecap="round"
                :stroke-linejoin="row.id === 'radiation' ? 'miter' : 'round'"
                vector-effect="non-scaling-stroke"
                :class="row.strokeClass"
              />
            </svg>
          </div>

          <div class="px-3 py-2 flex flex-col justify-between hud-value min-w-0">
            <div
              class="flex items-baseline justify-between gap-1 text-[11px] font-semibold"
              :class="cardSecondaryTextClass(row.status, 'text-side')"
            >
              <span class="truncate">{{ row.subtitle === "BP" ? "BP" : "" }}</span>
              <span>{{ row.unit }}</span>
            </div>
            <strong
              class="text-2xl sm:text-3xl font-bold leading-none truncate"
              :class="statusTextClass(row.status)"
            >
              {{ row.value }}
            </strong>
            <small
              class="text-[11px] font-semibold truncate"
              :class="cardSecondaryTextClass(row.status, 'text-side')"
            >
              {{ row.subline }}
            </small>
          </div>

          <button
            type="button"
            class="flex items-center justify-center border-l border-solid border-lineDim bg-surface/50 cursor-pointer transition-colors hover:bg-surface hover:text-main focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-line"
            :class="cardSecondaryTextClass(row.status, 'text-main')"
            :aria-label="`Open Medical Manual procedures for ${row.title}`"
            :title="`Open Medical Manual: ${row.title}`"
            @click="openMedicalManual(row.manualQuery)"
          >
            <HelpCircle class="w-5 h-5" :stroke-width="2.2" aria-hidden="true" />
          </button>
        </article>
      </section>

      <aside class="hud-card p-5 flex flex-col items-center gap-3 text-center">
        <svg
          class="w-32 sm:w-36 h-auto text-side"
          viewBox="0 0 160 360"
          aria-label="Astronaut silhouette"
        >
          <g class="fill-surface stroke-main" stroke-width="2.5">
            <path
              d="M68 12h30l10 10 4 29-10 22-8 8 19 13 16 8 4 57 13 29-9 20-11-6-5-26-8-23 4 55-4 41-12 68-11 4-5-7 4-64-7-34-7 34 4 64-6 7-11-4-12-68-4-41 5-55-9 23-5 26-11 6-9-20 13-29 4-57 16-8 19-13-8-8-10-22 4-29 10-10Z"
            />
            <path d="M68 12h30l10 10-2 24-7 25-21 10-21-10-7-25-2-24 10-10Z" />
          </g>
          <g fill="none" class="stroke-side" stroke-width="1.5">
            <path
              d="M60 81l-15 18-8 58-17 29 8 20 13-6 7-27 11-24M100 81l15 18 8 58 17 29-8 20-13-6-7-27-11-24M51 181l8 54-4 41 12 68M109 181l-8 54 4 41-12 68M53 98h54M50 154h60M53 205h54M58 236h44"
            />
          </g>
        </svg>

        <h2 class="text-sm font-semibold text-main">{{ selectedAstronautName }}</h2>

        <div class="hud-pill border-lineDim px-3 py-1.5 flex flex-col items-start gap-1 text-xs">
          <span class="inline-flex items-center gap-1.5 text-warn">
            <AlertTriangle class="w-3 h-3 shrink-0" aria-hidden="true" />
            <span>Caution</span>
          </span>
          <span class="inline-flex items-center gap-1.5 text-crit">
            <AlertCircle class="w-3 h-3 shrink-0" aria-hidden="true" />
            <span>Warning</span>
          </span>
        </div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
@keyframes vital-frame-blink-crit {
  0%,
  100% {
    border-color: #f87171;
    box-shadow: 0 0 0 1.5px rgba(248, 113, 113, 0.75), 0 0 18px rgba(248, 113, 113, 0.42);
  }
  50% {
    border-color: rgba(248, 113, 113, 0.22);
    box-shadow: 0 0 0 0 rgba(248, 113, 113, 0), 0 0 0 rgba(248, 113, 113, 0);
  }
}

@keyframes vital-text-blink-crit {
  0%,
  100% {
    color: #f87171;
    text-shadow: 0 0 8px rgba(248, 113, 113, 0.45);
  }
  50% {
    color: rgba(248, 113, 113, 0.25);
    text-shadow: 0 0 0 rgba(248, 113, 113, 0);
  }
}

.vital-frame-warn {
  border-width: 2px;
  border-color: #facc15;
  box-shadow: 0 0 0 1px rgba(250, 204, 21, 0.35), 0 0 10px rgba(250, 204, 21, 0.16);
}

.vital-frame-crit {
  border-width: 2px;
  border-color: #f87171;
  animation: vital-frame-blink-crit 0.65s ease-in-out infinite;
}

.vital-text-crit {
  color: #f87171;
  animation: vital-text-blink-crit 0.65s ease-in-out infinite;
}
</style>
