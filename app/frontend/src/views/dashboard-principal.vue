<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { AlertCircle, AlertTriangle, HelpCircle, Home, Search } from "lucide-vue-next"
import { ReadingStatus } from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type TelemetryDisplayStatus,
} from "@/composables/useTelemetry"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const route = useRoute()
const {
  astronauts,
  activeAstronauts,
  metricsByCode,
  getCrewReading,
  getEnvReading,
  formatMetricValue,
  formatBloodPressure,
} = useTelemetry()
const {
  selectedAstronautId,
  selectedAstronautFallbackName,
  setSelectedAstronaut,
  goToHome,
} = useTelemetryNavigation()

const metricFilter = ref("")

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

function statusTextClass(status: TelemetryDisplayStatus | "fallback"): string {
  switch (status) {
    case ReadingStatus.StatusNormal:
    case "fallback":
      return "text-ok"
    case ReadingStatus.StatusAbnormal:
      return "text-warn"
    case ReadingStatus.StatusCritical:
      return "text-crit"
    case ReadingStatus.StatusStale:
      return "text-stale"
    default:
      return "text-side"
  }
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
  value: string
  unit: string
  subline: string
  status: TelemetryDisplayStatus | "fallback"
  wavePath: string
  waveClass: string
  fillPath?: string
}

