<script setup lang="ts">
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const { goToCrewOverview } = useTelemetryNavigation()
</script>

<template>
  <main class="support-view" aria-label="Caution low blood pressure alert view">
    <div class="top-action-bar">
      <button type="button" class="back-btn" @click="goToCrewOverview">← Crew Overview</button>
    </div>

    <section class="support-card">
      <header class="support-header">
        <div class="support-title-group">
          <svg class="caution-icon" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 2.5L22.5 20.5H1.5L12 2.5Z" fill="#f5c518" />
          </svg>
          <h1>SUPPORT</h1>
        </div>
        <span class="caution-badge">CAUTION</span>
      </header>

      <h2 class="condition-heading">LOW BLOOD PRESSURE DETECTED</h2>

      <p class="reading-line">
        Current Reading: 86/55 mmHg <em>(Nominal Range: 90–120 / 60–80 mmHg)</em>
      </p>

      <p class="metadata-line">
        Source: Blood pressure cuff (baumanometro_01) · Demo sample · MET 000:00:00 · 2026-11-15T21:42:00Z (UTC)
      </p>

      <p class="instruction-lead">
        <strong>Your blood pressure is below normal. Follow the option below that matches your current condition</strong>
      </p>
    </section>

    <section class="actions-card">
      <h2>RECOMMENDED ACTIONS</h2>
      <!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->
      <ol>
        <li>
          <strong>If you feel dizzy, lightheaded, or have tunnel vision</strong> → Sit down or strap into a restraint immediately, drink 250ml of water with electrolytes, and remain seated for 10 minutes before re-measuring
        </li>
        <li>
          <strong>If you just transitioned to Lunar Gravity or stood up quickly</strong> → Review the “Orthostatic Intolerance Countermeasures”
        </li>
        <li>
          <strong>If you just finished intense exercise or an EVA</strong> → Check your Body Temperature and Heart Rate for dehydration signs.
        </li>
      </ol>
    </section>
  </main>
</template>

<style scoped>
.support-view {
  box-sizing: border-box;
  width: 100%;
  min-height: 100dvh;
  padding: clamp(20px, 5vw, 37px) clamp(16px, 6vw, 47px);
  display: flex;
  flex-direction: column;
  gap: 18px;
  overflow-x: hidden;
  color: #f5f5f5;
  background:
    radial-gradient(circle at 48% 61%, rgba(5, 22, 42, 0.62), transparent 39%),
    linear-gradient(115deg, #090909 0%, #000 63%, #020202 100%);
  font-family: Arial, Helvetica, sans-serif;
}

.top-action-bar {
  display: flex;
  align-items: center;
}

.back-btn {
  border: 1px solid #8a8a8a;
  background: #0a0a0a;
  color: #f5f5f5;
  min-height: 36px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}

.back-btn:hover {
  background: #141920;
}

.back-btn:focus-visible {
  outline: 2px solid #00a7ff;
  outline-offset: 2px;
}

.support-card,
.actions-card {
  box-sizing: border-box;
  width: 100%;
  padding: clamp(16px, 3.2vw, 24px);
  border: 1px solid #5b5b5b;
  background: rgba(0, 0, 0, 0.45);
}

.support-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
}

.support-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.caution-icon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
}

.support-header h1 {
  margin: 0;
  color: #f5c518;
  font-size: clamp(20px, 4.5vw, 26px);
  line-height: 1.1;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.caution-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border: 1px solid #f5c518;
  background: rgba(0, 0, 0, 0.45);
  color: #f5c518;
  font-size: clamp(12px, 2.5vw, 13px);
  line-height: 1.2;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.condition-heading {
  margin: 0 0 10px;
  color: #f5f5f5;
  font-size: clamp(15px, 3.2vw, 18px);
  line-height: 1.25;
  font-weight: 800;
}

.reading-line {
  margin: 0 0 8px;
  color: #f5f5f5;
  font-size: clamp(14px, 2.8vw, 16px);
  line-height: 1.4;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.reading-line em {
  color: #d4d4d4;
  font-style: italic;
  font-weight: 400;
}

.metadata-line {
  margin: 0 0 14px;
  color: #d4d4d4;
  font-size: clamp(12px, 2.4vw, 13px);
  line-height: 1.4;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.instruction-lead {
  margin: 0;
  color: #f5f5f5;
  font-size: clamp(14px, 2.8vw, 15px);
  line-height: 1.45;
  font-weight: 700;
}

.actions-card h2 {
  margin: 0 0 12px;
  color: #f5f5f5;
  font-size: clamp(14px, 2.8vw, 16px);
  line-height: 1.2;
  font-weight: 800;
  letter-spacing: 0.03em;
}

.actions-card ol {
  margin: 0;
  padding-left: 22px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  color: #f5f5f5;
  font-size: clamp(13px, 2.6vw, 15px);
  line-height: 1.55;
}

.actions-card li {
  padding-left: 4px;
}

@media (max-width: 390px) {
  .support-view {
    padding: 16px 12px;
  }

  .support-card,
  .actions-card {
    padding: 14px 12px;
  }
}
</style>
