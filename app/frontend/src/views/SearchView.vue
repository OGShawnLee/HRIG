<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  ComboboxContent,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxLabel,
  ComboboxRoot,
  ComboboxViewport,
} from "reka-ui"
import {
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  Clock,
  Info,
  Minus,
  PhoneCall,
  RefreshCw,
  Search as SearchIcon,
  ShieldAlert,
  Stethoscope,
  UserCheck,
  WifiOff,
  X,
} from "lucide-vue-next"
import { Search as SearchRPC } from "@bindings/hrig/app/app.js"
import {
  EscalationLevel,
  ReadingStatus,
  SearchResults,
  type Astronaut,
  type MedicalProcedure,
  type Metric,
  type SpaceHazard,
} from "@bindings/hrig/shared/model/models.js"
import SearchFeaturedEntry, { type FeaturedEntry } from "@/components/SearchFeaturedEntry.vue"
import { useTelemetry, type TelemetryDisplayStatus } from "@/composables/useTelemetry"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

type SearchTab = "all" | "crew" | "procedures" | "hazards_metrics"

const TABS: { id: SearchTab; label: string }[] = [
  { id: "all", label: "All" },
  { id: "crew", label: "Crew" },
  { id: "procedures", label: "Procedures" },
  { id: "hazards_metrics", label: "Hazards & Metrics" },
]

const STATUS_PRIORITY: Record<TelemetryDisplayStatus, number> = {
  unknown: 0,
  [ReadingStatus.$zero]: 0,
  [ReadingStatus.StatusNormal]: 1,
  [ReadingStatus.StatusAbnormal]: 2,
  [ReadingStatus.StatusStale]: 3,
  [ReadingStatus.StatusCritical]: 4,
}

const route = useRoute()
const router = useRouter()
const {
  activeAstronauts,
  alerts,
  metricsByCode,
  getCrewReading,
  getEnvReading,
  resolveReadingStatus,
  formatMetricValue,
} = useTelemetry()
const { openPrimaryVitals } = useTelemetryNavigation()

function normalizeTab(raw: unknown): SearchTab {
  if (raw === "crew" || raw === "procedures" || raw === "hazards_metrics" || raw === "all") {
    return raw
  }
  return "all"
}

const query = ref<string>(typeof route.query.q === "string" ? route.query.q : "")
const activeTab = ref<SearchTab>(normalizeTab(route.query.tab))
const results = ref<SearchResults>(new SearchResults())
const loading = ref<boolean>(false)
const hasLoadedOnce = ref<boolean>(false)
const searchError = ref<string | null>(null)

const selectedKey = ref<string | undefined>(undefined)
const highlightedKey = ref<string | null>(null)
const pinnedKey = ref<string | null>(null)

const omniboxContainerRef = ref<HTMLElement | null>(null)
const featuredPaneRef = ref<HTMLElement | null>(null)

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let requestSequence = 0

function crewKey(a: Astronaut): string {
  return `crew:${a.id}`
}

function procKey(p: MedicalProcedure): string {
  return `proc:${p.code}`
}

function hazardKey(h: SpaceHazard): string {
  return `hazard:${h.code}`
}

function metricKey(m: Metric): string {
  return `metric:${m.code}`
}

type ResultGroupKind = "crew" | "metrics" | "procedures" | "hazards"

const hasActiveQuery = computed<boolean>(() => query.value.trim().length > 0)

const visibleGroupOrder = computed<ResultGroupKind[]>(() =>
  hasActiveQuery.value
    ? ["crew", "metrics", "procedures", "hazards"]
    : ["crew", "procedures", "hazards", "metrics"],
)

const entriesByKey = computed<Map<string, FeaturedEntry>>(() => {
  const map = new Map<string, FeaturedEntry>()
  for (const astronaut of results.value.astronauts) {
    const key = crewKey(astronaut)
    map.set(key, { kind: "crew", key, astronaut })
  }
  if (hasActiveQuery.value) {
    for (const metric of results.value.metrics) {
      const key = metricKey(metric)
      map.set(key, { kind: "metric", key, metric })
    }
  }
  for (const procedure of results.value.procedures) {
    const key = procKey(procedure)
    map.set(key, { kind: "procedure", key, procedure })
  }
  for (const hazard of results.value.hazards) {
    const key = hazardKey(hazard)
    map.set(key, { kind: "hazard", key, hazard })
  }
  if (!hasActiveQuery.value) {
    for (const metric of results.value.metrics) {
      const key = metricKey(metric)
      map.set(key, { kind: "metric", key, metric })
    }
  }
  return map
})

