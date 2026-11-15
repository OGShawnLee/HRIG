<script setup lang="ts">
import { computed } from "vue"
import { AlertTriangle, Home, X } from "lucide-vue-next"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const { lastTriggeredAlert, goToHome, closeActiveAlert } = useTelemetryNavigation()

const activeAlert = computed(() =>
  lastTriggeredAlert.value?.status === "abnormal" ? lastTriggeredAlert.value : null,
)

const conditionHeading = computed(() => {
  const alert = activeAlert.value
  if (!alert) {
    return "LOW BLOOD PRESSURE DETECTED"
  }
  if (alert.message && alert.message.trim() !== "") {
    return alert.message.trim()
  }
  if (alert.metricCode && alert.metricCode !== "bp_systolic" && alert.metricCode !== "bp_diastolic") {
    return `ABNORMAL TELEMETRY: ${alert.metricCode.toUpperCase()}`
  }
  return "LOW BLOOD PRESSURE DETECTED"
})

const readingSummary = computed(() => {
  const alert = activeAlert.value
  if (alert?.valueDisplay) {
    return alert.valueDisplay
  }
  return "86/55 mmHg"
})

const nominalRangeNote = computed(() => {
  const code = activeAlert.value?.metricCode
  if (!code || code === "bp_systolic" || code === "bp_diastolic") {
    return "(Nominal Range: 90–120 / 60–80 mmHg)"
  }
  return ""
})

const metadataLine = computed(() => {
  const alert = activeAlert.value
  if (!alert) {
    return "Source: Blood pressure cuff (baumanometro_01) · Demo sample · MET 000:00:00 · 2026-11-15T21:42:00Z (UTC)"
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
  } else {
    parts.push("Source: Blood pressure cuff (baumanometro_01)")
  }
  if (alert.metricCode) {
    parts.push(`Metric: ${alert.metricCode}`)
  }
  parts.push(`MET ${alert.met || "000:00:00"}`)
  parts.push(alert.timestampUtc ? `${alert.timestampUtc} (UTC)` : "2026-11-15T21:42:00Z (UTC)")
  return parts.join(" · ")
})
</script>

<template>
  <main
    class="flex-1 w-full bg-ground text-main px-4 sm:px-8 py-6 flex flex-col items-center gap-5"
    aria-label="Caution low blood pressure alert view"
  >
    <div class="w-full max-w-4xl flex items-center justify-between gap-3">
      <button
        type="button"
        class="button-side h-10 px-3 text-xs gap-1.5"
        aria-label="Return to Home"
        @click="goToHome"
      >
        <Home class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Home</span>
      </button>

      <button
        type="button"
        class="button-side border-warn text-warn h-10 px-3 text-xs gap-1.5"
        aria-label="Close alert and return to previous page"
        @click="closeActiveAlert"
      >
        <X class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>Close Alert</span>
      </button>
    </div>

    <section class="w-full max-w-4xl hud-card border-warn p-6 flex flex-col gap-3">
      <header class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-2.5">
          <AlertTriangle class="w-6 h-6 text-warn shrink-0" aria-hidden="true" />
          <h1 class="text-xl sm:text-2xl font-bold tracking-wider text-warn">SUPPORT</h1>
        </div>
        <span
          class="hud-pill border-warn text-warn px-3 py-1 text-xs font-bold tracking-widest"
        >
          CAUTION
        </span>
      </header>

      <h2 class="text-base sm:text-lg font-bold text-main">
        {{ conditionHeading }}
      </h2>

      <p class="text-sm sm:text-base font-semibold text-main hud-value">
        Current Reading: {{ readingSummary }}
        <em v-if="nominalRangeNote" class="text-side font-normal not-italic ml-1">
          {{ nominalRangeNote }}
        </em>
      </p>

      <p class="text-xs sm:text-sm text-dull hud-value break-words">
        {{ metadataLine }}
      </p>

      <p class="text-sm sm:text-base font-semibold text-main pt-1">
        Your blood pressure is below normal. Follow the option below that matches your current condition
      </p>
    </section>

    <section class="w-full max-w-4xl hud-card border-lineDim bg-surface p-6 flex flex-col gap-3">
      <h2 class="text-sm sm:text-base font-bold tracking-wide text-main">
        RECOMMENDED ACTIONS
      </h2>
      <!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->
      <ol class="list-decimal pl-5 flex flex-col gap-3 text-sm sm:text-base text-side leading-relaxed">
        <li>
          <strong class="text-main">If you feel dizzy, lightheaded, or have tunnel vision</strong> → Sit down or strap into a restraint immediately, drink 250ml of water with electrolytes, and remain seated for 10 minutes before re-measuring
        </li>
        <li>
          <strong class="text-main">If you just transitioned to Lunar Gravity or stood up quickly</strong> → Review the “Orthostatic Intolerance Countermeasures”
        </li>
        <li>
          <strong class="text-main">If you just finished intense exercise or an EVA</strong> → Check your Body Temperature and Heart Rate for dehydration signs.
        </li>
      </ol>
    </section>
  </main>
</template>
