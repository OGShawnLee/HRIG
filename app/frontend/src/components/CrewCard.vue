<script setup lang="ts">
import { computed } from "vue"
import { AlertCircle, AlertTriangle, WifiOff } from "lucide-vue-next"
import { ReadingStatus, type Astronaut } from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type FormattedMetric,
  type TelemetryDisplayStatus,
} from "@/composables/useTelemetry"

const props = defineProps<{
  astronaut: Astronaut
}>()

const { metricsByCode, getCrewReading, formatMetricValue, formatBloodPressure } = useTelemetry()

const nameLines = computed(() => {
  const parts = props.astronaut.fullName.trim().split(/\s+/).filter(Boolean)
  if (parts.length <= 1) {
    return { first: props.astronaut.fullName, rest: "" }
  }
  return {
    first: parts[0],
    rest: parts.slice(1).join(" "),
  }
})

const initials = computed(() => {
  const parts = props.astronaut.fullName.trim().split(/\s+/).filter(Boolean)
  return parts
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("")
})

interface MetricRow extends FormattedMetric {
  key: string
  label: string
}

const rows = computed<MetricRow[]>(() => {
  const id = props.astronaut.id
  const spo2 = formatMetricValue(getCrewReading(id, "spo2"), metricsByCode.value["spo2"])
  const bp = formatBloodPressure(id)
  const radRate = formatMetricValue(
    getCrewReading(id, "radiation_rate"),
    metricsByCode.value["radiation_rate"],
  )
  const radCum = formatMetricValue(
    getCrewReading(id, "radiation_dose_cum"),
    metricsByCode.value["radiation_dose_cum"],
  )
  const hr = formatMetricValue(getCrewReading(id, "heart_rate"), metricsByCode.value["heart_rate"])
  const temp = formatMetricValue(getCrewReading(id, "core_temp"), metricsByCode.value["core_temp"])

  return [
    { key: "spo2", label: "Oxygenation:", ...spo2 },
    { key: "bp", label: "Blood Pressure:", ...bp },
    { key: "radiation_rate", label: "Rad. Exposure:", ...radRate },
    { key: "radiation_dose_cum", label: "Acc. Radiation:", ...radCum },
    { key: "heart_rate", label: "Heart Beat:", ...hr },
    { key: "core_temp", label: "Temperature:", ...temp },
  ]
})

function statusTextClass(status: TelemetryDisplayStatus): string {
  switch (status) {
    case ReadingStatus.StatusNormal:
      return "text-ok"
    case ReadingStatus.StatusAbnormal:
      return "text-warn"
    case ReadingStatus.StatusCritical:
      return "text-crit"
    case ReadingStatus.StatusStale:
      return "text-stale"
    default:
      return "text-dull"
  }
}

function statusLabel(status: TelemetryDisplayStatus): string {
  switch (status) {
    case ReadingStatus.StatusNormal:
      return "Nominal"
    case ReadingStatus.StatusAbnormal:
      return "Abnormal"
    case ReadingStatus.StatusCritical:
      return "Critical"
    case ReadingStatus.StatusStale:
      return "Stale"
    default:
      return "No telemetry"
  }
}
</script>

<template>
  <article
    class="hud-card p-6 flex flex-col gap-5"
    :aria-label="`Crew member ${astronaut.fullName}`"
  >
    <div class="flex items-center gap-5">
      <!-- Offline-first NASA-inspired crew portrait badge -->
      <div
        class="w-22 h-24 shrink-0 rounded-lg overflow-hidden border border-solid border-lineDim bg-[#0B1726] relative flex items-center justify-center select-none"
        aria-hidden="true"
      >
        <svg viewBox="0 0 88 96" class="w-full h-full" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect width="88" height="96" fill="#0B1726" />
          <!-- Subtle mission backdrop stripes & agency seal -->
          <rect x="0" y="14" width="18" height="4" fill="#991B1B" opacity="0.65" />
          <rect x="0" y="22" width="18" height="4" fill="#E5E7EB" opacity="0.45" />
          <rect x="0" y="30" width="18" height="4" fill="#991B1B" opacity="0.65" />
          <rect x="0" y="38" width="18" height="4" fill="#E5E7EB" opacity="0.45" />
          <circle cx="72" cy="30" r="14" stroke="#60A5FA" stroke-opacity="0.45" stroke-width="1.5" fill="#1E3A8A" fill-opacity="0.35" />
          <path d="M64 30 L80 30 M72 22 L72 38" stroke="#93C5FD" stroke-opacity="0.45" stroke-width="1" />
          <!-- Crew silhouette in blue mission polo -->
          <path d="M16 96 C16 74 28 66 44 66 C60 66 72 74 72 96 Z" fill="#1D4ED8" />
          <path d="M35 66 L44 76 L53 66" stroke="#93C5FD" stroke-width="1.5" fill="#1E40AF" />
          <circle cx="44" cy="40" r="16" fill="#1E293B" stroke="#94A3B8" stroke-width="1.5" />
        </svg>
        <span
          class="absolute top-[29px] text-sm font-bold tracking-wide text-main"
        >
          {{ initials }}
        </span>
        <span
          v-if="astronaut.nationalityCode"
          class="absolute bottom-1.5 right-1.5 text-[10px] font-semibold tracking-wider px-1.5 py-0.5 rounded bg-ground/80 text-side border border-solid border-lineDim"
        >
          {{ astronaut.nationalityCode }}
        </span>
      </div>

      <div class="flex-1 min-w-0 flex flex-col items-center justify-center text-center">
        <h2 class="text-xl font-medium leading-snug text-main break-words">
          <span class="block">{{ nameLines.first }}</span>
          <span v-if="nameLines.rest" class="block">{{ nameLines.rest }}</span>
        </h2>
      </div>
    </div>

    <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2.5 items-baseline text-sm sm:text-base">
      <template v-for="row in rows" :key="row.key">
        <dt class="font-semibold text-main whitespace-nowrap">
          {{ row.label }}
        </dt>
        <dd
          class="hud-value inline-flex items-center gap-1.5 min-w-0"
          :class="statusTextClass(row.status)"
          :title="`${row.label.replace(':', '')} — ${statusLabel(row.status)}`"
        >
          <span class="truncate">{{ row.display }}</span>
          <AlertTriangle
            v-if="row.status === ReadingStatus.StatusAbnormal"
            class="w-4 h-4 shrink-0"
            aria-hidden="true"
          />
          <AlertCircle
            v-else-if="row.status === ReadingStatus.StatusCritical"
            class="w-4 h-4 shrink-0"
            aria-hidden="true"
          />
          <WifiOff
            v-else-if="row.status === ReadingStatus.StatusStale"
            class="w-4 h-4 shrink-0"
            aria-hidden="true"
          />
          <span v-if="row.status !== ReadingStatus.StatusNormal && row.status !== 'unknown'" class="sr-only">
            ({{ statusLabel(row.status) }})
          </span>
        </dd>
      </template>
    </dl>
  </article>
</template>