const firstVisibleKey = computed<string | null>(() => {
  const first = entriesByKey.value.keys().next()
  return first.done ? null : first.value
})

const featuredKey = computed<string | null>(() => {
  const map = entriesByKey.value
  if (pinnedKey.value && map.has(pinnedKey.value)) {
    return pinnedKey.value
  }
  if (highlightedKey.value && map.has(highlightedKey.value)) {
    return highlightedKey.value
  }
  return firstVisibleKey.value
})

const featuredEntry = computed<FeaturedEntry | null>(() => {
  if (!featuredKey.value) {
    return null
  }
  return entriesByKey.value.get(featuredKey.value) ?? null
})

const isFeaturedPinned = computed<boolean>(
  () => Boolean(pinnedKey.value && featuredKey.value === pinnedKey.value),
)

const totalResults = computed<number>(
  () =>
    results.value.astronauts.length +
    results.value.procedures.length +
    results.value.hazards.length +
    results.value.metrics.length,
)

function padCount(count: number): string {
  return String(count).padStart(2, "0")
}

function focusOmniboxInput(selectAll = false): void {
  void nextTick(() => {
    const inputEl = omniboxContainerRef.value?.querySelector<HTMLInputElement>("#hrig-omnibox-input")
    if (!inputEl) {
      return
    }
    inputEl.focus()
    if (selectAll && inputEl.value.length > 0) {
      inputEl.select()
    }
  })
}

function reconcileKeys(): void {
  const map = entriesByKey.value
  if (pinnedKey.value && !map.has(pinnedKey.value)) {
    pinnedKey.value = null
  }
  if (!highlightedKey.value || !map.has(highlightedKey.value)) {
    highlightedKey.value = firstVisibleKey.value
  }
}

function syncRouteQuery(q: string, tab: SearchTab): void {
  const nextQuery: Record<string, string> = {}
  if (q.trim() !== "") {
    nextQuery.q = q
  }
  if (tab !== "all") {
    nextQuery.tab = tab
  }
  const currentQ = typeof route.query.q === "string" ? route.query.q : ""
  const currentTab = normalizeTab(route.query.tab)
  if (currentQ === (nextQuery.q ?? "") && currentTab === tab) {
    return
  }
  void router.replace({ name: "search", query: nextQuery })
}

async function executeSearch(q: string, tab: SearchTab): Promise<void> {
  const seq = ++requestSequence
  loading.value = true
  searchError.value = null
  syncRouteQuery(q, tab)

  try {
    const res = await SearchRPC(q, tab)
    if (seq !== requestSequence) {
      return
    }
    results.value = res ?? new SearchResults()
    hasLoadedOnce.value = true
    reconcileKeys()
  } catch (err) {
    if (seq !== requestSequence) {
      return
    }
    searchError.value = err instanceof Error ? err.message : String(err)
    hasLoadedOnce.value = true
  } finally {
    if (seq === requestSequence) {
      loading.value = false
    }
  }
}

function scheduleSearch(q: string, tab: SearchTab): void {
  if (debounceTimer !== null) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    debounceTimer = null
    void executeSearch(q, tab)
  }, 120)
}

watch([query, activeTab], ([nextQ, nextTab]) => {
  scheduleSearch(nextQ, nextTab)
})

watch(
  () => [route.query.q, route.query.tab] as const,
  ([nextRouteQ, nextRouteTab]) => {
    const normalizedQ = typeof nextRouteQ === "string" ? nextRouteQ : ""
    const normalizedTab = normalizeTab(nextRouteTab)
    if (normalizedQ !== query.value) {
      query.value = normalizedQ
    }
    if (normalizedTab !== activeTab.value) {
      activeTab.value = normalizedTab
    }
  },
)

