<script setup lang="ts">
import { computed } from "vue"
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  ClipboardList,
  Clock,
  Info,
  Minus,
  PhoneCall,
  PinOff,
  ShieldAlert,
  Stethoscope,
  UserCheck,
  WifiOff,
  Wrench,
} from "lucide-vue-next"
import {
  EscalationLevel,
  ReadingStatus,
  type Astronaut,
  type MedicalProcedure,
  type Metric,
  type MetricThreshold,
  type SpaceHazard,
} from "@bindings/hrig/shared/model/models.js"
import {
  useTelemetry,
  type FormattedMetric,
  type TelemetryDisplayStatus,
} from "@/composables/useTelemetry"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

export type FeaturedEntry =
  | { kind: "crew"; key: string; astronaut: Astronaut }
  | { kind: "procedure"; key: string; procedure: MedicalProcedure }
  | { kind: "hazard"; key: string; hazard: SpaceHazard }
  | { kind: "metric"; key: string; metric: Metric }

const props = defineProps<{
  entry: FeaturedEntry | null
  isPinned?: boolean
  loading?: boolean
}>()

const emit = defineEmits<{
  filterProcedures: [query: string]
  filterSearch: [query: string]
  unpin: []
}>()

const {
  activeAstronauts,
  alerts,
  metricsByCode,
  getCrewReading,
  getEnvReading,
  resolveReadingStatus,
  formatMetricValue,
  formatBloodPressure,
} = useTelemetry()

const { openPrimaryVitals } = useTelemetryNavigation()

function humanizeCode(raw: string | null | undefined): string {
  if (!raw) {
    return ""
  }
  return raw
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ")
}

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

