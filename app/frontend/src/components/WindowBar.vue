<script setup lang="ts">
import { computed } from "vue"
import { Wifi, WifiOff } from "lucide-vue-next"
import { OnClose, OnMaximize, OnMinimize } from "@bindings/hrig/app/app.js"
import { useTelemetry } from "@/composables/useTelemetry"

const { connectionState } = useTelemetry()

const isLive = computed(() => connectionState.value.connected)
const hasConnError = computed(() => !isLive.value && Boolean(connectionState.value.lastError))
const statusLabel = computed(() => (isLive.value ? "Live" : "Reconnecting"))
const statusTitle = computed(() => {
  if (isLive.value) {
    return connectionState.value.url
      ? `Real-time telemetry connected (${connectionState.value.url})`
      : "Real-time telemetry connected"
  }
  return connectionState.value.lastError
    ? `Reconnecting: ${connectionState.value.lastError}`
    : "Connecting to real-time telemetry stream"
})
</script>

<template>
  <header
    class="sticky top-0 z-50 h-10 w-full bg-ground border-b border-solid border-lineDim px-4 grid grid-cols-12 items-center select-none"
    style="--wails-draggable: drag; -webkit-app-region: drag;"
    @dblclick="OnMaximize"
  >
    <div
      class="col-span-4 flex items-center gap-2"
      style="--wails-draggable: no-drag; -webkit-app-region: no-drag;"
    >
      <button
        type="button"
        aria-label="Close window"
        title="Close"
        class="w-3 h-3 rounded-full bg-[#FF5F56] border border-solid border-[#E0443E]"
        @click="OnClose"
      />
      <button
        type="button"
        aria-label="Minimize window"
        title="Minimize"
        class="w-3 h-3 rounded-full bg-[#FFBD2E] border border-solid border-[#DEA123]"
        @click="OnMinimize"
      />
      <button
        type="button"
        aria-label="Maximize window"
        title="Maximize"
        class="w-3 h-3 rounded-full bg-[#27C93F] border border-solid border-[#1AAB29]"
        @click="OnMaximize"
      />
    </div>
    <span class="col-span-4 text-center font-black tracking-tight text-main leading-tight">
      HRIG
    </span>
    <div
      class="col-span-4 flex items-center justify-end"
      style="--wails-draggable: no-drag; -webkit-app-region: no-drag;"
    >
      <span
        class="inline-flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-full border border-solid border-lineDim bg-surface"
        :class="isLive ? 'text-ok' : hasConnError ? 'text-warn' : 'text-stale'"
        :title="statusTitle"
        :aria-label="`Telemetry stream status: ${statusLabel}`"
      >
        <Wifi v-if="isLive" class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <WifiOff v-else class="w-3.5 h-3.5 shrink-0" aria-hidden="true" />
        <span>{{ statusLabel }}</span>
      </span>
    </div>
  </header>
</template>
