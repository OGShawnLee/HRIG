<script setup lang="ts">
import { useRouter } from "vue-router"
import { AlertTriangle, RefreshCw, Search } from "lucide-vue-next"
import CrewCard from "@/components/CrewCard.vue"
import EnvironmentCard from "@/components/EnvironmentCard.vue"
import MedicalManualButton from "@/components/MedicalManualButton.vue"
import { useTelemetry } from "@/composables/useTelemetry"

const router = useRouter()
const { activeAstronauts, loading, error, refresh } = useTelemetry()

function openSearch(initialQuery?: string) {
  if (initialQuery && initialQuery.length > 0) {
    void router.push({ name: "search", query: { q: initialQuery } })
    return
  }
  void router.push({ name: "search" })
}

function openMedicalManual() {
  void router.push({ name: "search", query: { tab: "procedures" } })
}

function handleSearchTriggerKeydown(event: KeyboardEvent) {
  if (
    event.key.length === 1 &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.altKey
  ) {
    event.preventDefault()
    openSearch(event.key)
  }
}
</script>

<template>
  <main class="flex-1 w-full bg-ground text-main px-4 sm:px-8 py-6 sm:py-8 flex flex-col items-center gap-8 sm:gap-10">
    <!-- Top centered Search Trigger -->
    <div class="w-full max-w-xl relative flex items-center">
      <button
        type="button"
        aria-label="Search crew, medical procedures, hazards, and metrics"
        class="field w-full h-12 rounded-2xl pl-12 pr-20 flex items-center justify-center relative text-side font-normal hover:bg-surface focus-visible:ring-2 focus-visible:ring-line cursor-pointer transition-colors"
        @click="openSearch()"
        @keydown="handleSearchTriggerKeydown"
      >
        <Search
          class="w-5 h-5 text-main absolute left-4 pointer-events-none"
          aria-hidden="true"
        />
        <span>Search</span>
        <kbd
          class="hidden sm:inline-flex items-center px-2 py-0.5 text-xs font-medium text-dull border border-solid border-lineDim rounded-md absolute right-4 pointer-events-none"
        >
          Ctrl+K
        </kbd>
      </button>
    </div>

    <!-- Inline error notice if server unreachable while cached crew exists -->
    <div
      v-if="error && activeAstronauts.length > 0"
      role="alert"
      class="w-full max-w-3xl hud-pill border-warn px-4 py-3 flex items-center justify-between gap-4 text-sm text-warn"
    >
      <div class="flex items-center gap-2 min-w-0">
        <AlertTriangle class="w-4 h-4 shrink-0" aria-hidden="true" />
        <span class="truncate">Telemetry sync warning: {{ error }}</span>
      </div>
      <button
        type="button"
        class="button-side h-8 px-3 text-xs shrink-0 gap-1.5"
        :disabled="loading"
        @click="refresh"
      >
        <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
        <span>Retry</span>
      </button>
    </div>

    <!-- Crew Cards Grid / Offline & Empty States -->
    <section
      v-if="loading && activeAstronauts.length === 0"
      aria-label="Loading crew telemetry"
      class="w-full max-w-5xl flex-1 flex items-center justify-center py-12"
    >
      <div class="hud-card px-8 py-6 flex items-center gap-3 text-side">
        <RefreshCw class="w-5 h-5 animate-spin text-main" aria-hidden="true" />
        <span class="text-base font-medium">Loading crew telemetry…</span>
      </div>
    </section>

    <section
      v-else-if="error && activeAstronauts.length === 0"
      aria-label="Telemetry connection error"
      class="w-full max-w-xl flex-1 flex items-center justify-center py-8"
    >
      <div class="hud-card w-full p-6 flex flex-col items-center text-center gap-4">
        <AlertTriangle class="w-8 h-8 text-warn" aria-hidden="true" />
        <div class="flex flex-col gap-1">
          <h2 class="text-lg font-semibold text-main">Telemetry Server Unreachable</h2>
          <p class="text-sm text-dull break-words">{{ error }}</p>
        </div>
        <button
          type="button"
          class="button-main gap-2"
          :disabled="loading"
          @click="refresh"
        >
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
          <span>Retry</span>
        </button>
      </div>
    </section>

    <section
      v-else-if="activeAstronauts.length === 0"
      aria-label="No active crew"
      class="w-full max-w-xl flex-1 flex items-center justify-center py-8"
    >
      <div class="hud-card w-full p-6 flex flex-col items-center text-center gap-4">
        <p class="text-base text-side">No active crew members found.</p>
        <button type="button" class="button-side gap-2" :disabled="loading" @click="refresh">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
          <span>Refresh</span>
        </button>
      </div>
    </section>

    <section
      v-else
      aria-label="Crew vitals"
      class="w-full max-w-5xl grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 justify-items-center"
    >
      <CrewCard
        v-for="astronaut in activeAstronauts"
        :key="astronaut.id"
        :astronaut="astronaut"
        class="w-full max-w-sm"
      />
    </section>

    <!-- Bottom Row: Medical Manual Button (left) + Environment Card (right) -->
    <section
      aria-label="Medical manual and cabin environment"
      class="w-full max-w-4xl flex flex-col sm:flex-row items-center sm:items-stretch justify-between gap-6 mt-auto pt-2"
    >
      <div class="flex items-center justify-center">
        <MedicalManualButton @click="openMedicalManual" />
      </div>
      <EnvironmentCard />
    </section>
  </main>
</template>
