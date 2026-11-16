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
export const lastTriggeredAlert = ref<AlertDispatchPayload | null>(null)
export const previousNonAlertPath = ref<string | null>(null)

const ALERT_ROUTE_NAMES = new Set<string>(["support-alert", "warning-alert"])
const NON_ALERT_ROUTE_NAMES = new Set<string>(["home", "search", "dashboard-principal", "other-stats"])

const STATUS_RANK: Record<NavReadingStatus, number> = {
  normal: 0,
  stale: 1,
  abnormal: 2,
  critical: 3,
}

const handledEpisodeKeys = new Set<string>()
const guardedRouters = new WeakSet<Router>()

let registeredRouter: Router | null = null

export function buildAlertEpisodeKey(payload: {
  status: NavReadingStatus
  astronautId?: string | null
  metricCode?: string | null
}): string {
  const astronautId = payload.astronautId?.trim() || "env"
  const metricCode = payload.metricCode?.trim() || "unknown"
  return `${payload.status}:${astronautId}:${metricCode}`
}

function suppressCurrentEpisode(): void {
  if (
    lastTriggeredAlert.value &&
    (lastStatus.value === "abnormal" || lastStatus.value === "critical")
  ) {
    handledEpisodeKeys.add(buildAlertEpisodeKey(lastTriggeredAlert.value))
  }
}

function isAlertRouteName(name: unknown): boolean {
  return typeof name === "string" && ALERT_ROUTE_NAMES.has(name)
}

function isNonAlertRouteName(name: unknown): boolean {
  return typeof name === "string" && NON_ALERT_ROUTE_NAMES.has(name)
}

function capturePreviousNonAlertRoute(router: Router): void {
  const current = router.currentRoute.value
  if (isNonAlertRouteName(current.name)) {
    previousNonAlertPath.value = current.fullPath
  }
}

export function registerTelemetryRouter(router: Router): void {
  registeredRouter = router
  if (!guardedRouters.has(router)) {
    guardedRouters.add(router)
    router.beforeEach((to, from) => {
      if (isNonAlertRouteName(from.name) && isAlertRouteName(to.name)) {
        previousNonAlertPath.value = from.fullPath
      }
    })
  }
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
  suppressCurrentEpisode()

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

function openOtherStatsWithRouter(
  router: Router | null,
  astronaut?: { id: string; fullName?: string } | string,
): void {
  suppressCurrentEpisode()

  let id: string | undefined
  if (astronaut !== undefined) {
    id = typeof astronaut === "string" ? astronaut : astronaut.id
    const fullName = typeof astronaut === "string" ? undefined : astronaut.fullName
    setSelectedAstronaut(id, fullName)
  } else if (selectedAstronautId.value.trim() !== "") {
    id = selectedAstronautId.value.trim()
  }

  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({
      name: "other-stats",
      query: id ? { astronaut: id } : undefined,
    })
  }
}

export function openOtherStats(astronaut?: { id: string; fullName?: string } | string): void {
  openOtherStatsWithRouter(null, astronaut)
}

export const openSecondaryStats = openOtherStats

function goToHomeWithRouter(router: Router | null): void {
  suppressCurrentEpisode()

  const targetRouter = resolveRouter(router)
  if (targetRouter) {
    void targetRouter.push({ name: "home" })
  }
}

function openActiveAlertWithRouter(router: Router | null): void {
  const targetRouter = resolveRouter(router)
  if (!targetRouter) {
    return
  }
  if (lastStatus.value === "critical") {
    capturePreviousNonAlertRoute(targetRouter)
    void targetRouter.push({ name: "warning-alert" })
  } else if (lastStatus.value === "abnormal") {
    capturePreviousNonAlertRoute(targetRouter)
    void targetRouter.push({ name: "support-alert" })
  }
}

function closeActiveAlertWithRouter(router: Router | null): void {
  suppressCurrentEpisode()

  const targetRouter = resolveRouter(router)
  if (!targetRouter) {
    return
  }

  const returnPath = previousNonAlertPath.value
  previousNonAlertPath.value = null
  if (returnPath) {
    void targetRouter.push(returnPath)
  } else {
    void targetRouter.push({ name: "home" })
  }
}

export function closeActiveAlert(): void {
  closeActiveAlertWithRouter(null)
}

