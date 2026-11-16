<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  Activity,
  ArrowLeft,
  CalendarClock,
  HelpCircle,
  Home,
  Search,
  WifiOff,
} from "lucide-vue-next"
import {
  ReadingStatus,
  type AlertEvent,
  type LatestReading,
} from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type TelemetryDisplayStatus,
  type TelemetrySeriesPoint,
} from "@/composables/useTelemetry"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

type CadenceFilter = "all" | "continuous" | "periodic"
type CardStatus = TelemetryDisplayStatus

const CONTINUOUS_CREW_METRICS = [
  "resp_rate",
  "hrv_rmssd",
  "blood_glucose",
  "radiation_dose_cum",
  "radiation_rate",
] as const

const CONTINUOUS_ENV_METRICS = [
  "cabin_ppco2",
  "cabin_o2",
  "suit_pressure",
  "suit_temp",
] as const

const PERIODIC_CREW_METRICS = [
  "sleep_hours",
  "mood",
  "hydration_intake",
  "body_mass",
  "rbc_count",
  "wbc_count",
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
  getEnvSeries,
  resolveReadingStatus,
  formatMetricValue,
} = useTelemetry()
const {
  selectedAstronautId,
  lastTriggeredAlert,
  setSelectedAstronaut,
  openPrimaryVitals,
  goToHome,
} = useTelemetryNavigation()

const metricFilter = ref("")
const cadenceFilter = ref<CadenceFilter>("all")

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

const crewPool = computed(() =>
  activeAstronauts.value.length > 0 ? activeAstronauts.value : astronauts.value,
)

