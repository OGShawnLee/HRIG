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
  type AlertDispatchPayload,
  type NavReadingStatus,
} from "@/composables/useTelemetryNavigation"

const router = useRouter()
registerTelemetryRouter(router)

const { dispatchTelemetryStatus } = useTelemetryNavigation()

function isNavReadingStatus(value: unknown): value is NavReadingStatus {
  return value === "normal" || value === "abnormal" || value === "critical" || value === "stale"
}

function onTelemetryStatusEvent(event: Event) {
  const customEvent = event as CustomEvent<AlertDispatchPayload | undefined>
  const detail = customEvent.detail
  if (!detail || !isNavReadingStatus(detail.status)) {
    return
  }
  dispatchTelemetryStatus(detail)
}

function onGlobalKeydown(event: KeyboardEvent) {
  if ((event.ctrlKey || event.metaKey) && !event.altKey && event.key.toLowerCase() === "k") {
    event.preventDefault()
    if (router.currentRoute.value.name === "search") {
      window.dispatchEvent(new CustomEvent("hrig:focus-omnibox"))
    } else {
      void router.push({ name: "search" })
    }
  }
}

onMounted(() => {
  window.addEventListener(TELEMETRY_STATUS_EVENT, onTelemetryStatusEvent)
  window.addEventListener("keydown", onGlobalKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener(TELEMETRY_STATUS_EVENT, onTelemetryStatusEvent)
  window.removeEventListener("keydown", onGlobalKeydown)
})
</script>

<template>
  <div class="min-h-screen w-full flex flex-col bg-ground text-main">
    <WindowBar />
    <ViewNavBar class="shrink-0" />
    <div class="flex-1 flex flex-col">
      <RouterView />
    </div>
    <DemoAlertStrip class="shrink-0 sticky bottom-0 z-40" />
  </div>
</template>