function dispatchTelemetryStatusWithRouter(
  router: Router | null,
  payload: AlertDispatchPayload,
): void {
  if (payload.status === "normal") {
    clearAlertIfResolvedWithRouter(router)
    return
  }

  if (payload.status === "stale") {
    if (STATUS_RANK[lastStatus.value] <= STATUS_RANK.stale) {
      lastStatus.value = "stale"
    }
    return
  }

  // Keep higher-severity active alert active unless overridden by a new demo trigger
  const currentRank = STATUS_RANK[lastStatus.value]
  const nextRank = STATUS_RANK[payload.status]
  if (nextRank < currentRank && lastTriggeredAlert.value?.source !== "mockup-demo") {
    return
  }

  const episodeKey = buildAlertEpisodeKey(payload)
  const isNewEpisode = !handledEpisodeKeys.has(episodeKey)
  handledEpisodeKeys.add(episodeKey)

  lastStatus.value = payload.status
  lastTriggeredAlert.value = payload

  if (!isNewEpisode) {
    return
  }

  const targetRouter = resolveRouter(router)
  const targetRouteName = payload.status === "critical" ? "warning-alert" : "support-alert"
  if (targetRouter && targetRouter.currentRoute.value.name !== targetRouteName) {
    capturePreviousNonAlertRoute(targetRouter)
    void targetRouter.push({ name: targetRouteName })
  }
}

export function dispatchTelemetryStatus(payload: AlertDispatchPayload): void {
  dispatchTelemetryStatusWithRouter(null, payload)
}

function clearAlertIfResolvedWithRouter(
  router: Router | null,
  remainingAlert: AlertDispatchPayload | null = null,
  activeEpisodeKeys?: ReadonlySet<string>,
): void {
  if (activeEpisodeKeys) {
    for (const key of handledEpisodeKeys) {
      if (!activeEpisodeKeys.has(key)) {
        handledEpisodeKeys.delete(key)
      }
    }
  }

  if (!remainingAlert) {
    lastStatus.value = "normal"
    lastTriggeredAlert.value = null
    handledEpisodeKeys.clear()

    const targetRouter = resolveRouter(router)
    if (targetRouter && isAlertRouteName(targetRouter.currentRoute.value.name)) {
      closeActiveAlertWithRouter(targetRouter)
    }
    return
  }

  lastStatus.value = remainingAlert.status
  lastTriggeredAlert.value = remainingAlert

  const targetRouter = resolveRouter(router)
  if (
    targetRouter &&
    targetRouter.currentRoute.value.name === "warning-alert" &&
    remainingAlert.status === "abnormal"
  ) {
    void targetRouter.push({ name: "support-alert" })
  }
}

export function clearAlertIfResolved(
  remainingAlert: AlertDispatchPayload | null = null,
  activeEpisodeKeys?: ReadonlySet<string>,
): void {
  clearAlertIfResolvedWithRouter(null, remainingAlert, activeEpisodeKeys)
}

function resetDemoStatusWithRouter(router: Router | null): void {
  lastStatus.value = "normal"
  lastTriggeredAlert.value = null
  handledEpisodeKeys.clear()

  const targetRouter = resolveRouter(router)
  if (targetRouter && isAlertRouteName(targetRouter.currentRoute.value.name)) {
    closeActiveAlertWithRouter(targetRouter)
  }
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
    previousNonAlertPath,
    setSelectedAstronaut,
    openPrimaryVitals: (astronaut: { id: string; fullName?: string } | string) =>
      openPrimaryVitalsWithRouter(router, astronaut),
    openOtherStats: (astronaut?: { id: string; fullName?: string } | string) =>
      openOtherStatsWithRouter(router, astronaut),
    openSecondaryStats: (astronaut?: { id: string; fullName?: string } | string) =>
      openOtherStatsWithRouter(router, astronaut),
    goToHome: () => goToHomeWithRouter(router),
    openActiveAlert: () => openActiveAlertWithRouter(router),
    closeActiveAlert: () => closeActiveAlertWithRouter(router),
    dispatchTelemetryStatus: (payload: AlertDispatchPayload) =>
      dispatchTelemetryStatusWithRouter(router, payload),
    clearAlertIfResolved: (
      remainingAlert: AlertDispatchPayload | null = null,
      activeEpisodeKeys?: ReadonlySet<string>,
    ) => clearAlertIfResolvedWithRouter(router, remainingAlert, activeEpisodeKeys),
    resetDemoStatus: () => resetDemoStatusWithRouter(router),
  }
}