function statusBadgeClass(status: TelemetryDisplayStatus): string {
  switch (status) {
    case ReadingStatus.StatusNormal:
      return "border-ok/60 text-ok"
    case ReadingStatus.StatusAbnormal:
      return "border-warn/60 text-warn"
    case ReadingStatus.StatusCritical:
      return "border-crit/60 text-crit"
    case ReadingStatus.StatusStale:
      return "border-lineDim text-stale"
    default:
      return "border-lineDim text-dull"
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

function escalationMeta(level: EscalationLevel) {
  switch (level) {
    case EscalationLevel.EscalationSelfCare:
      return {
        label: "Self-Care",
        classes: "border-ok/60 text-ok bg-surface",
        icon: UserCheck,
      }
    case EscalationLevel.EscalationCMO:
      return {
        label: "Crew Medical Officer (CMO)",
        classes: "border-warn/60 text-warn bg-surface",
        icon: Stethoscope,
      }
    case EscalationLevel.EscalationFlightSurgeon:
      return {
        label: "Flight Surgeon",
        classes: "border-crit/60 text-crit bg-surface",
        icon: PhoneCall,
      }
    default:
      return {
        label: humanizeCode(level) || "Guidance",
        classes: "border-lineDim text-side bg-surface",
        icon: Info,
      }
  }
}

interface CrewVitalRow extends FormattedMetric {
  key: string
  label: string
}

const crewVitalRows = computed<CrewVitalRow[]>(() => {
  if (props.entry?.kind !== "crew") {
    return []
  }
  const id = props.entry.astronaut.id
  const spo2 = formatMetricValue(getCrewReading(id, "spo2"), metricsByCode.value["spo2"])
  const bp = formatBloodPressure(id)
  const hr = formatMetricValue(getCrewReading(id, "heart_rate"), metricsByCode.value["heart_rate"])
  const temp = formatMetricValue(getCrewReading(id, "core_temp"), metricsByCode.value["core_temp"])
  const radRate = formatMetricValue(
    getCrewReading(id, "radiation_rate"),
    metricsByCode.value["radiation_rate"],
  )
  const radCum = formatMetricValue(
    getCrewReading(id, "radiation_dose_cum"),
    metricsByCode.value["radiation_dose_cum"],
  )

  return [
    { key: "spo2", label: "SpO₂", ...spo2 },
    { key: "bp", label: "Blood Pressure", ...bp },
    { key: "heart_rate", label: "Heart Rate", ...hr },
    { key: "core_temp", label: "Core Temp", ...temp },
    { key: "radiation_rate", label: "Radiation Rate", ...radRate },
    { key: "radiation_dose_cum", label: "Acc. Radiation", ...radCum },
  ]
})

const crewActiveAlerts = computed(() => {
  if (props.entry?.kind !== "crew") {
    return []
  }
  const id = props.entry.astronaut.id
  return alerts.value.filter((a) => a.astronautId === id && !a.acknowledged)
})

const effectiveMetric = computed<Metric | null>(() => {
  if (props.entry?.kind !== "metric") {
    return null
  }
  const searchMetric = props.entry.metric
  const catalogMetric = metricsByCode.value[searchMetric.code]
  if (searchMetric.threshold) {
    return searchMetric
  }
  return catalogMetric ?? searchMetric
})

const metricUnit = computed(() => {
  const m = effectiveMetric.value
  if (!m) {
    return ""
  }
  return m.displayUnitSymbol ?? m.baseUnitSymbol ?? m.baseUnitCode ?? ""
})

function formatThresholdRange(
  minVal: number | null | undefined,
  maxVal: number | null | undefined,
  unit: string,
  kind: "warning" | "critical",
): string {
  const suffix = unit ? ` ${unit}` : ""
  const hasMin = minVal !== null && minVal !== undefined && Number.isFinite(minVal)
  const hasMax = maxVal !== null && maxVal !== undefined && Number.isFinite(maxVal)

  if (kind === "warning") {
    if (hasMin && hasMax) {
      return `${minVal} – ${maxVal}${suffix}`
    }
    if (hasMin) {
      return `≥ ${minVal}${suffix}`
    }
    if (hasMax) {
      return `≤ ${maxVal}${suffix}`
    }
    return "—"
  }

  if (hasMin && hasMax) {
    return `< ${minVal} or > ${maxVal}${suffix}`
  }
  if (hasMin) {
    return `< ${minVal}${suffix}`
  }
  if (hasMax) {
    return `> ${maxVal}${suffix}`
  }
  return "—"
}

function hasAnyThreshold(t: MetricThreshold | null | undefined): boolean {
  if (!t) {
    return false
  }
  return (
    t.warningMin !== null && t.warningMin !== undefined ||
    t.warningMax !== null && t.warningMax !== undefined ||
    t.criticalMin !== null && t.criticalMin !== undefined ||
    t.criticalMax !== null && t.criticalMax !== undefined
  )
}

const metricLiveReadings = computed(() => {
  const m = effectiveMetric.value
  if (!m) {
    return []
  }

  if (m.metricType === 2) {
    const reading = getEnvReading(m.code)
    const formatted = formatMetricValue(reading, m)
    return [
      {
        key: "env",
        label: "Cabin / Suit Sensor",
        display: formatted.display,
        status: formatted.status,
        met: reading?.met ?? "",
        deviceCode: reading?.deviceCode ?? "",
      },
    ]
  }

  return activeAstronauts.value.map((astro) => {
    const reading = getCrewReading(astro.id, m.code)
    const formatted = formatMetricValue(reading, m)
    return {
      key: astro.id,
      label: astro.fullName,
      display: formatted.display,
      status: resolveReadingStatus(reading),
      met: reading?.met ?? "",
      deviceCode: reading?.deviceCode ?? "",
    }
  })
})
</script>

<template>
  <aside
    aria-label="Featured entry details"
    class="flex flex-col h-full min-h-0 bg-ground"
  >
    <!-- Pane Subheader -->
    <div class="px-4 py-2.5 border-b border-solid border-lineDim flex items-center justify-between gap-2 shrink-0">
      <span class="text-[11px] font-semibold uppercase tracking-[0.14em] text-dull">
        Featured Entry
      </span>

      <div class="flex items-center gap-2">
        <span
          v-if="isPinned && entry?.kind === 'procedure'"
          class="inline-flex items-center gap-1 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider border border-solid border-line bg-main text-ground"
        >
          <ClipboardList class="w-3 h-3" aria-hidden="true" />
          Active Checklist
        </span>
        <span
          v-else-if="isPinned"
          class="inline-flex items-center gap-1 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider border border-solid border-lineDim bg-surface text-main"
        >
          Pinned
        </span>
        <button
          v-if="isPinned"
          type="button"
          class="inline-flex items-center gap-1 text-[11px] text-dull hover:text-main transition-colors cursor-pointer"
          title="Return preview to follow highlighted item"
          @click="emit('unpin')"
        >
          <PinOff class="w-3 h-3" aria-hidden="true" />
          <span>Unpin</span>
        </button>
      </div>
    </div>

    <!-- Empty / Loading State -->
    <div
      v-if="!entry"
      class="flex-1 p-6 flex flex-col items-center justify-center text-center gap-2 text-dull"
    >
      <Activity class="w-7 h-7 text-dull" aria-hidden="true" />
      <p class="text-sm text-side font-medium">
        {{ loading ? "Searching HRIG catalog…" : "No matching entry selected" }}
      </p>
      <p class="text-xs text-dull max-w-xs">
        Highlight or select a crew member, medical checklist, spaceflight hazard, or metric to inspect clinical details.
      </p>
    </div>

    <!-- CREW DETAIL -->
    <div
      v-else-if="entry.kind === 'crew'"
      class="flex-1 overflow-y-auto p-5 flex flex-col gap-5"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2 mb-1.5">
            <span class="px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider border border-solid border-lineDim bg-surface text-side">
              Crew
            </span>
            <span
              class="px-2 py-0.5 text-[11px] font-medium border border-solid"
              :class="entry.astronaut.isActive ? 'border-ok/60 text-ok' : 'border-lineDim text-dull'"
            >
              {{ entry.astronaut.isActive ? "Active" : "Inactive" }}
            </span>
          </div>
          <h2 class="text-xl font-semibold text-main leading-tight break-words">
            {{ entry.astronaut.fullName }}
          </h2>
          <p class="text-xs text-dull mt-1">
            {{ entry.astronaut.nationalityLabel || entry.astronaut.nationalityCode }}
            <span v-if="entry.astronaut.nationalityLabel && entry.astronaut.nationalityCode">
              ({{ entry.astronaut.nationalityCode }})
            </span>
            <span v-if="entry.astronaut.age > 0"> · Age {{ entry.astronaut.age }}</span>
          </p>
        </div>
      </div>

      <!-- Active Alert Callout -->
      <div
        v-if="crewActiveAlerts.length > 0"
        role="status"
        class="border border-solid border-crit bg-surface px-3.5 py-2.5 flex items-start gap-2.5 text-xs text-crit"
      >
        <AlertCircle class="w-4 h-4 shrink-0 mt-0.5" aria-hidden="true" />
        <div class="min-w-0">
          <p class="font-semibold">
            {{ crewActiveAlerts.length }} Active Unacknowledged {{ crewActiveAlerts.length === 1 ? "Alert" : "Alerts" }}
          </p>
          <p class="text-side mt-0.5 break-words">
            {{ crewActiveAlerts[0].message }}
          </p>
        </div>
      </div>

      <!-- Live Vitals Summary -->
      <div class="border border-solid border-lineDim bg-surface p-4 flex flex-col gap-3">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-side">
          Live Vitals Summary
        </h3>
        <dl class="grid grid-cols-[1fr_auto] gap-x-4 gap-y-2.5 items-center text-sm">
          <template v-for="row in crewVitalRows" :key="row.key">
            <dt class="text-side font-medium">{{ row.label }}</dt>
            <dd
              class="hud-value inline-flex items-center justify-end gap-1.5"
              :class="statusTextClass(row.status)"
            >
              <span>{{ row.display }}</span>
              <CheckCircle2
                v-if="row.status === ReadingStatus.StatusNormal"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <AlertTriangle
                v-else-if="row.status === ReadingStatus.StatusAbnormal"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <AlertCircle
                v-else-if="row.status === ReadingStatus.StatusCritical"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <WifiOff
                v-else-if="row.status === ReadingStatus.StatusStale"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <Minus
                v-else
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <span class="sr-only">({{ statusLabel(row.status) }})</span>
            </dd>
          </template>
        </dl>
      </div>

      <div class="mt-auto pt-2 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="button-main rounded-none w-full justify-between px-4 text-sm cursor-pointer"
          @click="openPrimaryVitals(entry.astronaut)"
        >
          <span>Open Primary Vitals</span>
          <ArrowRight class="w-4 h-4 shrink-0" aria-hidden="true" />
        </button>
      </div>
    </div>

    <!-- MEDICAL PROCEDURE DETAIL (Option A: full checklist directly inside Featured Entry) -->
    <div
      v-else-if="entry.kind === 'procedure'"
      class="flex-1 overflow-y-auto p-5 flex flex-col gap-4"
    >
      <div class="flex flex-col gap-2">
        <div class="flex flex-wrap items-center gap-1.5">
          <span class="px-2 py-0.5 text-[11px] font-medium border border-solid border-lineDim bg-surface text-side">
            {{ humanizeCode(entry.procedure.category) }}
          </span>
          <span class="px-2 py-0.5 text-[11px] font-medium border border-solid border-lineDim bg-surface text-side">
            {{ humanizeCode(entry.procedure.bodySystem) }}
          </span>
          <span
            v-if="entry.procedure.environment"
            class="px-2 py-0.5 text-[11px] font-semibold uppercase border border-solid border-lineDim text-main"
          >
            {{ entry.procedure.environment }}
          </span>
        </div>

        <h2 class="text-lg font-semibold text-main leading-snug break-words">
          {{ entry.procedure.title }}
        </h2>

        <div class="flex flex-wrap items-center gap-2 pt-0.5">
          <span
            class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-semibold border border-solid"
            :class="escalationMeta(entry.procedure.escalation).classes"
          >
            <component
              :is="escalationMeta(entry.procedure.escalation).icon"
              class="w-3.5 h-3.5 shrink-0"
              aria-hidden="true"
            />
            <span>{{ escalationMeta(entry.procedure.escalation).label }}</span>
          </span>

          <span class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium text-side border border-solid border-lineDim bg-surface tabular-nums">
            <Clock class="w-3.5 h-3.5 text-dull shrink-0" aria-hidden="true" />
            <span>~{{ entry.procedure.estimatedMinutes }} min</span>
          </span>
        </div>
      </div>

      <!-- When to Use -->
      <section class="border border-solid border-lineDim bg-surface p-3.5 flex flex-col gap-1.5">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          When to Use
        </h3>
        <p class="text-sm text-main leading-relaxed">
          {{ entry.procedure.whenToUse }}
        </p>
        <p v-if="entry.procedure.summary" class="text-xs text-dull leading-relaxed">
          {{ entry.procedure.summary }}
        </p>
      </section>

      <!-- Equipment Needed -->
      <section
        v-if="entry.procedure.equipment.length > 0"
        class="flex flex-col gap-2"
      >
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side flex items-center gap-1.5">
          <Wrench class="w-3.5 h-3.5 text-dull" aria-hidden="true" />
          <span>Equipment Needed</span>
        </h3>
        <ul class="flex flex-wrap gap-1.5">
          <li
            v-for="item in entry.procedure.equipment"
            :key="item"
            class="px-2.5 py-1 text-xs text-main border border-solid border-lineDim bg-surface"
          >
            {{ item }}
          </li>
        </ul>
      </section>

      <!-- Numbered Steps -->
      <section class="flex flex-col gap-2">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          Procedure Checklist ({{ entry.procedure.steps.length }} Steps)
        </h3>
        <ol class="flex flex-col gap-2">
          <li
            v-for="(step, idx) in entry.procedure.steps"
            :key="idx"
            class="flex items-start gap-3 border border-solid border-lineDim bg-surface/60 px-3 py-2.5 text-sm text-main leading-relaxed"
          >
            <span class="w-6 h-6 shrink-0 inline-flex items-center justify-center text-xs font-bold tabular-nums border border-solid border-line text-main bg-ground mt-0.5">
              {{ idx + 1 }}
            </span>
            <span class="min-w-0 flex-1">{{ step }}</span>
          </li>
        </ol>
      </section>

      <!-- Warnings / Escalation Triggers -->
      <section
        v-if="entry.procedure.warnings.length > 0"
        class="border border-solid border-warn bg-surface p-3.5 flex flex-col gap-2"
      >
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-warn flex items-center gap-1.5">
          <AlertTriangle class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
          <span>Warnings & Escalation Triggers</span>
        </h3>
        <ul class="flex flex-col gap-1.5 text-xs text-main leading-relaxed list-disc pl-4">
          <li v-for="(warning, idx) in entry.procedure.warnings" :key="idx">
            {{ warning }}
          </li>
        </ul>
      </section>

      <!-- Linked Metrics & Hazards -->
      <section
        v-if="entry.procedure.relatedMetricCodes.length > 0 || entry.procedure.hazardCodes.length > 0"
        class="flex flex-col gap-2.5 pt-1"
      >
        <div v-if="entry.procedure.relatedMetricCodes.length > 0" class="flex flex-col gap-1.5">
          <span class="text-[11px] font-semibold uppercase tracking-wider text-dull">
            Related Metrics (click to filter)
          </span>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="code in entry.procedure.relatedMetricCodes"
              :key="code"
              type="button"
              class="px-2 py-1 text-xs font-mono text-main border border-solid border-lineDim bg-surface hover:border-line transition-colors cursor-pointer"
              @click="emit('filterProcedures', code)"
            >
              {{ code }}
            </button>
          </div>
        </div>

        <div v-if="entry.procedure.hazardCodes.length > 0" class="flex flex-col gap-1.5">
          <span class="text-[11px] font-semibold uppercase tracking-wider text-dull">
            Linked NASA Hazards
          </span>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="hCode in entry.procedure.hazardCodes"
              :key="hCode"
              type="button"
              class="px-2 py-1 text-xs text-side border border-solid border-lineDim bg-ground hover:border-line hover:text-main transition-colors cursor-pointer"
              @click="emit('filterSearch', humanizeCode(hCode))"
            >
              {{ humanizeCode(hCode) }}
            </button>
          </div>
        </div>
      </section>

      <p class="text-[11px] text-dull border-t border-solid border-lineDim pt-3 mt-auto">
        Source: {{ entry.procedure.source }} · Code: <span class="font-mono">{{ entry.procedure.code }}</span>
      </p>
    </div>

    <!-- SPACE HAZARD DETAIL -->
    <div
      v-else-if="entry.kind === 'hazard'"
      class="flex-1 overflow-y-auto p-5 flex flex-col gap-4"
    >
      <div>
        <div class="flex flex-wrap items-center gap-2 mb-1.5">
          <span class="inline-flex items-center gap-1 px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider border border-solid border-warn/60 bg-surface text-warn">
            <ShieldAlert class="w-3 h-3" aria-hidden="true" />
            NASA HRP Hazard
          </span>
          <span class="text-[11px] font-mono text-dull">{{ entry.hazard.code }}</span>
        </div>
        <h2 class="text-xl font-semibold text-main leading-tight break-words">
          {{ entry.hazard.name }}
        </h2>
      </div>

      <section class="border border-solid border-lineDim bg-surface p-4 flex flex-col gap-1.5">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          What It Is
        </h3>
        <p class="text-sm text-main leading-relaxed">
          {{ entry.hazard.summary }}
        </p>
      </section>

      <section class="border border-solid border-lineDim bg-surface p-4 flex flex-col gap-1.5">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          Why It Matters (Clinical Impact)
        </h3>
        <p class="text-sm text-main leading-relaxed">
          {{ entry.hazard.clinicalImpact }}
        </p>
      </section>

      <section
        v-if="entry.hazard.relatedMetricCodes.length > 0"
        class="flex flex-col gap-2"
      >
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-dull">
          Monitored Telemetry Metrics
        </h3>
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="code in entry.hazard.relatedMetricCodes"
            :key="code"
            type="button"
            class="px-2.5 py-1 text-xs font-mono text-main border border-solid border-lineDim bg-surface hover:border-line transition-colors cursor-pointer"
            @click="emit('filterProcedures', code)"
          >
            {{ code }}
          </button>
        </div>
      </section>

      <div class="pt-1">
        <button
          type="button"
          class="button-main rounded-none w-full justify-between px-4 text-sm cursor-pointer"
          @click="emit('filterProcedures', entry.hazard.relatedMetricCodes[0] ?? entry.hazard.name)"
        >
          <span>View Matching Procedures</span>
          <ArrowRight class="w-4 h-4 shrink-0" aria-hidden="true" />
        </button>
      </div>

      <p class="text-[11px] text-dull border-t border-solid border-lineDim pt-3 mt-auto">
        Source: {{ entry.hazard.source }}
      </p>
    </div>

    <!-- METRIC DETAIL -->
    <div
      v-else-if="entry.kind === 'metric' && effectiveMetric"
      class="flex-1 overflow-y-auto p-5 flex flex-col gap-4"
    >
      <div>
        <div class="flex flex-wrap items-center gap-2 mb-1.5">
          <span class="px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider border border-solid border-lineDim bg-surface text-side">
            {{ effectiveMetric.metricType === 2 ? "Environment Metric" : "Crew Metric" }}
          </span>
          <span class="px-2 py-0.5 text-[11px] font-medium uppercase tracking-wider border border-solid border-lineDim text-main">
            Tier: {{ effectiveMetric.tier }}
          </span>
          <span v-if="metricUnit" class="px-2 py-0.5 text-[11px] font-mono border border-solid border-lineDim text-side">
            {{ metricUnit }}
          </span>
        </div>
        <h2 class="text-xl font-semibold text-main leading-tight break-words">
          {{ effectiveMetric.name }}
        </h2>
        <p class="text-xs text-dull mt-1 font-mono">
          {{ effectiveMetric.code }} · {{ humanizeCode(effectiveMetric.category) }}
        </p>
      </div>

      <!-- DB Thresholds -->
      <section class="border border-solid border-lineDim bg-surface p-4 flex flex-col gap-2.5">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          Catalog Thresholds
        </h3>
        <template v-if="hasAnyThreshold(effectiveMetric.threshold)">
          <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            <dt class="text-warn font-medium">Warning Band</dt>
            <dd class="hud-value text-right text-main">
              {{
                formatThresholdRange(
                  effectiveMetric.threshold?.warningMin,
                  effectiveMetric.threshold?.warningMax,
                  metricUnit,
                  "warning",
                )
              }}
            </dd>
            <dt class="text-crit font-medium">Critical Bounds</dt>
            <dd class="hud-value text-right text-main">
              {{
                formatThresholdRange(
                  effectiveMetric.threshold?.criticalMin,
                  effectiveMetric.threshold?.criticalMax,
                  metricUnit,
                  "critical",
                )
              }}
            </dd>
          </dl>
          <p v-if="effectiveMetric.threshold?.source" class="text-[11px] text-dull pt-1 border-t border-solid border-lineDim">
            Threshold source: {{ effectiveMetric.threshold.source }}
          </p>
        </template>
        <p v-else class="text-xs text-dull">
          No numeric warning/critical bounds defined for this metric ({{ effectiveMetric.valueType }}).
        </p>
      </section>

      <!-- Live Telemetry Readings -->
      <section class="border border-solid border-lineDim bg-surface p-4 flex flex-col gap-2.5">
        <h3 class="text-[11px] font-semibold uppercase tracking-wider text-side">
          Live Telemetry
        </h3>
        <div v-if="metricLiveReadings.length === 0" class="text-xs text-dull">
          No active telemetry stream readings available.
        </div>
        <ul v-else class="flex flex-col gap-2">
          <li
            v-for="item in metricLiveReadings"
            :key="item.key"
            class="flex items-center justify-between gap-3 text-sm border-b border-solid border-lineDim/60 last:border-b-0 pb-1.5 last:pb-0"
          >
            <div class="min-w-0">
              <p class="text-main font-medium truncate">{{ item.label }}</p>
              <p v-if="item.met" class="text-[11px] text-dull font-mono">
                MET {{ item.met }}
              </p>
            </div>
            <span
              class="inline-flex items-center gap-1.5 px-2 py-0.5 text-xs border border-solid shrink-0 hud-value"
              :class="statusBadgeClass(item.status)"
            >
              <span>{{ item.display }}</span>
              <CheckCircle2
                v-if="item.status === ReadingStatus.StatusNormal"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <AlertTriangle
                v-else-if="item.status === ReadingStatus.StatusAbnormal"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <AlertCircle
                v-else-if="item.status === ReadingStatus.StatusCritical"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <WifiOff
                v-else-if="item.status === ReadingStatus.StatusStale"
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
              <Minus
                v-else
                class="w-3.5 h-3.5 shrink-0"
                aria-hidden="true"
              />
            </span>
          </li>
        </ul>
      </section>

      <div class="mt-auto pt-2">
        <button
          type="button"
          class="button-main rounded-none w-full justify-between px-4 text-sm cursor-pointer"
          @click="emit('filterProcedures', effectiveMetric.code)"
        >
          <span>View Procedures for {{ effectiveMetric.code }}</span>
          <ArrowRight class="w-4 h-4 shrink-0" aria-hidden="true" />
        </button>
      </div>
    </div>
  </aside>
</template>