const activeAstronaut = computed(() => {
  const targetId = selectedAstronautId.value.trim()
  const pool = crewPool.value
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

function onAstronautSelect(event: Event) {
  const select = event.target as HTMLSelectElement
  const nextId = select.value.trim()
  if (!nextId) {
    return
  }
  const found = crewPool.value.find((a) => a.id === nextId)
  setSelectedAstronaut(nextId, found?.fullName)
  void router.replace({
    name: "other-stats",
    query: { ...route.query, astronaut: nextId },
  })
}

function loadCrewSecondaryTrends(astronautId: string) {
  const cleanId = astronautId.trim()
  if (!cleanId) {
    return
  }
  for (const metricCode of [...CONTINUOUS_CREW_METRICS, ...PERIODIC_CREW_METRICS]) {
    void loadMetricTrends(cleanId, metricCode, 40)
  }
}

function loadEnvironmentalSecondaryTrends() {
  for (const metricCode of CONTINUOUS_ENV_METRICS) {
    void loadMetricTrends("", metricCode, 40)
  }
}

watch(
  resolvedAstronautId,
  (astronautId) => {
    loadCrewSecondaryTrends(astronautId)
  },
  { immediate: true },
)

onMounted(() => {
  loadEnvironmentalSecondaryTrends()
  if (resolvedAstronautId.value.trim()) {
    loadCrewSecondaryTrends(resolvedAstronautId.value)
  }
})

const CARD_STATUS_SEVERITY: Record<CardStatus, number> = {
  unknown: 0,
  [ReadingStatus.$zero]: 0,
  [ReadingStatus.StatusNormal]: 1,
  [ReadingStatus.StatusStale]: 2,
  [ReadingStatus.StatusAbnormal]: 3,
  [ReadingStatus.StatusCritical]: 4,
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
    case ReadingStatus.StatusStale:
      return "text-stale"
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
    case ReadingStatus.StatusStale:
      return "border-2 border-dashed border-stale/75 opacity-85"
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
    default:
      return { stroke: nominalStroke, fill: nominalFill }
  }
}

function barFillClass(status: CardStatus, nominalFill: string): string {
  switch (status) {
    case ReadingStatus.StatusAbnormal:
      return "fill-warn/80"
    case ReadingStatus.StatusCritical:
      return "fill-crit/85"
    case ReadingStatus.StatusStale:
      return "fill-stale/45"
    case ReadingStatus.StatusNormal:
    default:
      return nominalFill
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
    return (parsedMs / 1000) * 14
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

function buildRespiratoryChart(
  series: TelemetrySeriesPoint[],
  fallbackRpm: number,
): ChartWaveData {
  const values = series.map((pt) => pt.value)
  const minVal = values.length > 0 ? Math.min(...values) : null
  const maxVal = values.length > 0 ? Math.max(...values) : null
  const activeValues = values.length > 0 ? values : [fallbackRpm]
  const latestRpm = activeValues[activeValues.length - 1]
  const cycleWidth = clamp((16 / clamp(latestRpm, 6, 36)) * 68, 32, 110)
  const phasePx = computeSeriesPhase(series) % cycleWidth

  const pts: ChartPointCoord[] = []
  for (let x = 0; x <= SVG_WIDTH; x += 4) {
    const tNorm = x / SVG_WIDTH
    const localRpm = sampleSeriesAt(activeValues, tNorm)
    const depthNorm = clamp((localRpm - 8) / 22, 0.25, 1)
    const u = ((x + phasePx) % cycleWidth) / cycleWidth
    const breath =
      u < 0.42
        ? Math.sin((u / 0.42) * Math.PI * 0.5)
        : Math.cos(((u - 0.42) / 0.58) * Math.PI * 0.5)
    const amp = 16 + depthNorm * 22
    const y = clamp(50 - Math.max(0, breath) * amp, 8, 56)
    pts.push({ x, y })
  }

  return {
    wavePath: smoothPathFromPoints(pts),
    minVal,
    maxVal,
  }
}

function buildContinuousTrendChart(
  series: TelemetrySeriesPoint[],
  minBound: number,
  maxBound: number,
  includeFill = true,
): ChartWaveData {
  const values = series.map((pt) => pt.value)
  if (values.length === 0) {
    return {
      wavePath: `M0 ${SVG_HEIGHT / 2}H${SVG_WIDTH}`,
      fillPath: includeFill ? `M0 ${SVG_HEIGHT / 2}H${SVG_WIDTH} V${SVG_HEIGHT} H0 Z` : undefined,
      minVal: null,
      maxVal: null,
    }
  }

  const minVal = Math.min(...values)
  const maxVal = Math.max(...values)
  const low = Math.min(minBound, minVal)
  const high = Math.max(maxBound, maxVal)
  const span = Math.max(0.001, high - low)
  const usableHeight = SVG_HEIGHT - PAD_Y * 2

  const toY = (val: number) =>
    roundCoord(clamp(SVG_HEIGHT - PAD_Y - ((val - low) / span) * usableHeight, PAD_Y, SVG_HEIGHT - PAD_Y))

  if (values.length === 1) {
    const y = toY(values[0])
    const wavePath = `M0 ${y}H${SVG_WIDTH}`
    return {
      wavePath,
      fillPath: includeFill ? `${wavePath} V${SVG_HEIGHT} H0 Z` : undefined,
      minVal,
      maxVal,
    }
  }

  const pts: ChartPointCoord[] = values.map((val, idx) => ({
    x: roundCoord((idx / (values.length - 1)) * SVG_WIDTH),
    y: toY(val),
  }))

  const wavePath = smoothPathFromPoints(pts)
  return {
    wavePath,
    fillPath: includeFill ? `${wavePath} V${SVG_HEIGHT} H0 Z` : undefined,
    minVal,
    maxVal,
  }
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

function formatReadout(
  reading: LatestReading | undefined,
  metricCode: string,
  fallbackUnit: string,
): {
  value: string
  unit: string
  status: CardStatus
  isDisconnected: boolean
  hasReading: boolean
} {
  const status = resolveReadingStatus(reading)
  const isDisconnected = status === ReadingStatus.StatusStale
  const metric = metricsByCode.value[metricCode]
  const unit = metric?.displayUnitSymbol ?? metric?.baseUnitSymbol ?? reading?.unitCode ?? fallbackUnit

  if (!reading || isDisconnected) {
    return {
      value: "—",
      unit,
      status,
      isDisconnected,
      hasReading: Boolean(reading),
    }
  }

  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === "unknown" || formatted.display === "—") {
    return {
      value: "—",
      unit,
      status,
      isDisconnected,
      hasReading: true,
    }
  }

  let valueOnly = formatted.display
  if (
    reading.valueNumeric !== null &&
    reading.valueNumeric !== undefined &&
    Number.isFinite(reading.valueNumeric)
  ) {
    valueOnly = reading.valueNumeric.toFixed(metric?.baseUnitDecimals ?? 0)
    if (unit === "%") {
      valueOnly = `${valueOnly}%`
    }
  } else if (reading.valueScale !== null && reading.valueScale !== undefined) {
    valueOnly = String(reading.valueScale)
  }

  return {
    value: valueOnly,
    unit: unit === "%" ? "" : unit,
    status: formatted.status,
    isDisconnected,
    hasReading: true,
  }
}

function formatCrewSubLine(
  astronautId: string,
  metricCode: string,
  prefix: string,
): { text: string; isDisconnected: boolean } {
  const reading = astronautId ? getCrewReading(astronautId, metricCode) : undefined
  const metric = metricsByCode.value[metricCode]
  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === ReadingStatus.StatusStale) {
    return { text: `${prefix} NO SIGNAL`, isDisconnected: true }
  }
  if (formatted.status === "unknown" || formatted.display === "—") {
    return { text: `${prefix} —`, isDisconnected: false }
  }
  return { text: `${prefix} ${formatted.display}`, isDisconnected: false }
}

function formatEnvSubLine(
  metricCode: string,
  prefix: string,
): { text: string; isDisconnected: boolean } {
  const reading = getEnvReading(metricCode)
  const metric = metricsByCode.value[metricCode]
  const formatted = formatMetricValue(reading, metric)
  if (formatted.status === ReadingStatus.StatusStale) {
    return { text: `${prefix} NO SIGNAL`, isDisconnected: true }
  }
  if (formatted.status === "unknown" || formatted.display === "—") {
    return { text: `${prefix} —`, isDisconnected: false }
  }
  return { text: `${prefix} ${formatted.display}`, isDisconnected: false }
}

function formatLastRecorded(reading: LatestReading | undefined): string {
  if (!reading || !reading.recordedAtUtc || reading.recordedAtUtc.startsWith("0001-01-01")) {
    return "Awaiting check-in"
  }
  const metText = reading.met && reading.met.trim() !== "" ? `MET ${reading.met.trim()}` : ""
  const parsed = Date.parse(reading.recordedAtUtc)
  let utcText = ""
  if (Number.isFinite(parsed)) {
    const d = new Date(parsed)
    const hh = String(d.getUTCHours()).padStart(2, "0")
    const mm = String(d.getUTCMinutes()).padStart(2, "0")
    const day = String(d.getUTCDate()).padStart(2, "0")
    const mon = d.toLocaleString("en-US", { month: "short", timeZone: "UTC" })
    utcText = `${day} ${mon} ${hh}:${mm} UTC`
  }
  if (metText && utcText) {
    return `${metText} · ${utcText}`
  }
  return metText || utcText || "Recorded"
}

interface ContinuousCardRow {
  id: string
  title: string
  subtitle: string
  manualQuery: string
  keywords: string
  value: string
  unit: string
  subline: string
  status: CardStatus
  isDisconnected: boolean
  isSublineDisconnected: boolean
  rangeLabel: string
  chart: ChartWaveData
  strokeClass: string
  fillClass: string
}

const continuousRows = computed<ContinuousCardRow[]>(() => {
  const id = resolvedAstronautId.value

  // 1. Respiratory rate
  const respReading = id ? getCrewReading(id, "resp_rate") : undefined
  const respOut = formatReadout(respReading, "resp_rate", "rpm")
  const respStatus = resolveCardStatus(id, respOut.status, ["resp_rate"])
  const respSeries = id ? getCrewSeries(id, "resp_rate") : []
  const respChart = buildRespiratoryChart(respSeries, respReading?.valueNumeric ?? 16)
  const respStyles = vitalChartStyles("stroke-ok", "fill-none", respStatus)

  // 2. Heart rate variability (RMSSD)
  const hrvReading = id ? getCrewReading(id, "hrv_rmssd") : undefined
  const hrvOut = formatReadout(hrvReading, "hrv_rmssd", "ms")
  const hrvStatus = resolveCardStatus(id, hrvOut.status, ["hrv_rmssd"])
  const hrvSeries = id ? getCrewSeries(id, "hrv_rmssd") : []
  const hrvChart = buildContinuousTrendChart(hrvSeries, 10, 90, true)
  const hrvStyles = vitalChartStyles("stroke-ok", "fill-ok/15", hrvStatus)

  // 3. Blood glucose
  const gluReading = id ? getCrewReading(id, "blood_glucose") : undefined
  const gluOut = formatReadout(gluReading, "blood_glucose", "mg/dL")
  const gluStatus = resolveCardStatus(id, gluOut.status, ["blood_glucose"])
  const gluSeries = id ? getCrewSeries(id, "blood_glucose") : []
  const gluChart = buildContinuousTrendChart(gluSeries, 55, 180, true)
  const gluStyles = vitalChartStyles("stroke-main", "fill-main/10", gluStatus)

  // 4. Cabin ppCO2 (paired with cabin_o2)
  const co2Reading = getEnvReading("cabin_ppco2")
  const co2Out = formatReadout(co2Reading, "cabin_ppco2", "mmHg")
  const co2Status = resolveCardStatus(id, co2Out.status, [], ["cabin_ppco2", "cabin_o2"])
  const o2Sub = formatEnvSubLine("cabin_o2", "O2:")
  const co2Series = getEnvSeries("cabin_ppco2")
  const co2Chart = buildContinuousTrendChart(co2Series, 0, 5, true)
  const co2Styles = vitalChartStyles("stroke-ok", "fill-ok/15", co2Status)

  // 5. Suit pressure (paired with suit_temp)
  const suitReading = getEnvReading("suit_pressure")
  const suitOut = formatReadout(suitReading, "suit_pressure", "kPa")
  const suitStatus = resolveCardStatus(id, suitOut.status, [], ["suit_pressure", "suit_temp"])
  const suitTempSub = formatEnvSubLine("suit_temp", "TEMP:")
  const suitSeries = getEnvSeries("suit_pressure")
  const suitChart = buildContinuousTrendChart(suitSeries, 25, 33, true)
  const suitStyles = vitalChartStyles("stroke-main", "fill-main/10", suitStatus)

  // 6. Cumulative radiation dose (paired with radiation_rate)
  const radCumReading = id ? getCrewReading(id, "radiation_dose_cum") : undefined
  const radCumOut = formatReadout(radCumReading, "radiation_dose_cum", "mSv")
  const radCumStatus = resolveCardStatus(id, radCumOut.status, ["radiation_dose_cum", "radiation_rate"])
  const radRateSub = formatCrewSubLine(id, "radiation_rate", "RATE:")
  const radCumSeries = id ? getCrewSeries(id, "radiation_dose_cum") : []
  const radCumChart = buildContinuousTrendChart(radCumSeries, 0, 25, true)
  const radCumStyles = vitalChartStyles("stroke-warn", "fill-warn/15", radCumStatus)

  const all: ContinuousCardRow[] = [
    {
      id: "resp_rate",
      title: "RESPIRATORY RATE",
      subtitle: "CREW",
      manualQuery: "resp_rate",
      keywords: "resp_rate breathing ventilation rpm",
      value: respOut.value,
      unit: respOut.unit,
      subline: "TARGET 10–20 rpm",
      status: respStatus,
      isDisconnected: respOut.isDisconnected,
      isSublineDisconnected: false,
      rangeLabel: respOut.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(respChart.minVal, respChart.maxVal, "resp_rate", respOut.unit),
      chart: respChart,
      strokeClass: respStyles.stroke,
      fillClass: respStyles.fill,
    },
    {
      id: "hrv_rmssd",
      title: "HEART RATE VARIABILITY",
      subtitle: "HRV",
      manualQuery: "hrv_rmssd",
      keywords: "hrv_rmssd autonomic stress recovery rmssd ms",
      value: hrvOut.value,
      unit: hrvOut.unit,
      subline: "TARGET ≥ 20 ms",
      status: hrvStatus,
      isDisconnected: hrvOut.isDisconnected,
      isSublineDisconnected: false,
      rangeLabel: hrvOut.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(hrvChart.minVal, hrvChart.maxVal, "hrv_rmssd", hrvOut.unit),
      chart: hrvChart,
      strokeClass: hrvStyles.stroke,
      fillClass: hrvStyles.fill,
    },
    {
      id: "blood_glucose",
      title: "BLOOD GLUCOSE",
      subtitle: "METABOLIC",
      manualQuery: "blood_glucose",
      keywords: "blood_glucose glycemic metabolic mg/dl",
      value: gluOut.value,
      unit: gluOut.unit,
      subline: "TARGET 70–140 mg/dL",
      status: gluStatus,
      isDisconnected: gluOut.isDisconnected,
      isSublineDisconnected: false,
      rangeLabel: gluOut.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(gluChart.minVal, gluChart.maxVal, "blood_glucose", gluOut.unit),
      chart: gluChart,
      strokeClass: gluStyles.stroke,
      fillClass: gluStyles.fill,
    },
    {
      id: "cabin_ppco2",
      title: "CABIN ppCO₂",
      subtitle: "ENV",
      manualQuery: "cabin_ppco2",
      keywords: "cabin_ppco2 cabin_o2 co2 atmosphere scrubber mmhg",
      value: co2Out.value,
      unit: co2Out.unit,
      subline: o2Sub.text,
      status: co2Status,
      isDisconnected: co2Out.isDisconnected,
      isSublineDisconnected: o2Sub.isDisconnected,
      rangeLabel: co2Out.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(co2Chart.minVal, co2Chart.maxVal, "cabin_ppco2", co2Out.unit),
      chart: co2Chart,
      strokeClass: co2Styles.stroke,
      fillClass: co2Styles.fill,
    },
    {
      id: "suit_pressure",
      title: "SUIT PRESSURE",
      subtitle: "EVA",
      manualQuery: "suit_pressure",
      keywords: "suit_pressure suit_temp eva spacesuit thermal kpa",
      value: suitOut.value,
      unit: suitOut.unit,
      subline: suitTempSub.text,
      status: suitStatus,
      isDisconnected: suitOut.isDisconnected,
      isSublineDisconnected: suitTempSub.isDisconnected,
      rangeLabel: suitOut.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(suitChart.minVal, suitChart.maxVal, "suit_pressure", suitOut.unit),
      chart: suitChart,
      strokeClass: suitStyles.stroke,
      fillClass: suitStyles.fill,
    },
    {
      id: "radiation_dose_cum",
      title: "CUMULATIVE DOSE",
      subtitle: "RADIATION",
      manualQuery: "radiation_dose_cum",
      keywords: "radiation_dose_cum radiation_rate dosimeter exposure msv",
      value: radCumOut.value,
      unit: radCumOut.unit,
      subline: radRateSub.text,
      status: radCumStatus,
      isDisconnected: radCumOut.isDisconnected,
      isSublineDisconnected: radRateSub.isDisconnected,
      rangeLabel: radCumOut.isDisconnected
        ? "NO SIGNAL"
        : formatRangeLabel(radCumChart.minVal, radCumChart.maxVal, "radiation_dose_cum", radCumOut.unit),
      chart: radCumChart,
      strokeClass: radCumStyles.stroke,
      fillClass: radCumStyles.fill,
    },
  ]

  const q = metricFilter.value.trim().toLowerCase()
  if (!q) {
    return all
  }
  return all.filter(
    (row) =>
      row.title.toLowerCase().includes(q) ||
      row.subtitle.toLowerCase().includes(q) ||
      row.subline.toLowerCase().includes(q) ||
      row.keywords.toLowerCase().includes(q),
  )
})

interface CheckpointBar {
  x: number
  y: number
  width: number
  height: number
  fillClass: string
}

function buildCheckpointBars(
  series: TelemetrySeriesPoint[],
  minScale: number,
  maxScale: number,
  nominalFill: string,
): {
  bars: CheckpointBar[]
  targetBandY?: { top: number; height: number }
} {
  const recent = series.slice(-14)
  if (recent.length === 0) {
    return { bars: [] }
  }

  const values = recent.map((pt) => pt.value)
  const low = Math.min(minScale, ...values)
  const high = Math.max(maxScale, ...values)
  const span = Math.max(0.001, high - low)
  const slotWidth = SVG_WIDTH / Math.max(recent.length, 6)
  const barWidth = clamp(slotWidth * 0.56, 8, 22)
  const usableHeight = SVG_HEIGHT - 12

  const bars: CheckpointBar[] = recent.map((pt, idx) => {
    const norm = clamp((pt.value - low) / span, 0.08, 1)
    const height = roundCoord(norm * usableHeight)
    const x = roundCoord(idx * slotWidth + (slotWidth - barWidth) / 2)
    const y = roundCoord(SVG_HEIGHT - 4 - height)
    return {
      x,
      y,
      width: roundCoord(barWidth),
      height,
      fillClass: barFillClass(pt.status === "unknown" ? ReadingStatus.StatusNormal : pt.status, nominalFill),
    }
  })

  return { bars }
}

interface DualBarPair {
  x1: number
  y1: number
  h1: number
  fill1: string
  x2: number
  y2: number
  h2: number
  fill2: string
  barWidth: number
}

function buildDualCheckpointBars(
  seriesA: TelemetrySeriesPoint[],
  minA: number,
  maxA: number,
  seriesB: TelemetrySeriesPoint[],
  minB: number,
  maxB: number,
): DualBarPair[] {
  const recentA = seriesA.slice(-8)
  const recentB = seriesB.slice(-8)
  const count = Math.max(recentA.length, recentB.length)
  if (count === 0) {
    return []
  }

  const slotWidth = SVG_WIDTH / Math.max(count, 5)
  const barWidth = clamp(slotWidth * 0.3, 6, 14)
  const usableHeight = SVG_HEIGHT - 12

  const toHeight = (val: number, minBound: number, maxBound: number) => {
    const span = Math.max(0.001, maxBound - minBound)
    return roundCoord(clamp((val - minBound) / span, 0.1, 1) * usableHeight)
  }

  const pairs: DualBarPair[] = []
  for (let i = 0; i < count; i += 1) {
    const ptA = recentA[i]
    const ptB = recentB[i]
    const center = i * slotWidth + slotWidth / 2
    const h1 = ptA ? toHeight(ptA.value, minA, maxA) : 0
    const h2 = ptB ? toHeight(ptB.value, minB, maxB) : 0
    pairs.push({
      x1: roundCoord(center - barWidth - 1.5),
      y1: roundCoord(SVG_HEIGHT - 4 - h1),
      h1,
      fill1: ptA ? barFillClass(ptA.status === "unknown" ? ReadingStatus.StatusNormal : ptA.status, "fill-ok/75") : "fill-none",
      x2: roundCoord(center + 1.5),
      y2: roundCoord(SVG_HEIGHT - 4 - h2),
      h2,
      fill2: ptB ? barFillClass(ptB.status === "unknown" ? ReadingStatus.StatusNormal : ptB.status, "fill-main/75") : "fill-none",
      barWidth: roundCoord(barWidth),
    })
  }
  return pairs
}

type PeriodicVisualKind = "bars" | "mood-scale" | "dual-bars" | "ultrasound"

interface PeriodicCardRow {
  id: string
  kind: PeriodicVisualKind
  title: string
  subtitle?: string
  cadencePill: string
  targetLabel: string
  manualQuery: string
  keywords: string
  value: string
  unit: string
  primaryStatus?: CardStatus
  secondaryValue?: string
  secondaryLabel?: string
  secondaryStatus?: CardStatus
  primaryLabel?: string
  lastRecorded: string
  status: CardStatus
  isDisconnected: boolean
  bars?: CheckpointBar[]
  dualBars?: DualBarPair[]
  moodScore?: number | null
  findingText?: string
  findingMeta?: string
  qualitativeLabel?: string
}

const periodicRows = computed<PeriodicCardRow[]>(() => {
  const id = resolvedAstronautId.value

  // 1. Sleep hours
  const sleepReading = id ? getCrewReading(id, "sleep_hours") : undefined
  const sleepOut = formatReadout(sleepReading, "sleep_hours", "h")
  const sleepStatus = resolveCardStatus(id, sleepOut.status, ["sleep_hours"])
  const sleepSeries = id ? getCrewSeries(id, "sleep_hours") : []
  const sleepChart = buildCheckpointBars(sleepSeries, 4, 11, "fill-ok/75")

  // 2. Mood scale (1-5)
  const moodReading = id ? getCrewReading(id, "mood") : undefined
  const moodOut = formatReadout(moodReading, "mood", "/ 5")
  const moodStatus = resolveCardStatus(id, moodOut.status, ["mood"])
  const rawMood =
    moodReading?.valueScale ??
    (moodReading?.valueNumeric !== null && moodReading?.valueNumeric !== undefined
      ? Math.round(moodReading.valueNumeric)
      : null)
  const moodScore = !moodOut.isDisconnected && rawMood !== null ? clamp(rawMood, 1, 5) : null

  // 3. Hydration intake
  const hydReading = id ? getCrewReading(id, "hydration_intake") : undefined
  const hydOut = formatReadout(hydReading, "hydration_intake", "mL")
  const hydStatus = resolveCardStatus(id, hydOut.status, ["hydration_intake"])
  const hydSeries = id ? getCrewSeries(id, "hydration_intake") : []
  const hydChart = buildCheckpointBars(hydSeries, 500, 4000, "fill-ok/75")

  // 4. Body mass
  const massReading = id ? getCrewReading(id, "body_mass") : undefined
  const massOut = formatReadout(massReading, "body_mass", "kg")
  const massStatus = resolveCardStatus(id, massOut.status, ["body_mass"])
  const massSeries = id ? getCrewSeries(id, "body_mass") : []
  const massChart = buildCheckpointBars(massSeries, 40, 115, "fill-main/75")

  // 5. Dual Hematology Summary (rbc_count + wbc_count)
  const rbcReading = id ? getCrewReading(id, "rbc_count") : undefined
  const wbcReading = id ? getCrewReading(id, "wbc_count") : undefined
  const rbcOut = formatReadout(rbcReading, "rbc_count", "10⁶/µL")
  const wbcOut = formatReadout(wbcReading, "wbc_count", "10³/µL")
  const hemaPrimaryStatus = pickWorstCardStatus(rbcOut.status, wbcOut.status)
  const hemaStatus = resolveCardStatus(id, hemaPrimaryStatus, ["rbc_count", "wbc_count"])
  const hemaDisconnected =
    (rbcOut.isDisconnected && wbcOut.isDisconnected) ||
    (rbcOut.isDisconnected && !wbcReading) ||
    (wbcOut.isDisconnected && !rbcReading)
  const rbcSeries = id ? getCrewSeries(id, "rbc_count") : []
  const wbcSeries = id ? getCrewSeries(id, "wbc_count") : []
  const dualBars = buildDualCheckpointBars(rbcSeries, 2.5, 7.5, wbcSeries, 2, 16)
  const latestHemaReading =
    rbcReading && wbcReading
      ? Date.parse(rbcReading.recordedAtUtc) >= Date.parse(wbcReading.recordedAtUtc)
        ? rbcReading
        : wbcReading
      : rbcReading ?? wbcReading

  // 6. Diagnostic Ultrasound
  const usReading = id ? getCrewReading(id, "ultrasound") : undefined
  const usRawStatus = resolveReadingStatus(usReading)
  const usDisconnected = usRawStatus === ReadingStatus.StatusStale
  const usStatus = resolveCardStatus(id, usRawStatus, ["ultrasound"])
  const usText =
    usReading?.valueText?.trim() ||
    usReading?.note?.trim() ||
    "No scheduled ultrasound study recorded yet"
  const usQualitative = usDisconnected
    ? "NO SIGNAL"
    : usStatus === ReadingStatus.StatusCritical
      ? "CRITICAL"
      : usStatus === ReadingStatus.StatusAbnormal
        ? "REVIEW"
        : usReading
          ? "NOMINAL"
          : "PENDING"
  const usMeta = usReading
    ? `${usReading.deviceCode || "ultrasound_01"} · ${(usReading.environment || "iva").toUpperCase()}`
    : "ultrasound_01 · MANUAL REVIEW"

  const all: PeriodicCardRow[] = [
    {
      id: "sleep_hours",
      kind: "bars",
      title: "SLEEP DURATION",
      cadencePill: "DAILY LOG",
      targetLabel: "TARGET 7.0–9.0 h",
      manualQuery: "sleep_hours",
      keywords: "sleep_hours rest fatigue circadian daily",
      value: sleepOut.value,
      unit: sleepOut.unit,
      lastRecorded: sleepOut.isDisconnected ? "NO SIGNAL" : formatLastRecorded(sleepReading),
      status: sleepStatus,
      isDisconnected: sleepOut.isDisconnected,
      bars: sleepChart.bars,
    },
    {
      id: "mood",
      kind: "mood-scale",
      title: "MOOD & WELLBEING",
      cadencePill: "DAILY CHECK-IN",
      targetLabel: "TARGET 2–4 / 5",
      manualQuery: "mood",
      keywords: "mood behavioral psychological isolation scale score",
      value: moodOut.value,
      unit: "/ 5",
      lastRecorded: moodOut.isDisconnected ? "NO SIGNAL" : formatLastRecorded(moodReading),
      status: moodStatus,
      isDisconnected: moodOut.isDisconnected,
      moodScore,
    },
    {
      id: "hydration_intake",
      kind: "bars",
      title: "HYDRATION INTAKE",
      cadencePill: "DAILY LOG",
      targetLabel: "TARGET 1500–3500 mL",
      manualQuery: "hydration_intake",
      keywords: "hydration_intake fluid water intake ml",
      value: hydOut.value,
      unit: hydOut.unit,
      lastRecorded: hydOut.isDisconnected ? "NO SIGNAL" : formatLastRecorded(hydReading),
      status: hydStatus,
      isDisconnected: hydOut.isDisconnected,
      bars: hydChart.bars,
    },
    {
      id: "body_mass",
      kind: "bars",
      title: "BODY MASS",
      cadencePill: "DAILY CHECK-IN",
      targetLabel: "TARGET 45–110 kg",
      manualQuery: "body_mass",
      keywords: "body_mass weight nutrition mass kg",
      value: massOut.value,
      unit: massOut.unit,
      lastRecorded: massOut.isDisconnected ? "NO SIGNAL" : formatLastRecorded(massReading),
      status: massStatus,
      isDisconnected: massOut.isDisconnected,
      bars: massChart.bars,
    },
    {
      id: "hematology",
      kind: "dual-bars",
      title: "HEMATOLOGY PANEL",
      subtitle: "RBC · WBC",
      cadencePill: "LAB ASSAY",
      targetLabel: "RBC 4.0–6.0 · WBC 4.0–11.0",
      manualQuery: "rbc_count",
      keywords: "rbc_count wbc_count hematology blood cell lab assay",
      value: rbcOut.isDisconnected ? "—" : rbcOut.value,
      unit: rbcOut.unit,
      primaryLabel: "RBC",
      primaryStatus: rbcOut.status,
      secondaryLabel: "WBC",
      secondaryValue: wbcOut.isDisconnected ? "—" : `${wbcOut.value} ${wbcOut.unit}`,
      secondaryStatus: wbcOut.status,
      lastRecorded: hemaDisconnected ? "NO SIGNAL" : formatLastRecorded(latestHemaReading),
      status: hemaStatus,
      isDisconnected: hemaDisconnected,
      dualBars,
    },
    {
      id: "ultrasound",
      kind: "ultrasound",
      title: "DIAGNOSTIC ULTRASOUND",
      cadencePill: "SCHEDULED IMAGING",
      targetLabel: "MANUAL REVIEW",
      manualQuery: "ultrasound",
      keywords: "ultrasound imaging scan ocular sans diagnostic",
      value: usQualitative,
      unit: "",
      lastRecorded: usDisconnected ? "NO SIGNAL" : formatLastRecorded(usReading),
      status: usStatus,
      isDisconnected: usDisconnected,
      findingText: usText,
      findingMeta: usMeta,
      qualitativeLabel: usQualitative,
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
      row.cadencePill.toLowerCase().includes(q) ||
      row.targetLabel.toLowerCase().includes(q) ||
      (row.findingText?.toLowerCase().includes(q) ?? false) ||
      row.keywords.toLowerCase().includes(q),
  )
})

const showContinuousSection = computed(
  () => cadenceFilter.value !== "periodic" && continuousRows.value.length > 0,
)
const showPeriodicSection = computed(
  () => cadenceFilter.value !== "continuous" && periodicRows.value.length > 0,
)
const hasVisibleCards = computed(
  () => showContinuousSection.value || showPeriodicSection.value,
)

function clearFilters() {
  metricFilter.value = ""
  cadenceFilter.value = "all"
}
</script>

<template>
  <main
    class="flex-1 w-full bg-ground text-main px-4 sm:px-8 py-6 flex flex-col items-center gap-6"
    aria-label="Other Stats Dashboard"
  >
    <!-- Top Controls Bar -->
    <header class="w-full max-w-5xl flex flex-wrap items-center gap-3">
      <button
        type="button"
        class="button-side h-10 px-3 text-xs gap-1.5 shrink-0"
        aria-label="Return to Home"
        @click="goToHome"
      >
        <Home class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Home</span>
      </button>

      <button
        type="button"
        class="button-side h-10 px-3 text-xs gap-1.5 shrink-0"
        aria-label="Return to Primary Vitals for selected astronaut"
        @click="openPrimaryVitals(activeAstronaut ?? resolvedAstronautId)"
      >
        <ArrowLeft class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Primary Vitals</span>
      </button>

      <h1 class="text-lg sm:text-xl font-semibold tracking-wide text-main shrink-0">
        OTHER STATS
      </h1>

      <div v-if="crewPool.length > 0" class="shrink-0">
        <label class="sr-only" for="secondary-crew-select">Select crew member</label>
        <select
          id="secondary-crew-select"
          :value="resolvedAstronautId"
          aria-label="Select crew member"
          class="field h-10 px-3 text-xs sm:text-sm min-w-40 cursor-pointer"
          @change="onAstronautSelect"
        >
          <option
            v-for="astro in crewPool"
            :key="astro.id"
            :value="astro.id"
            class="bg-ground text-main"
          >
            {{ astro.fullName }}
          </option>
        </select>
      </div>

      <div class="relative flex-1 min-w-44 flex items-center">
        <Search class="w-4 h-4 text-side absolute left-3 pointer-events-none" aria-hidden="true" />
        <input
          v-model="metricFilter"
          type="text"
          aria-label="Filter other metrics"
          placeholder="Search other stats metrics"
          class="field w-full h-10 pl-9 pr-3 text-sm"
        />
      </div>

      <div
        class="inline-flex items-center gap-1 p-1 rounded-xl border border-solid border-lineDim bg-surface shrink-0"
        role="group"
        aria-label="Filter by monitoring cadence"
      >
        <button
          v-for="option in [
            { id: 'all', label: 'All' },
            { id: 'continuous', label: 'Continuous' },
            { id: 'periodic', label: 'Periodic' },
          ] as const"
          :key="option.id"
          type="button"
          class="h-8 px-3 rounded-lg text-xs font-medium transition-colors cursor-pointer"
          :class="
            cadenceFilter === option.id
              ? 'bg-main text-ground'
              : 'bg-transparent text-side hover:text-main hover:bg-ground/60'
          "
          :aria-pressed="cadenceFilter === option.id"
          @click="cadenceFilter = option.id"
        >
          {{ option.label }}
        </button>
      </div>
    </header>

    <!-- Section 1: Continuous Telemetry -->
    <section
      v-if="showContinuousSection"
      class="w-full max-w-5xl flex flex-col gap-3"
      aria-label="Continuous Telemetry"
    >
      <div class="flex flex-wrap items-center justify-between gap-2 px-1">
        <div class="flex items-center gap-2">
          <Activity class="w-4 h-4 text-ok shrink-0" aria-hidden="true" />
          <h2 class="text-xs font-bold tracking-widest text-main">
            CONTINUOUS MONITORING · LIVE STREAM
          </h2>
        </div>
        <span class="text-[0.6875rem] font-medium tracking-wide text-dull">
          Real-time subsystem &amp; metabolic stream
        </span>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <article
          v-for="row in continuousRows"
          :key="row.id"
          class="hud-card overflow-hidden grid grid-cols-[minmax(0,1fr)_9.5rem_2.25rem] sm:grid-cols-[minmax(0,1fr)_11.25rem_2.75rem] min-h-24"
          :class="cardFrameClass(row.status)"
        >
          <!-- Waveform Panel -->
          <div class="bg-surface border-r border-solid border-lineDim flex flex-col justify-between min-w-0">
            <div class="px-3 pt-2 flex items-center justify-between gap-2 text-[0.6875rem] font-bold tracking-wider">
              <div class="flex items-center gap-2 min-w-0">
                <h3 :class="statusTextClass(row.status)" class="truncate">{{ row.title }}</h3>
                <span
                  class="shrink-0"
                  :class="cardSecondaryTextClass(row.status, 'text-main')"
                >
                  {{ row.subtitle }}
                </span>
              </div>
              <span
                class="text-[0.625rem] font-medium hud-value tabular-nums shrink-0"
                :class="row.isDisconnected ? 'text-stale' : cardSecondaryTextClass(row.status, 'text-dull')"
              >
                {{ row.rangeLabel }}
              </span>
            </div>

            <div class="relative w-full h-14 px-1">
              <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-full" aria-hidden="true">
                <line
                  v-if="row.isDisconnected"
                  x1="0"
                  y1="32"
                  x2="300"
                  y2="32"
                  stroke-dasharray="6 5"
                  stroke-width="2"
                  vector-effect="non-scaling-stroke"
                  class="stroke-stale/60"
                />
                <template v-else>
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
                    stroke-linejoin="round"
                    vector-effect="non-scaling-stroke"
                    :class="row.strokeClass"
                  />
                </template>
              </svg>
              <div
                v-if="row.isDisconnected"
                class="absolute inset-0 flex items-center justify-center pointer-events-none px-2"
              >
                <span
                  class="inline-flex items-center gap-1.5 max-w-full px-2.5 py-0.5 text-[0.6875rem] font-semibold tracking-wide text-stale border border-dashed border-stale/75 bg-ground/90 rounded-md"
                >
                  <WifiOff class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
                  <span class="truncate">Sensor Disconnected · No Signal</span>
                </span>
              </div>
            </div>
          </div>

          <!-- Readout Column -->
          <div class="px-3 py-2 flex flex-col justify-between hud-value min-w-0">
            <div
              class="flex items-baseline justify-between gap-1 text-[0.6875rem] font-semibold"
              :class="row.isDisconnected ? 'text-stale' : cardSecondaryTextClass(row.status, 'text-side')"
            >
              <span class="text-[0.625rem] tracking-wider uppercase text-dull">LIVE</span>
              <span>{{ row.unit }}</span>
            </div>
            <strong
              class="text-2xl sm:text-[1.75rem] font-bold leading-none whitespace-nowrap"
              :class="row.isDisconnected ? 'text-stale' : statusTextClass(row.status)"
            >
              {{ row.isDisconnected ? "—" : row.value }}
            </strong>
            <small
              class="text-[0.6875rem] font-semibold inline-flex items-center gap-1 min-w-0"
              :class="
                row.isDisconnected || row.isSublineDisconnected
                  ? 'text-stale'
                  : cardSecondaryTextClass(row.status, 'text-side')
              "
            >
              <WifiOff
                v-if="row.isDisconnected || row.isSublineDisconnected"
                class="w-3 h-3 shrink-0"
                aria-hidden="true"
              />
              <span class="truncate">{{ row.isDisconnected ? "NO SIGNAL" : row.subline }}</span>
            </small>
          </div>

          <!-- Medical Manual Button -->
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
      </div>
    </section>

    <!-- Section 2: Periodic & Check-In Stats -->
    <section
      v-if="showPeriodicSection"
      class="w-full max-w-5xl flex flex-col gap-3"
      aria-label="Periodic and Check-In Stats"
    >
      <div class="flex flex-wrap items-center justify-between gap-2 px-1">
        <div class="flex items-center gap-2">
          <CalendarClock class="w-4 h-4 text-side shrink-0" aria-hidden="true" />
          <h2 class="text-xs font-bold tracking-widest text-main">
            PERIODIC &amp; CHECK-IN STATS · NON-SECOND CADENCE
          </h2>
        </div>
        <span class="text-[0.6875rem] font-medium tracking-wide text-dull">
          Daily logs, behavioral check-ins, lab assays &amp; scheduled imaging
        </span>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-3">
        <article
          v-for="row in periodicRows"
          :key="row.id"
          class="hud-card overflow-hidden grid grid-cols-[minmax(0,1fr)_10rem_2.25rem] sm:grid-cols-[minmax(0,1fr)_11.75rem_2.75rem] min-h-28"
          :class="cardFrameClass(row.status)"
        >
          <!-- Visualization / Checkpoint Panel -->
          <div class="bg-surface border-r border-solid border-lineDim flex flex-col justify-between min-w-0">
            <div class="px-3 pt-2 flex items-center justify-between gap-2 text-[0.6875rem] font-bold tracking-wider">
              <div class="flex items-center gap-2 min-w-0">
                <h3 :class="statusTextClass(row.status)" class="truncate">{{ row.title }}</h3>
                <span
                  v-if="row.subtitle"
                  class="shrink-0"
                  :class="cardSecondaryTextClass(row.status, 'text-main')"
                >
                  {{ row.subtitle }}
                </span>
              </div>
              <span
                class="px-1.5 py-0.5 rounded border border-solid border-lineDim bg-ground text-[0.5625rem] font-semibold tracking-wider text-side shrink-0"
              >
                {{ row.cadencePill }}
              </span>
            </div>

            <!-- Disconnected Overlay -->
            <div v-if="row.isDisconnected" class="relative w-full h-15 px-1">
              <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-full" aria-hidden="true">
                <line
                  x1="0"
                  y1="32"
                  x2="300"
                  y2="32"
                  stroke-dasharray="6 5"
                  stroke-width="2"
                  vector-effect="non-scaling-stroke"
                  class="stroke-stale/60"
                />
              </svg>
              <div class="absolute inset-0 flex items-center justify-center pointer-events-none px-2">
                <span
                  class="inline-flex items-center gap-1.5 max-w-full px-2.5 py-0.5 text-[0.6875rem] font-semibold tracking-wide text-stale border border-dashed border-stale/75 bg-ground/90 rounded-md"
                >
                  <WifiOff class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
                  <span class="truncate">Sensor Disconnected · No Signal</span>
                </span>
              </div>
            </div>

            <!-- Mood 1-5 Scale Indicator -->
            <div
              v-else-if="row.kind === 'mood-scale'"
              class="px-3 py-2 flex flex-col justify-center gap-1.5"
            >
              <div class="grid grid-cols-5 gap-1.5" role="img" :aria-label="`Mood scale ${row.moodScore ?? 'unrecorded'} of 5`">
                <div
                  v-for="step in 5"
                  :key="step"
                  class="h-6 rounded border border-solid flex items-center justify-center text-[0.6875rem] font-bold tabular-nums transition-colors"
                  :class="
                    row.moodScore !== null && row.moodScore !== undefined && step <= row.moodScore
                      ? row.status === ReadingStatus.StatusCritical
                        ? 'border-crit bg-crit/25 text-crit'
                        : row.status === ReadingStatus.StatusAbnormal
                          ? 'border-warn bg-warn/25 text-warn'
                          : 'border-ok bg-ok/20 text-ok'
                      : 'border-lineDim bg-ground text-dull'
                  "
                >
                  {{ step }}
                </div>
              </div>
              <div class="flex items-center justify-between text-[0.625rem] text-dull">
                <span>1 Low</span>
                <span>{{ row.targetLabel }}</span>
                <span>5 High</span>
              </div>
            </div>

            <!-- Ultrasound Qualitative Card -->
            <div
              v-else-if="row.kind === 'ultrasound'"
              class="px-3 py-2 flex flex-col justify-center gap-1 min-w-0"
            >
              <p class="text-xs font-medium text-main leading-snug line-clamp-2">
                {{ row.findingText }}
              </p>
              <div class="flex items-center justify-between gap-2 text-[0.625rem] text-dull">
                <span class="truncate uppercase tracking-wide">{{ row.findingMeta }}</span>
                <span class="shrink-0">{{ row.targetLabel }}</span>
              </div>
            </div>

            <!-- Dual Hematology Checkpoint Bars -->
            <div
              v-else-if="row.kind === 'dual-bars'"
              class="px-2 pb-1.5 flex flex-col justify-end gap-1"
            >
              <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-13" aria-hidden="true">
                <line x1="0" y1="60" x2="300" y2="60" stroke-width="1" class="stroke-lineDim" />
                <template v-if="row.dualBars && row.dualBars.length > 0">
                  <g v-for="(pair, idx) in row.dualBars" :key="idx">
                    <rect
                      v-if="pair.h1 > 0"
                      :x="pair.x1"
                      :y="pair.y1"
                      :width="pair.barWidth"
                      :height="pair.h1"
                      rx="2"
                      :class="pair.fill1"
                    />
                    <rect
                      v-if="pair.h2 > 0"
                      :x="pair.x2"
                      :y="pair.y2"
                      :width="pair.barWidth"
                      :height="pair.h2"
                      rx="2"
                      :class="pair.fill2"
                    />
                  </g>
                </template>
                <line
                  v-else
                  x1="0"
                  y1="32"
                  x2="300"
                  y2="32"
                  stroke-dasharray="4 4"
                  stroke-width="1.5"
                  class="stroke-lineDim"
                />
              </svg>
              <div class="px-1 flex items-center justify-between text-[0.625rem] text-dull">
                <span>{{ row.targetLabel }}</span>
                <span class="shrink-0">CHECKPOINTS</span>
              </div>
            </div>

            <!-- Discrete Checkpoint Bars -->
            <div v-else class="px-2 pb-1.5 flex flex-col justify-end gap-1">
              <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-13" aria-hidden="true">
                <line x1="0" y1="60" x2="300" y2="60" stroke-width="1" class="stroke-lineDim" />
                <template v-if="row.bars && row.bars.length > 0">
                  <rect
                    v-for="(bar, idx) in row.bars"
                    :key="idx"
                    :x="bar.x"
                    :y="bar.y"
                    :width="bar.width"
                    :height="bar.height"
                    rx="2"
                    :class="bar.fillClass"
                  />
                </template>
                <line
                  v-else
                  x1="0"
                  y1="32"
                  x2="300"
                  y2="32"
                  stroke-dasharray="4 4"
                  stroke-width="1.5"
                  class="stroke-lineDim"
                />
              </svg>
              <div class="px-1 flex items-center justify-between text-[0.625rem] text-dull">
                <span>{{ row.targetLabel }}</span>
                <span class="shrink-0">{{ row.bars?.length ?? 0 }} LOGS</span>
              </div>
            </div>
          </div>

          <!-- Readout Column -->
          <div class="px-3 py-2 flex flex-col justify-between hud-value min-w-0">
            <template v-if="row.kind === 'dual-bars'">
              <div class="flex items-baseline justify-between gap-1 text-[0.6875rem] font-semibold">
                <span class="text-dull">{{ row.primaryLabel }}</span>
                <span class="text-side">{{ row.unit }}</span>
              </div>
              <div class="flex flex-col gap-0.5 min-w-0">
                <strong
                  class="text-lg sm:text-xl font-bold leading-tight whitespace-nowrap"
                  :class="row.isDisconnected ? 'text-stale' : statusTextClass(row.primaryStatus ?? row.status)"
                >
                  {{ row.isDisconnected ? "—" : row.value }}
                </strong>
                <div
                  class="text-[0.6875rem] font-semibold truncate"
                  :class="
                    row.isDisconnected || row.secondaryStatus === ReadingStatus.StatusStale
                      ? 'text-stale'
                      : statusTextClass(row.secondaryStatus ?? 'unknown')
                  "
                >
                  {{ row.secondaryLabel }}: {{ row.isDisconnected ? "—" : row.secondaryValue }}
                </div>
              </div>
              <small
                class="text-[0.625rem] font-semibold inline-flex items-center gap-1 min-w-0"
                :class="row.isDisconnected ? 'text-stale' : 'text-dull'"
              >
                <WifiOff v-if="row.isDisconnected" class="w-3 h-3 shrink-0" aria-hidden="true" />
                <span class="truncate">{{ row.lastRecorded }}</span>
              </small>
            </template>

            <template v-else>
              <div
                class="flex items-baseline justify-between gap-1 text-[0.6875rem] font-semibold"
                :class="row.isDisconnected ? 'text-stale' : cardSecondaryTextClass(row.status, 'text-side')"
              >
                <span class="text-[0.625rem] tracking-wider uppercase text-dull">CHECK-IN</span>
                <span>{{ row.unit }}</span>
              </div>
              <strong
                class="font-bold leading-none whitespace-nowrap"
                :class="[
                  row.kind === 'ultrasound' ? 'text-lg sm:text-xl tracking-wide' : 'text-2xl sm:text-[1.75rem]',
                  row.isDisconnected ? 'text-stale' : statusTextClass(row.status),
                ]"
              >
                {{ row.isDisconnected ? "—" : row.value }}
              </strong>
              <small
                class="text-[0.625rem] font-semibold inline-flex items-center gap-1 min-w-0"
                :class="row.isDisconnected ? 'text-stale' : 'text-dull'"
              >
                <WifiOff v-if="row.isDisconnected" class="w-3 h-3 shrink-0" aria-hidden="true" />
                <span class="truncate">{{ row.isDisconnected ? "NO SIGNAL" : row.lastRecorded }}</span>
              </small>
            </template>
          </div>

          <!-- Medical Manual Button -->
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
      </div>
    </section>

    <!-- Empty Filter State -->
    <div
      v-if="!hasVisibleCards"
      class="w-full max-w-5xl hud-card p-8 flex flex-col items-center justify-center gap-3 text-center"
    >
      <p class="text-sm font-medium text-side">
        No other stats metrics match the current filter.
      </p>
      <button
        type="button"
        class="button-side h-9 px-4 text-xs cursor-pointer"
        @click="clearFilters"
      >
        Reset Filters
      </button>
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
