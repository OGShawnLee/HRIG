<script setup lang="ts">
import { AlertCircle, AlertTriangle, Home } from "lucide-vue-next"
import { RouterLink, useRoute } from "vue-router"

const route = useRoute()
</script>

<template>
  <nav
    class="bg-ground border-b border-solid border-lineDim px-4 py-2 flex flex-wrap items-center gap-2 text-xs text-main"
    aria-label="Window navigation"
  >
    <RouterLink
      :to="{ name: 'home' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'home' }"
      aria-label="Home (Crew Overview)"
    >
      <Home class="w-3.5 h-3.5" aria-hidden="true" />
      <span>Home</span>
    </RouterLink>

    <RouterLink
      :to="{ name: 'dashboard-principal' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'dashboard-principal' }"
    >
      <span>Primary Vitals</span>
    </RouterLink>

    <RouterLink
      :to="{ name: 'support-alert' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'support-alert' }"
    >
      <AlertTriangle class="w-3.5 h-3.5 text-warn" aria-hidden="true" />
      <span>Support</span>
    </RouterLink>

    <RouterLink
      :to="{ name: 'warning-alert' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'warning-alert' }"
    >
      <AlertCircle class="w-3.5 h-3.5 text-crit" aria-hidden="true" />
      <span>Warning</span>
    </RouterLink>
  </nav>
</template>