function selectTab(tab: SearchTab): void {
  if (activeTab.value !== tab) {
    pinnedKey.value = null
    activeTab.value = tab
  }
  focusOmniboxInput(false)
}

function clearQuery(): void {
  query.value = ""
  pinnedKey.value = null
  focusOmniboxInput(false)
}

function resetFilters(): void {
  query.value = ""
  activeTab.value = "all"
  pinnedKey.value = null
  focusOmniboxInput(false)
}

function closeSearch(): void {
  if (window.history.length > 1 && window.history.state?.back) {
    router.back()
  } else {
    void router.push({ name: "home" })
  }
}

function onHighlight(payload: { ref: HTMLElement; value: unknown } | undefined): void {
  if (payload && typeof payload.value === "string" && payload.value !== "") {
    highlightedKey.value = payload.value
  }
}

function activateEntryByKey(key: string): void {
  const entry = entriesByKey.value.get(key)
  if (!entry) {
    return
  }
  if (entry.kind === "crew") {
    openPrimaryVitals(entry.astronaut)
    return
  }
  pinnedKey.value = key
  highlightedKey.value = key
  if (entry.kind === "procedure") {
    void nextTick(() => {
      featuredPaneRef.value?.scrollIntoView({ behavior: "smooth", block: "nearest" })
    })
  }
}

function onSelectModelValue(val: unknown): void {
  if (typeof val === "string" && val !== "") {
    selectedKey.value = val
    activateEntryByKey(val)
    return
  }
  // Re-selecting an already-selected item toggles modelValue to undefined in Listbox;
  // re-pin the currently highlighted item cleanly so Enter/click remains responsive.
  if (highlightedKey.value) {
    selectedKey.value = highlightedKey.value
    activateEntryByKey(highlightedKey.value)
  }
}

function handleFilterProcedures(nextQuery: string): void {
  pinnedKey.value = null
  activeTab.value = "procedures"
  query.value = nextQuery
  focusOmniboxInput(false)
}

function handleFilterSearch(nextQuery: string): void {
  pinnedKey.value = null
  activeTab.value = "all"
  query.value = nextQuery
  focusOmniboxInput(false)
}

function handleFocusOmniboxEvent(): void {
  focusOmniboxInput(true)
}

function crewStatusSummary(astronaut: Astronaut): {
  status: TelemetryDisplayStatus
  label: string
  snippet: string
  activeAlertCount: number
} {
  const alertCount = alerts.value.filter(
    (a) => a.astronautId === astronaut.id && !a.acknowledged,
  ).length

  const codes = ["spo2", "heart_rate", "core_temp", "bp_systolic", "bp_diastolic", "radiation_rate"]
  let worst: TelemetryDisplayStatus = "unknown"
  for (const code of codes) {
    const st = resolveReadingStatus(getCrewReading(astronaut.id, code))
    if (STATUS_PRIORITY[st] > STATUS_PRIORITY[worst]) {
      worst = st
    }
  }
  if (alertCount > 0 && STATUS_PRIORITY[ReadingStatus.StatusCritical] > STATUS_PRIORITY[worst]) {
    worst = ReadingStatus.StatusCritical
  }

  const hr = formatMetricValue(
    getCrewReading(astronaut.id, "heart_rate"),
    metricsByCode.value["heart_rate"],
  )
  const spo2 = formatMetricValue(
    getCrewReading(astronaut.id, "spo2"),
    metricsByCode.value["spo2"],
  )
  const parts: string[] = []
  if (hr.display !== "—") {
    parts.push(`HR ${hr.display}`)
  }
  if (spo2.display !== "—") {
    parts.push(`SpO₂ ${spo2.display}`)
  }

  return {
    status: worst,
    label: statusLabel(worst),
    snippet: parts.join(" · ") || "Telemetry awaiting frame",
    activeAlertCount: alertCount,
  }
}

interface MetricLiveRow {
  key: string
  label: string
  display: string
  status: TelemetryDisplayStatus
}

