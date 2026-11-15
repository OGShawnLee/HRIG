<script setup lang="ts">
import { computed } from "vue"
import { AlertCircle, AlertTriangle, WifiOff } from "lucide-vue-next"
import { ReadingStatus } from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type FormattedMetric,
  type TelemetryDisplayStatus,
} from "@/composables/useTelemetry"

const { metricsByCode, getEnvReading, formatMetricValue } = useTelemetry()

interface EnvRow extends FormattedMetric {
  key: string
  label: string
}

const rows = computed<EnvRow[]>(() => [
  {
    key: "cabin_o2",
    label: "O₂:",
    ...formatMetricValue(getEnvReading("cabin_o2"), metricsByCode.value["cabin_o2"]),
  },
  {
    key: "cabin_pressure",
    label: "Cabin Pressure:",
    ...formatMetricValue(getEnvReading("cabin_pressure"), metricsByCode.value["cabin_pressure"]),
  },
  {
    key: "cabin_ppco2",
    label: "Cabin ppCO₂:",
    ...formatMetricValue(getEnvReading("cabin_ppco2"), metricsByCode.value["cabin_ppco2"]),
  },
  {
    key: "cabin_temp",
    label: "Temperature:",
    ...formatMetricValue(getEnvReading("cabin_temp"), metricsByCode.value["cabin_temp"]),
  },
])

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
  <section
    class="hud-card w-full h-full p-5 sm:p-6 flex flex-col justify-between gap-4"
    aria-label="Cabin Environment"
  >
    <h2 class="text-base font-medium text-main leading-tight">
      Environment
    </h2>

    <dl class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-x-6 gap-y-3 text-sm sm:text-base">
      <div
        v-for="row in rows"
        :key="row.key"
        class="flex items-baseline justify-between gap-3 border-b border-solid border-lineDim/40 pb-1.5 min-w-0"
      >
        <dt class="font-semibold text-main whitespace-nowrap shrink-0">
          {{ row.label }}
        </dt>
        <dd
          class="hud-value inline-flex items-center justify-end gap-1.5 min-w-0"
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
      </div>
    </dl>
  </section>
</template>
