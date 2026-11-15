import { getCurrentInstance, ref } from "vue"
import { useRouter, type Router } from "vue-router"

export type NavReadingStatus = "normal" | "abnormal" | "critical" | "stale"

export interface AlertDispatchPayload {
  status: NavReadingStatus
  astronautId?: string
  astronautName?: string
  metricCode?: string
  source?: string
  deviceCode?: string
  valueDisplay?: string
  message?: string
  met?: string
  timestampUtc?: string
}

export const TELEMETRY_STATUS_EVENT = "hrig:telemetry-status"

const selectedAstronautId = ref<string>("")
const selectedAstronautFallbackName = ref<string>("Sam Sonia")
const lastStatus = ref<NavReadingStatus>("normal")
const lastTriggeredAlert = ref<AlertDispatchPayload | null>(null)

let registeredRouter: Router | null = null

export function registerTelemetryRouter(router: Router): void {
  registeredRouter = router
}

function resolveRouter(componentRouter?: Router | null): Router | null {
  return registeredRouter ?? componentRouter ?? null
}

function setSelectedAstronaut(id: string, name?: string): void {
  selectedAstronautId.value = id
  if (name && name.trim() !== "") {
    selectedAstronautFallbackName.value = name.trim()
  }
}

function openPrimaryVitalsWithRouter(
  router: Router | null,
  astronaut: { id: string; fullName?: string } | string,
): void {
  const id = typeof astronaut === "string" ? astronaut : astronaut.id
  const fullName = typeof astronaut === "string" ? undefined : astronaut.fullName
  setSelectedAstronaut(id, fullName)

  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({
      name: "dashboard-principal",
      query: { astronaut: id },
    })
  }
}

function goToHomeWithRouter(router: Router | null): void {
  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({ name: "home" })
  }
}

function dispatchTelemetryStatusWithRouter(
  router: Router | null,
  payload: AlertDispatchPayload,
): void {
  lastStatus.value = payload.status
  const targetRouter = resolveRouter(router)

  if (payload.status === "critical") {
    lastTriggeredAlert.value = payload
    if (targetRouter && targetRouter.currentRoute.value.name !== "warning-alert") {
      void targetRouter.push({ name: "warning-alert" })
    }
    return
  }

  if (payload.status === "abnormal") {
    lastTriggeredAlert.value = payload
    if (targetRouter && targetRouter.currentRoute.value.name !== "support-alert") {
      void targetRouter.push({ name: "support-alert" })
    }
    return
  }

  if (payload.status === "normal") {
    lastTriggeredAlert.value = null
  }
}

export function dispatchTelemetryStatus(payload: AlertDispatchPayload): void {
  dispatchTelemetryStatusWithRouter(null, payload)
}

function resetDemoStatus(): void {
  lastStatus.value = "normal"
  lastTriggeredAlert.value = null
}

export function useTelemetryNavigation() {
  const router = getCurrentInstance() ? useRouter() : null
  if (router && !registeredRouter) {
    registeredRouter = router
  }

  return {
    selectedAstronautId,
    selectedAstronautFallbackName,
    lastStatus,
    lastTriggeredAlert,
    setSelectedAstronaut,
    openPrimaryVitals: (astronaut: { id: string; fullName?: string } | string) =>
      openPrimaryVitalsWithRouter(router, astronaut),
    goToHome: () => goToHomeWithRouter(router),
    dispatchTelemetryStatus: (payload: AlertDispatchPayload) =>
      dispatchTelemetryStatusWithRouter(router, payload),
    resetDemoStatus,
  }
}
