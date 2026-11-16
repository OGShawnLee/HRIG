<script setup lang="ts">
import { Home, Search } from "lucide-vue-next"
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
      :to="{ name: 'search' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'search' }"
      aria-label="Search crew, medical procedures, hazards, and metrics (Ctrl+K)"
    >
      <Search class="w-3.5 h-3.5" aria-hidden="true" />
      <span>Search</span>
      <kbd
        class="hidden sm:inline-flex items-center px-1.5 py-0.5 text-[10px] font-medium text-dull border border-solid border-lineDim rounded"
      >
        Ctrl+K
      </kbd>
    </RouterLink>

    <RouterLink
      :to="{ name: 'dashboard-principal' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'dashboard-principal' }"
    >
      <span>Primary Vitals</span>
    </RouterLink>

    <RouterLink
      :to="{ name: 'secondary-stats' }"
      class="button-side h-9 px-3 text-xs gap-1.5"
      :class="{ 'bg-surface border-line text-main ring-1 ring-line': route.name === 'secondary-stats' }"
    >
      <span>Secondary Stats</span>
    </RouterLink>
  </nav>
</template>
