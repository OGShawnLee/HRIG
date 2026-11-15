import { computed, getCurrentInstance, ref } from "vue"
import { useRouter } from "vue-router"
import type { Router } from "vue-router"

export type ReadingStatus = "normal" | "abnormal" | "critical" | "stale"

export interface AlertDispatchPayload {
  status: ReadingStatus
  astronautId?: string
  metricCode?: string
  source?: string
}

export const CREW_MEMBERS = {
  sam: "Sam Sonia",
  patrick: "Patrick Cuckburn",
} as const

export type CrewMemberId = keyof typeof CREW_MEMBERS

export const TELEMETRY_STATUS_EVENT = "hrig:telemetry-status"

const selectedAstronautId = ref<CrewMemberId>("sam")
const selectedAstronautName = computed(() => CREW_MEMBERS[selectedAstronautId.value])
const lastStatus = ref<ReadingStatus>("normal")
const lastTriggeredAlert = ref<AlertDispatchPayload | null>(null)

let registeredRouter: Router | null = null

export function registerTelemetryRouter(router: Router): void {
  registeredRouter = router
}

function resolveRouter(componentRouter?: Router | null): Router | null {
  return registeredRouter ?? componentRouter ?? null
}

function setSelectedAstronaut(id: string): void {
  if (id === "sam" || id === "patrick") {
    selectedAstronautId.value = id
  }
}

function openPrimaryVitalsWithRouter(router: Router | null, astronautId: CrewMemberId): void {
  selectedAstronautId.value = astronautId
  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({
      name: "dashboard-principal",
      query: { astronaut: astronautId },
    })
  }
}

function goToCrewOverviewWithRouter(router: Router | null): void {
  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({ name: "pagina-principal" })
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

export function handleIncomingStatus(
  status: ReadingStatus,
  context?: Omit<AlertDispatchPayload, "status">,
): void {
  dispatchTelemetryStatusWithRouter(null, { status, ...context })
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
    selectedAstronautName,
    lastStatus,
    lastTriggeredAlert,
    setSelectedAstronaut,
    openPrimaryVitals: (astronautId: CrewMemberId) =>
      openPrimaryVitalsWithRouter(router, astronautId),
    goToCrewOverview: () => goToCrewOverviewWithRouter(router),
    dispatchTelemetryStatus: (payload: AlertDispatchPayload) =>
      dispatchTelemetryStatusWithRouter(router, payload),
    handleIncomingStatus: (
      status: ReadingStatus,
      context?: Omit<AlertDispatchPayload, "status">,
    ) => dispatchTelemetryStatusWithRouter(router, { status, ...context }),
    resetDemoStatus,
  }
}
