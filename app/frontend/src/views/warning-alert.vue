<script setup lang="ts">
import { computed } from "vue"
import { AlertCircle, ArrowDownRight, Home } from "lucide-vue-next"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const { lastTriggeredAlert, goToHome } = useTelemetryNavigation()

const activeAlert = computed(() =>
  lastTriggeredAlert.value?.status === "critical" ? lastTriggeredAlert.value : null,
)

const primaryIssueLine = computed(() => {
  const alert = activeAlert.value
  if (!alert) {
    return "Pressure: 78 kPa (Drop below 90 kPa)"
  }
  if (alert.valueDisplay && alert.metricCode) {
    return `${alert.metricCode}: ${alert.valueDisplay}`
  }
  if (alert.valueDisplay) {
    return `Current Reading: ${alert.valueDisplay}`
  }
  if (alert.metricCode && alert.metricCode !== "cabin_pressure") {
    return `Critical Metric: ${alert.metricCode}`
  }
  return "Pressure: 78 kPa (Drop below 90 kPa)"
})

const secondaryRiskLine = computed(() => {
  const alert = activeAlert.value
  if (alert?.message && alert.message.trim() !== "") {
    return alert.message.trim()
  }
  return "Associated Risk: Decompression Sickness (DCS)"
})

const metadataLine = computed(() => {
  const alert = activeAlert.value
  if (!alert) {
    return null
  }
  const parts: string[] = []
  if (alert.astronautName) {
    parts.push(`Crew: ${alert.astronautName}`)
  } else if (alert.astronautId) {
    parts.push(`Crew ID: ${alert.astronautId}`)
  }
  if (alert.source || alert.deviceCode) {
    const dev = alert.deviceCode ? ` (${alert.deviceCode})` : ""
    parts.push(`Source: ${alert.source ?? "telemetry"}${dev}`)
  }
  if (alert.met) {
    parts.push(`MET ${alert.met}`)
  }
  if (alert.timestampUtc) {
    parts.push(`${alert.timestampUtc} (UTC)`)
  }
  return parts.length > 0 ? parts.join(" · ") : null
})
</script>

<template>
  <main
    class="flex-1 w-full bg-ground text-main px-4 sm:px-8 py-6 flex flex-col items-center gap-5"
    aria-label="Warning view"
  >
    <div class="w-full max-w-4xl flex items-center">
      <button
        type="button"
        class="button-side h-10 px-3 text-xs gap-1.5"
        aria-label="Return to Home"
        @click="goToHome"
      >
        <Home class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Home</span>
      </button>
    </div>

    <section class="w-full max-w-4xl hud-card border-crit p-6 flex items-start justify-between gap-4">
      <div class="flex flex-col gap-3 min-w-0 flex-1">
        <header class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-2.5">
            <AlertCircle class="w-6 h-6 text-crit shrink-0" aria-hidden="true" />
            <h1 class="text-xl sm:text-2xl font-bold tracking-wider text-crit">WARNING</h1>
          </div>
          <span
            class="hud-pill border-crit text-crit px-3 py-1 text-xs font-bold tracking-widest"
          >
            CRITICAL
          </span>
        </header>

        <p class="text-base sm:text-lg font-bold text-main hud-value">
          {{ primaryIssueLine }}
        </p>

        <p class="text-sm sm:text-base font-semibold text-side">
          {{ secondaryRiskLine }}
        </p>

        <p v-if="metadataLine" class="text-xs sm:text-sm text-dull hud-value break-words">
          {{ metadataLine }}
        </p>
      </div>

      <ArrowDownRight class="w-16 h-16 sm:w-20 sm:h-20 text-crit shrink-0 hidden sm:block" aria-hidden="true" />
    </section>

    <div class="w-full max-w-4xl flex flex-col gap-2">
      <h2 class="text-xs sm:text-sm font-bold tracking-wider text-dull px-1">
        REQUIRED ACTIONS
      </h2>

      <section class="w-full hud-card border-lineDim bg-surface p-6 flex flex-col gap-3">
        <h3 class="text-sm sm:text-base font-bold tracking-wide text-main">
          IMMEDIATE EMERGENCY PROCEDURES
        </h3>
        <ol class="list-decimal pl-5 flex flex-col gap-3 text-sm sm:text-base text-side leading-relaxed">
          <li>
            <strong class="text-main">INITIATE 100% O2 IN-SUIT PURGE</strong> - Purge nitrogen immediately to prevent gas bubbles
          </li>
          <li>
            <strong class="text-main">INCREASE CABIN / SUIT PRESSURE</strong> - Initiate repressurization or don pressure suits
          </li>
          <li>
            <strong class="text-main">DOWNLINK &amp; MONITOR</strong> - Alert Crew Medical Officer (CMO) and flight surgeon.
          </li>
        </ol>
      </section>
    </div>
  </main>
</template>