function metricLiveRows(metric: Metric): MetricLiveRow[] {
  if (metric.metricType === 2) {
    const formatted = formatMetricValue(getEnvReading(metric.code), metric)
    return [
      {
        key: "env",
        label: "Cabin / Suit",
        display: formatted.display,
        status: formatted.status,
      },
    ]
  }
  return activeAstronauts.value.map((astro) => {
    const formatted = formatMetricValue(getCrewReading(astro.id, metric.code), metric)
    return {
      key: astro.id,
      label: astro.fullName,
      display: formatted.display,
      status: formatted.status,
    }
  })
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
      return "No data"
  }
}

function escalationShort(level: EscalationLevel) {
  switch (level) {
    case EscalationLevel.EscalationSelfCare:
      return { label: "Self-Care", classes: "text-ok border-ok/50", icon: UserCheck }
    case EscalationLevel.EscalationCMO:
      return { label: "CMO", classes: "text-warn border-warn/50", icon: Stethoscope }
    case EscalationLevel.EscalationFlightSurgeon:
      return { label: "Flight Surgeon", classes: "text-crit border-crit/50", icon: PhoneCall }
    default:
      return { label: "Guide", classes: "text-side border-lineDim", icon: Info }
  }
}

onMounted(() => {
  window.addEventListener("hrig:focus-omnibox", handleFocusOmniboxEvent)
  void executeSearch(query.value, activeTab.value)
  focusOmniboxInput(query.value.length > 0)
})

onBeforeUnmount(() => {
  window.removeEventListener("hrig:focus-omnibox", handleFocusOmniboxEvent)
  if (debounceTimer !== null) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
})
</script>

