<script setup lang="ts">
import { computed, type Component } from "vue"
import { AlertCircle, AlertTriangle, CheckCircle2, Home, WifiOff } from "lucide-vue-next"
import {
  useTelemetryNavigation,
  type NavReadingStatus,
} from "@/composables/useTelemetryNavigation"

const { lastStatus, goToHome, openActiveAlert, dispatchTelemetryStatus, resetDemoStatus } =
  useTelemetryNavigation()

const STATUS_META: Record<
  NavReadingStatus,
  { icon: Component; label: string; toneClass: string }
> = {
  normal: { icon: CheckCircle2, label: "NOMINAL", toneClass: "text-ok" },
  abnormal: { icon: AlertTriangle, label: "CAUTION / ABNORMAL", toneClass: "text-warn border-warn" },
  critical: { icon: AlertCircle, label: "CRITICAL", toneClass: "text-crit border-crit" },
  stale: { icon: WifiOff, label: "STALE", toneClass: "text-stale border-dashed" },
}

const currentStatusMeta = computed(() => STATUS_META[lastStatus.value])
const isAlertNavigable = computed(
  () => lastStatus.value === "abnormal" || lastStatus.value === "critical",
)
const alertButtonLabel = computed(() =>
  lastStatus.value === "critical"
    ? "Open critical warning alert view"
    : "Open caution support alert view",
)
</script>

<template>
  <aside
    class="bg-ground border-t border-solid border-lineDim px-4 py-2 flex flex-wrap items-center justify-between gap-2 text-xs text-main hud-value"
    role="region"
    aria-label="Navigation and alert simulator"
  >
    <button
      v-if="isAlertNavigable"
      type="button"
      class="hud-pill px-3 py-1.5 inline-flex items-center gap-1.5 border-lineDim font-semibold tracking-wider cursor-pointer hover:bg-surface transition-colors"
      :class="currentStatusMeta.toneClass"
      :aria-label="alertButtonLabel"
      :title="alertButtonLabel"
      aria-live="polite"
      @click="openActiveAlert()"
    >
      <component :is="currentStatusMeta.icon" class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
      <span>{{ currentStatusMeta.label }}</span>
    </button>
    <div
      v-else
      class="hud-pill px-3 py-1.5 inline-flex items-center gap-1.5 border-lineDim font-semibold tracking-wider"
      :class="currentStatusMeta.toneClass"
      aria-live="polite"
    >
      <component :is="currentStatusMeta.icon" class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
      <span>{{ currentStatusMeta.label }}</span>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="button-side h-9 px-3 text-xs gap-1.5"
        @click="goToHome()"
      >
        <Home class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Home</span>
      </button>
      <button
        type="button"
        class="button-side h-9 px-3 text-xs gap-1.5 border-warn text-warn"
        @click="
          dispatchTelemetryStatus({
            status: 'abnormal',
            metricCode: 'bp_systolic',
            source: 'mockup-demo',
          })
        "
      >
        <AlertTriangle class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Simulate Abnormal</span>
      </button>
      <button
        type="button"
        class="button-side h-9 px-3 text-xs gap-1.5 border-crit text-crit"
        @click="
          dispatchTelemetryStatus({
            status: 'critical',
            metricCode: 'cabin_pressure',
            source: 'mockup-demo',
          })
        "
      >
        <AlertCircle class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Simulate Critical</span>
      </button>
      <button
        type="button"
        class="button-side h-9 px-3 text-xs"
        @click="resetDemoStatus()"
      >
        Reset
      </button>
    </div>
  </aside>
</template>