const vitalRows = computed<VitalRow[]>(() => {
  const id = resolvedAstronautId.value

  const hr = formatNumericPart(id, "heart_rate", "72", "bpm")
  const hrvSub = formatCrewSubLine(id, "hrv_rmssd", "HRV", "25 ms")

  const spo2 = formatNumericPart(id, "spo2", "98%", "SpO2")
  const cabinO2Sub = formatEnvLine("cabin_o2", "O2:", "20.9%")

  const bpFormatted = id ? formatBloodPressure(id) : { display: "—", status: "unknown" as const }
  const bpLive = bpFormatted.status !== "unknown" && bpFormatted.display !== "—"
  const bpValue = bpLive ? bpFormatted.display.replace(/\s*mmHg$/i, "") : "118/77"
  const bpStatus: TelemetryDisplayStatus | "fallback" = bpLive ? bpFormatted.status : "fallback"
  const cabinPressureSub = formatEnvLine("cabin_pressure", "CABIN:", "101.3 kPa")

  const radRate = formatNumericPart(id, "radiation_rate", "32", "µSv/h")
  const radCumSub = formatCrewSubLine(id, "radiation_dose_cum", "TOTAL", "1.2 mSv")

  const coreTemp = formatNumericPart(id, "core_temp", "36.6", "°C")
  const cabinTempSub = formatEnvLine("cabin_temp", "CABIN", "22.0°C")

  const all: VitalRow[] = [
    {
      id: "cardiac",
      title: "CARDIAC RHYTHM",
      value: hr.value,
      unit: hr.unit,
      subline: hrvSub,
      status: hr.status,
      wavePath:
        "M0 38h18l7-2 5 3 6-4 6 2 8-1 4-3 5 13 7-46 7 37 8 1 4-7 5 7 9 0 5-2 5 1 5-4 5 3 5-3 5 4 8-1 5-2 5 3 6-3 5 4 7-2 6 2 5-4 5 3 7-1 8 2 5-4 5 2 7-3 8 4 8-2 5 2 5-5 5 4 7-2 7 2 6-2 7 2 8-2 8 2 6-3 6 3 6-2 7 2 5-3 8 2 6-2 10 2",
      waveClass: "stroke-ok",
    },
    {
      id: "oxygenation",
      title: "OXYGENATION",
      value: spo2.value,
      unit: spo2.unit,
      subline: cabinO2Sub,
      status: spo2.status,
      wavePath:
        "M0 31c12 4 16 20 24 18 8-2 8-48 18-43 10 5 10 38 22 37 12-1 10-20 19-20 9 0 13 27 24 25 11-2 11-40 21-38 10 2 11 35 22 34 11-1 10-24 20-24 10 0 13 29 24 28 11-1 10-40 20-38 10 2 11 36 22 34 11-2 10-25 20-24 10 1 13 27 24 26 11-1 11-40 20-38 9 2 10 36 21 34 11-2 11-24 20-23 9 1 13 23 24 22",
      waveClass: "stroke-ok",
    },
    {
      id: "pressure",
      title: "PRESSURE",
      subtitle: "BP",
      value: bpValue,
      unit: "mmHg",
      subline: cabinPressureSub,
      status: bpStatus,
      wavePath:
        "M0 54c8-1 12-39 20-38 8 1 10 17 18 14 8-3 10-8 17-5 7 3 8 5 15 3 7-2 8-26 15-28 7-2 8 44 17 43 9-1 10-22 18-21 8 1 10 12 18 9 8-3 8-35 16-35 8 0 10 46 19 45 9-1 10-26 18-24 8 2 11 8 18 4 7-4 8-27 16-27 8 0 10 41 18 40 8-1 11-14 19-13 8 1 11 6 18 2 7-4 8-26 16-26 8 0 11 37 19 37 8 0 12-12 20-11 8 1 11 7 20 2",
      waveClass: "stroke-main",
    },
    {
      id: "radiation",
      title: "RADIATION",
      subtitle: "SENSOR",
      value: radRate.value,
      unit: radRate.unit,
      subline: radCumSub,
      status: radRate.status,
      wavePath:
        "M0 11L29 38 29 62 55 62 55 13 84 39 84 62 110 62 110 12 139 41 139 62 166 62 166 12 194 39 194 62 220 62 220 12 249 39 249 62 275 62 300 38",
      fillPath:
        "M0 11L29 38 29 62 55 62 55 13 84 39 84 62 110 62 110 12 139 41 139 62 166 62 166 12 194 39 194 62 220 62 220 12 249 39 249 62 275 62 300 38V64H0Z",
      waveClass: "stroke-warn",
    },
    {
      id: "temperature",
      title: "TEMPERATURE",
      subtitle: "BODY",
      value: coreTemp.value,
      unit: coreTemp.unit,
      subline: cabinTempSub,
      status: coreTemp.status,
      wavePath: "M0 56H300",
      waveClass: "stroke-main",
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
          class="hud-card overflow-hidden grid grid-cols-[minmax(0,1fr)_140px_28px] sm:grid-cols-[minmax(0,1fr)_168px_32px] min-h-22"
        >
          <div class="bg-surface border-r border-solid border-lineDim flex flex-col justify-between min-w-0">
            <div class="px-3 pt-2 flex items-center justify-between gap-2 text-[11px] font-bold tracking-wider">
              <h2 class="text-ok">{{ row.title }}</h2>
              <span v-if="row.subtitle" class="text-main">{{ row.subtitle }}</span>
            </div>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" class="w-full h-14 px-1" aria-hidden="true">
              <path v-if="row.fillPath" :d="row.fillPath" class="fill-warn/20 stroke-none" />
              <path :d="row.wavePath" fill="none" stroke-width="2" :class="row.waveClass" />
            </svg>
          </div>

          <div
            class="px-3 py-2 flex flex-col justify-between hud-value min-w-0"
            :class="statusTextClass(row.status)"
          >
            <div class="flex items-baseline justify-between gap-1 text-[11px] font-semibold text-side">
              <span class="truncate">{{ row.subtitle === "BP" ? "BP" : "" }}</span>
              <span>{{ row.unit }}</span>
            </div>
            <strong class="text-2xl sm:text-3xl font-bold leading-none truncate">
              {{ row.value }}
            </strong>
            <small class="text-[11px] font-semibold text-side truncate">
              {{ row.subline }}
            </small>
          </div>

          <div class="flex items-end justify-center pb-2 text-dull">
            <HelpCircle class="w-3.5 h-3.5" aria-hidden="true" />
          </div>
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