<template>
  <main class="font-sans flex-1 w-full bg-ground text-main px-4 sm:px-6 py-4 sm:py-6 flex flex-col min-h-0">
    <div
      ref="omniboxContainerRef"
      class="max-w-6xl w-full mx-auto rounded-2xl overflow-hidden border border-solid border-line bg-ground flex flex-col flex-1 min-h-[540px]"
    >
      <ComboboxRoot
        :model-value="selectedKey"
        :open="true"
        :ignore-filter="true"
        :reset-search-term-on-blur="false"
        :reset-search-term-on-select="false"
        class="flex flex-col flex-1 min-h-0"
        @update:model-value="onSelectModelValue"
        @highlight="onHighlight"
      >
        <!-- Top Omnibox Search Input Row -->
        <div class="px-4 py-3 border-b border-solid border-lineDim flex items-center gap-3 shrink-0">
          <SearchIcon class="w-4 h-4 text-side shrink-0" aria-hidden="true" />
          <ComboboxInput
            id="hrig-omnibox-input"
            v-model="query"
            type="text"
            aria-label="Search crew, medical procedures, hazards, or metrics"
            placeholder="Search crew, medical procedures, hazards, or metrics..."
            class="font-sans flex-1 min-w-0 bg-ground text-main placeholder-dull text-sm sm:text-base font-medium focus:outline-none"
            @keydown.escape.prevent="closeSearch"
          />

          <RefreshCw
            v-if="loading"
            class="w-4 h-4 text-dull animate-spin shrink-0"
            aria-label="Searching"
          />

          <button
            v-if="query.length > 0"
            type="button"
            aria-label="Clear search query"
            title="Clear query"
            class="font-sans p-1 rounded-lg text-side hover:text-main hover:bg-surface transition-colors cursor-pointer shrink-0"
            @click="clearQuery"
          >
            <X class="w-4 h-4" aria-hidden="true" />
          </button>

          <span class="w-px h-5 bg-lineDim shrink-0" aria-hidden="true" />

          <kbd
            class="font-sans hidden sm:inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium text-dull border border-solid border-lineDim shrink-0"
          >
            Ctrl + K
          </kbd>

          <button
            type="button"
            aria-label="Close search"
            class="font-sans inline-flex items-center gap-1.5 px-2.5 py-1 rounded-xl text-xs font-medium text-side border border-solid border-lineDim hover:border-line hover:text-main transition-colors cursor-pointer shrink-0"
            @click="closeSearch"
          >
            <span>Close</span>
          </button>
        </div>

        <!-- Category Tabs Row -->
        <div
          role="tablist"
          aria-label="Search categories"
          class="px-4 py-2.5 border-b border-solid border-lineDim flex flex-wrap items-center gap-2 shrink-0 bg-ground"
        >
          <button
            v-for="tab in TABS"
            :key="tab.id"
            type="button"
            role="tab"
            :aria-selected="activeTab === tab.id"
            class="font-sans px-3 py-1.5 rounded-xl text-xs border border-solid transition-colors cursor-pointer"
            :class="
              activeTab === tab.id
                ? 'bg-main text-ground border-line font-semibold'
                : 'bg-ground text-side border-lineDim hover:border-line hover:text-main'
            "
            @click="selectTab(tab.id)"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- Error Notice -->
        <div
          v-if="searchError"
          role="alert"
          class="px-4 py-2.5 border-b border-solid border-warn bg-surface flex items-center justify-between gap-3 text-xs text-warn shrink-0"
        >
          <div class="flex items-center gap-2 min-w-0">
            <AlertTriangle class="w-4 h-4 shrink-0" aria-hidden="true" />
            <span class="truncate">Search request failed: {{ searchError }}</span>
          </div>
          <button
            type="button"
            class="font-sans px-2.5 py-1 rounded-lg border border-solid border-lineDim text-main hover:border-line transition-colors cursor-pointer shrink-0"
            @click="executeSearch(query, activeTab)"
          >
            Retry
          </button>
        </div>

        <!-- Main 2-Pane Split Body: Left Featured Entry + Right Grouped Results -->
        <div class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-12">
          <!-- Left Pane: Featured Entry -->
          <div
            ref="featuredPaneRef"
            class="lg:col-span-5 border-b lg:border-b-0 lg:border-r border-solid border-lineDim min-h-[280px] lg:min-h-0 flex flex-col"
          >
            <SearchFeaturedEntry
              :entry="featuredEntry"
              :is-pinned="isFeaturedPinned"
              :loading="loading"
              @filter-procedures="handleFilterProcedures"
              @filter-search="handleFilterSearch"
              @unpin="pinnedKey = null"
            />
          </div>

          <!-- Right Pane: Reka UI ComboboxContent + Viewport -->
          <div class="lg:col-span-7 flex flex-col min-h-[300px] lg:min-h-0 bg-ground">
            <ComboboxContent
              position="inline"
              class="flex-1 min-h-0 flex flex-col"
              @escape-key-down.prevent="closeSearch"
              @pointer-down-outside.prevent
              @focus-outside.prevent
              @interact-outside.prevent
            >
              <ComboboxViewport class="flex-1 overflow-y-auto p-4 flex flex-col gap-5">
                <template v-for="group in visibleGroupOrder" :key="group">
                  <!-- Crew Group -->
                  <ComboboxGroup
                    v-if="group === 'crew' && results.astronauts.length > 0"
                    class="flex flex-col gap-2"
                  >
                    <ComboboxLabel
                      class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-[0.14em] text-dull pb-1 border-b border-solid border-lineDim"
                    >
                      <span>Crew</span>
                      <span class="tabular-nums text-side">{{ padCount(results.astronauts.length) }}</span>
                    </ComboboxLabel>

                    <div class="flex flex-col gap-2">
                      <ComboboxItem
                        v-for="astronaut in results.astronauts"
                        :key="crewKey(astronaut)"
                        :value="crewKey(astronaut)"
                        :text-value="astronaut.fullName"
                        class="w-full text-left px-3.5 py-2.5 rounded-xl border border-solid border-lineDim bg-ground hover:border-line hover:bg-surface data-[highlighted]:border-line data-[highlighted]:bg-surface data-[highlighted]:ring-1 data-[highlighted]:ring-line transition-colors cursor-pointer flex items-center justify-between gap-3 focus:outline-none"
                      >
                        <div class="min-w-0">
                          <div class="flex items-center gap-2">
                            <span class="text-sm font-semibold text-main truncate">
                              {{ astronaut.fullName }}
                            </span>
                            <span class="px-1.5 py-0.5 rounded-md text-[10px] font-sans border border-solid border-lineDim text-side shrink-0">
                              {{ astronaut.nationalityCode }}
                            </span>
                          </div>
                          <p class="text-xs text-dull mt-0.5 truncate">
                            {{ astronaut.nationalityLabel || astronaut.nationalityCode }} · {{ crewStatusSummary(astronaut).snippet }}
                          </p>
                        </div>

                        <div class="flex items-center gap-2 shrink-0">
                          <span
                            v-if="crewStatusSummary(astronaut).activeAlertCount > 0"
                            class="px-1.5 py-0.5 rounded-md text-[10px] font-semibold border border-solid border-crit text-crit tabular-nums"
                          >
                            {{ crewStatusSummary(astronaut).activeAlertCount }} Alert
                          </span>
                          <span
                            class="inline-flex items-center gap-1 text-xs font-medium"
                            :class="statusTextClass(crewStatusSummary(astronaut).status)"
                          >
                            <CheckCircle2
                              v-if="crewStatusSummary(astronaut).status === ReadingStatus.StatusNormal"
                              class="w-3.5 h-3.5"
                              aria-hidden="true"
                            />
                            <AlertTriangle
                              v-else-if="crewStatusSummary(astronaut).status === ReadingStatus.StatusAbnormal"
                              class="w-3.5 h-3.5"
                              aria-hidden="true"
                            />
                            <AlertCircle
                              v-else-if="crewStatusSummary(astronaut).status === ReadingStatus.StatusCritical"
                              class="w-3.5 h-3.5"
                              aria-hidden="true"
                            />
                            <WifiOff
                              v-else-if="crewStatusSummary(astronaut).status === ReadingStatus.StatusStale"
                              class="w-3.5 h-3.5"
                              aria-hidden="true"
                            />
                            <Minus
                              v-else
                              class="w-3.5 h-3.5"
                              aria-hidden="true"
                            />
                            <span>{{ crewStatusSummary(astronaut).label }}</span>
                          </span>
                        </div>
                      </ComboboxItem>
                    </div>
                  </ComboboxGroup>

                  <!-- Metrics Group -->
                  <ComboboxGroup
                    v-else-if="group === 'metrics' && results.metrics.length > 0"
                    class="flex flex-col gap-2"
                  >
                    <ComboboxLabel
                      class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-[0.14em] text-dull pb-1 border-b border-solid border-lineDim"
                    >
                      <span>Metrics</span>
                      <span class="tabular-nums text-side">{{ padCount(results.metrics.length) }}</span>
                    </ComboboxLabel>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <ComboboxItem
                        v-for="metric in results.metrics"
                        :key="metricKey(metric)"
                        :value="metricKey(metric)"
                        :text-value="metric.name"
                        class="text-left p-3 rounded-xl border border-solid bg-ground hover:border-line hover:bg-surface data-[highlighted]:border-line data-[highlighted]:bg-surface data-[highlighted]:ring-1 data-[highlighted]:ring-line transition-colors cursor-pointer flex flex-col gap-2 focus:outline-none"
                        :class="pinnedKey === metricKey(metric) ? 'border-line bg-surface' : 'border-lineDim'"
                      >
                        <div class="flex items-start justify-between gap-2">
                          <span class="text-sm font-semibold text-main leading-snug">
                            {{ metric.name }}
                          </span>
                          <span class="px-1.5 py-0.5 rounded-md text-[10px] uppercase border border-solid border-lineDim text-side shrink-0">
                            {{ metric.tier }}
                          </span>
                        </div>

                        <div class="text-xs font-sans text-dull truncate">
                          {{ metric.code }}
                          <span v-if="metric.displayUnitSymbol || metric.baseUnitSymbol">
                            ({{ metric.displayUnitSymbol || metric.baseUnitSymbol }})
                          </span>
                        </div>

                        <ul
                          v-if="metricLiveRows(metric).length > 0"
                          class="flex flex-col gap-1 pt-1.5 mt-auto border-t border-solid border-lineDim/70"
                        >
                          <li
                            v-for="row in metricLiveRows(metric)"
                            :key="row.key"
                            class="flex items-center justify-between gap-2 text-xs"
                          >
                            <span class="text-side truncate">{{ row.label }}</span>
                            <span
                              class="hud-value inline-flex items-center gap-1 shrink-0"
                              :class="statusTextClass(row.status)"
                            >
                              <span>{{ row.display }}</span>
                              <CheckCircle2
                                v-if="row.status === ReadingStatus.StatusNormal"
                                class="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                              <AlertTriangle
                                v-else-if="row.status === ReadingStatus.StatusAbnormal"
                                class="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                              <AlertCircle
                                v-else-if="row.status === ReadingStatus.StatusCritical"
                                class="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                              <WifiOff
                                v-else-if="row.status === ReadingStatus.StatusStale"
                                class="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                              <Minus
                                v-else
                                class="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                              <span class="sr-only">({{ statusLabel(row.status) }})</span>
                            </span>
                          </li>
                        </ul>
                      </ComboboxItem>
                    </div>
                  </ComboboxGroup>

                  <!-- Medical Procedures Group -->
                  <ComboboxGroup
                    v-else-if="group === 'procedures' && results.procedures.length > 0"
                    class="flex flex-col gap-2"
                  >
                    <ComboboxLabel
                      class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-[0.14em] text-dull pb-1 border-b border-solid border-lineDim"
                    >
                      <span>Medical Procedures</span>
                      <span class="tabular-nums text-side">{{ padCount(results.procedures.length) }}</span>
                    </ComboboxLabel>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <ComboboxItem
                        v-for="proc in results.procedures"
                        :key="procKey(proc)"
                        :value="procKey(proc)"
                        :text-value="proc.title"
                        class="text-left p-3 rounded-xl border border-solid bg-ground hover:border-line hover:bg-surface data-[highlighted]:border-line data-[highlighted]:bg-surface data-[highlighted]:ring-1 data-[highlighted]:ring-line transition-colors cursor-pointer flex flex-col gap-2 focus:outline-none"
                        :class="pinnedKey === procKey(proc) ? 'border-line bg-surface' : 'border-lineDim'"
                      >
                        <div class="flex items-start justify-between gap-2">
                          <h3 class="text-sm font-semibold text-main leading-snug">
                            {{ proc.title }}
                          </h3>
                        </div>

                        <p class="text-xs text-side line-clamp-2 leading-relaxed">
                          {{ proc.whenToUse || proc.summary }}
                        </p>

                        <div class="flex flex-wrap items-center gap-1.5 mt-auto pt-1">
                          <span
                            class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium border border-solid"
                            :class="escalationShort(proc.escalation).classes"
                          >
                            <component
                              :is="escalationShort(proc.escalation).icon"
                              class="w-3 h-3 shrink-0"
                              aria-hidden="true"
                            />
                            <span>{{ escalationShort(proc.escalation).label }}</span>
                          </span>
                          <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] text-dull border border-solid border-lineDim tabular-nums">
                            <Clock class="w-2.5 h-2.5" aria-hidden="true" />
                            <span>{{ proc.estimatedMinutes }}m</span>
                          </span>
                          <span
                            v-for="mCode in proc.relatedMetricCodes.slice(0, 2)"
                            :key="mCode"
                            class="px-1.5 py-0.5 rounded-md text-[10px] font-sans text-dull border border-solid border-lineDim"
                          >
                            {{ mCode }}
                          </span>
                        </div>
                      </ComboboxItem>
                    </div>
                  </ComboboxGroup>

                  <!-- NASA Hazards Group -->
                  <ComboboxGroup
                    v-else-if="group === 'hazards' && results.hazards.length > 0"
                    class="flex flex-col gap-2"
                  >
                    <ComboboxLabel
                      class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-[0.14em] text-dull pb-1 border-b border-solid border-lineDim"
                    >
                      <span>NASA Hazards</span>
                      <span class="tabular-nums text-side">{{ padCount(results.hazards.length) }}</span>
                    </ComboboxLabel>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <ComboboxItem
                        v-for="hazard in results.hazards"
                        :key="hazardKey(hazard)"
                        :value="hazardKey(hazard)"
                        :text-value="hazard.name"
                        class="text-left p-3 rounded-xl border border-solid bg-ground hover:border-line hover:bg-surface data-[highlighted]:border-line data-[highlighted]:bg-surface data-[highlighted]:ring-1 data-[highlighted]:ring-line transition-colors cursor-pointer flex flex-col gap-2 focus:outline-none"
                        :class="pinnedKey === hazardKey(hazard) ? 'border-line bg-surface' : 'border-lineDim'"
                      >
                        <div class="flex items-center justify-between gap-2">
                          <span class="inline-flex items-center gap-1.5 text-sm font-semibold text-main leading-snug">
                            <ShieldAlert class="w-3.5 h-3.5 text-warn shrink-0" aria-hidden="true" />
                            <span>{{ hazard.name }}</span>
                          </span>
                        </div>

                        <p class="text-xs text-side line-clamp-2 leading-relaxed">
                          {{ hazard.summary }}
                        </p>

                        <div class="flex items-center justify-between gap-2 mt-auto pt-1 text-[11px] text-dull">
                          <span class="tabular-nums">{{ hazard.relatedMetricCodes.length }} linked metrics</span>
                          <span class="font-sans">{{ hazard.code }}</span>
                        </div>
                      </ComboboxItem>
                    </div>
                  </ComboboxGroup>
                </template>

                <!-- Empty State -->
                <ComboboxEmpty class="flex-1 flex flex-col items-center justify-center text-center py-12 px-4 gap-3">
                  <template v-if="loading && !hasLoadedOnce">
                    <RefreshCw class="w-5 h-5 text-main animate-spin" aria-hidden="true" />
                    <p class="text-sm text-side">Loading HRIG search index…</p>
                  </template>
                  <template v-else>
                    <p class="text-sm font-medium text-main">
                      No matching entries found<span v-if="query.trim()"> for “{{ query.trim() }}”</span>.
                    </p>
                    <p class="text-xs text-dull max-w-sm">
                      Try a symptom (e.g. hypoxia, headache, dizziness), metric code (spo2, cabin_o2), hazard, or crew name.
                    </p>
                    <button
                      type="button"
                      class="font-sans px-3 py-1.5 rounded-xl text-xs font-medium border border-solid border-line bg-ground text-main hover:bg-surface transition-colors cursor-pointer mt-1"
                      @click="resetFilters"
                    >
                      Reset filters
                    </button>
                  </template>
                </ComboboxEmpty>
              </ComboboxViewport>
            </ComboboxContent>
          </div>
        </div>

        <!-- Bottom Footer Bar -->
        <footer class="px-4 py-2.5 border-t border-solid border-lineDim flex flex-wrap items-center justify-between gap-3 text-xs text-dull shrink-0 bg-ground">
          <div class="tabular-nums text-side">
            <span>{{ totalResults }} {{ totalResults === 1 ? "result" : "results" }}</span>
            <span v-if="query.trim()"> for '{{ query.trim() }}'</span>
          </div>

          <div class="flex flex-wrap items-center gap-3">
            <span class="inline-flex items-center gap-1">
              <kbd class="font-sans px-1.5 py-0.5 rounded-md text-[10px] font-medium text-side border border-solid border-lineDim">↑</kbd>
              <kbd class="font-sans px-1.5 py-0.5 rounded-md text-[10px] font-medium text-side border border-solid border-lineDim">↓</kbd>
              <span>Navigate</span>
            </span>
            <span class="inline-flex items-center gap-1">
              <kbd class="font-sans px-1.5 py-0.5 rounded-md text-[10px] font-medium text-side border border-solid border-lineDim">Enter</kbd>
              <span>Select</span>
            </span>
            <button
              type="button"
              class="font-sans inline-flex items-center gap-1 text-dull hover:text-main transition-colors cursor-pointer"
              @click="closeSearch"
            >
              <kbd class="font-sans px-1.5 py-0.5 rounded-md text-[10px] font-medium text-side border border-solid border-lineDim">Esc</kbd>
              <span>Close</span>
            </button>
          </div>
        </footer>
      </ComboboxRoot>
    </div>
  </main>
</template>
