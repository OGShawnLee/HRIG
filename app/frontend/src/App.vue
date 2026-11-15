<script setup lang="ts">
import { onBeforeUnmount, onMounted } from "vue"
import { RouterView, useRouter } from "vue-router"
import DemoAlertStrip from "@/components/DemoAlertStrip.vue"
import ViewNavBar from "@/components/ViewNavBar.vue"
import WindowBar from "@/components/WindowBar.vue"
import {
  TELEMETRY_STATUS_EVENT,
  registerTelemetryRouter,
  useTelemetryNavigation,
} from "@/composables/useTelemetryNavigation"
import type { AlertDispatchPayload, ReadingStatus } from "@/composables/useTelemetryNavigation"

const router = useRouter()
registerTelemetryRouter(router)

const { dispatchTelemetryStatus } = useTelemetryNavigation()

function isReadingStatus(value: unknown): value is ReadingStatus {
  return value === "normal" || value === "abnormal" || value === "critical" || value === "stale"
}

function onTelemetryStatusEvent(event: Event) {
  const customEvent = event as CustomEvent<AlertDispatchPayload | undefined>
  const detail = customEvent.detail
  if (!detail || !isReadingStatus(detail.status)) {
    return
  }
  dispatchTelemetryStatus(detail)
}

onMounted(() => {
  window.addEventListener(TELEMETRY_STATUS_EVENT, onTelemetryStatusEvent)
})

onBeforeUnmount(() => {
  window.removeEventListener(TELEMETRY_STATUS_EVENT, onTelemetryStatusEvent)
})
</script>

<template>
  <div class="min-h-screen w-full flex flex-col text-main">
    <WindowBar />
    <ViewNavBar class="shrink-0" />
    <div class="flex-1 flex flex-col">
      <RouterView />
    </div>
    <DemoAlertStrip class="shrink-0 sticky bottom-0 z-40" />
  </div>
</template>
