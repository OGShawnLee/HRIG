<script setup lang="ts">
import { computed } from "vue"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"
import type { ReadingStatus } from "@/composables/useTelemetryNavigation"

const { lastStatus, goToCrewOverview, dispatchTelemetryStatus, resetDemoStatus } =
  useTelemetryNavigation()

const STATUS_META: Record<
  ReadingStatus,
  { icon: string; label: string; toneClass: string }
> = {
  normal: { icon: "✓", label: "NOMINAL", toneClass: "status-normal" },
  abnormal: { icon: "▲", label: "CAUTION / ABNORMAL", toneClass: "status-abnormal" },
  critical: { icon: "⬣", label: "CRITICAL", toneClass: "status-critical" },
  stale: { icon: "⊘", label: "STALE", toneClass: "status-stale" },
}

const currentStatusMeta = computed(() => STATUS_META[lastStatus.value])
</script>

<template>
  <aside
    class="demo-alert-strip"
    role="region"
    aria-label="Mockup navigation and alert simulator"
  >
    <div class="strip-status" :class="currentStatusMeta.toneClass" aria-live="polite">
      <span class="status-icon" aria-hidden="true">{{ currentStatusMeta.icon }}</span>
      <span class="status-label">{{ currentStatusMeta.label }}</span>
    </div>

    <div class="strip-actions">
      <button type="button" class="strip-btn strip-btn-neutral" @click="goToCrewOverview()">
        Crew Overview
      </button>
      <button
        type="button"
        class="strip-btn strip-btn-abnormal"
        @click="
          dispatchTelemetryStatus({
            status: 'abnormal',
            metricCode: 'bp_systolic',
            source: 'mockup-demo',
          })
        "
      >
        ▲ Simulate Abnormal
      </button>
      <button
        type="button"
        class="strip-btn strip-btn-critical"
        @click="
          dispatchTelemetryStatus({
            status: 'critical',
            metricCode: 'cabin_pressure',
            source: 'mockup-demo',
          })
        "
      >
        ⬣ Simulate Critical
      </button>
      <button type="button" class="strip-btn strip-btn-neutral" @click="resetDemoStatus()">
        Reset
      </button>
    </div>
  </aside>
</template>

<style scoped>
.demo-alert-strip {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  padding: 8px 12px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  background: #080b10;
  border-top: 1px solid #475569;
  color: #f5f5f5;
  font-family: Arial, Helvetica, sans-serif;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}

.strip-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 28px;
  padding: 2px 10px;
  border: 1px solid #475569;
  background: #0d121a;
  font-weight: 800;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.status-normal {
  border-color: #2dd4bf;
  color: #2dd4bf;
}

.status-abnormal {
  border-color: #f5c518;
  color: #f5c518;
}

.status-critical {
  border-color: #ff4d4d;
  color: #ff6b6b;
}

.status-stale {
  border-style: dashed;
  border-color: #94a3b8;
  color: #cbd5e1;
}

.strip-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.strip-btn {
  box-sizing: border-box;
  min-height: 32px;
  padding: 4px 10px;
  background: #0d121a;
  border: 1px solid #8a8a8a;
  color: #f5f5f5;
  font-family: inherit;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.2;
  cursor: pointer;
  white-space: nowrap;
}

.strip-btn:hover {
  background: #151d28;
}

.strip-btn:focus-visible {
  outline: 2px solid #00a7ff;
  outline-offset: 2px;
}

.strip-btn-abnormal {
  border-color: #f5c518;
  color: #f5c518;
}

.strip-btn-critical {
  border-color: #ff4d4d;
  color: #ff4d4d;
}

@media (max-width: 390px) {
  .demo-alert-strip {
    padding: 8px;
    gap: 6px;
  }

  .strip-btn {
    padding: 4px 8px;
    font-size: 11px;
  }
}
</style>
