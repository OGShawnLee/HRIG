<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router"

const route = useRoute()
</script>

<template>
  <nav class="view-nav-bar" aria-label="Window navigation">
    <RouterLink
      :to="{ name: 'pagina-principal' }"
      class="nav-link"
      :class="{ 'is-active': route.name === 'pagina-principal' }"
      aria-label="Home (Crew Monitor)"
    >
      <svg
        class="nav-icon"
        viewBox="0 0 24 24"
        width="14"
        height="14"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M3 10.5L12 3l9 7.5" />
        <path d="M5 9.5V20h5v-6h4v6h5V9.5" />
      </svg>
      <span>Home</span>
    </RouterLink>

    <RouterLink
      :to="{ name: 'dashboard-principal' }"
      class="nav-link"
      :class="{ 'is-active': route.name === 'dashboard-principal' }"
    >
      Primary Vitals
    </RouterLink>

    <RouterLink
      :to="{ name: 'support-alert' }"
      class="nav-link"
      :class="{ 'is-active': route.name === 'support-alert' }"
    >
      Support
    </RouterLink>

    <RouterLink
      :to="{ name: 'warning-alert' }"
      class="nav-link"
      :class="{ 'is-active': route.name === 'warning-alert' }"
    >
      Warning
    </RouterLink>

    <RouterLink
      :to="{ name: 'home' }"
      class="nav-link nav-link-bridge"
      :class="{ 'is-active': route.name === 'home' }"
    >
      Bridge
    </RouterLink>
  </nav>
</template>

<style scoped>
.view-nav-bar {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  padding: 6px 12px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  background: #080b10;
  border-bottom: 1px solid #475569;
  color: #f5f5f5;
  font-family: Arial, Helvetica, sans-serif;
  font-size: 12px;
}

.nav-link {
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 32px;
  padding: 4px 10px;
  border: 1px solid #64748b;
  background: #0d121a;
  color: #f5f5f5;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.2;
  text-decoration: none;
  white-space: nowrap;
}

.nav-link:hover {
  background: #151d28;
}

.nav-link:focus-visible {
  outline: 2px solid #00a7ff;
  outline-offset: 2px;
}

.nav-link.is-active {
  border-color: #00a7ff;
  color: #38bdf8;
  background: #0f1b2d;
}

.nav-icon {
  flex-shrink: 0;
}

@media (min-width: 560px) {
  .nav-link-bridge {
    margin-left: auto;
  }
}

@media (max-width: 390px) {
  .view-nav-bar {
    padding: 6px 8px;
    gap: 6px;
  }

  .nav-link {
    padding: 4px 8px;
    font-size: 11px;
  }
}
</style>
