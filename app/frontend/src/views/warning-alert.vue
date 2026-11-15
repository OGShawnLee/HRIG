<script setup lang="ts">
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const { goToCrewOverview } = useTelemetryNavigation()
</script>

<template>
  <main class="view3" aria-label="Warning view">
    <div class="top-action-bar">
      <button type="button" class="back-btn" @click="goToCrewOverview">← Crew Overview</button>
    </div>

    <section class="warning-card">
      <div class="warning-copy">
        <h1><span class="warning-badge" aria-hidden="true" />WARNING</h1>
        <p>Pressure: 78 kPa (Drop below 90 kPa)</p>
        <p>
          Associated Risk: Decompression Sickness<br />
          (DCS)
        </p>
      </div>

      <svg class="warning-arrow" viewBox="0 0 100 100" aria-hidden="true">
        <line x1="16" y1="16" x2="78" y2="78" />
        <path d="M54 78H78V54" />
      </svg>
    </section>

    <h2>REQUIRED ACTIONS</h2>

    <section class="actions-card">
      <h3>IMMEDIATE EMERGENCY PROCEDURES</h3>
      <ol>
        <li><strong>INITIATE 100% O2 IN-SUIT PURGE</strong> - Purge nitrogen immediately to prevent gas bubbles</li>
        <li><strong>INCREASE CABIN / SUIT PRESSURE</strong> - Initiate repressurization or don pressure suits</li>
        <li><strong>DOWNLINK &amp; MONITOR</strong> - Alert Crew Medical Officer (CMO) and flight surgeon.</li>
      </ol>
    </section>
  </main>
</template>

<style scoped>
.view3 {
  box-sizing: border-box;
  width: 100%;
  min-height: 100dvh;
  padding: clamp(20px, 7.23vw, 37px) clamp(20px, 9.18vw, 47px);
  overflow: auto;
  color: #f5f5f5;
  background:
    radial-gradient(circle at 48% 61%, rgba(5, 22, 42, 0.62), transparent 39%),
    linear-gradient(115deg, #090909 0%, #000 63%, #020202 100%);
  font-family: Arial, Helvetica, sans-serif;
}

.top-action-bar {
  display: flex;
  align-items: center;
  margin-bottom: 14px;
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

.warning-card,
.actions-card {
  border: 1px solid #5b5b5b;
  background: rgba(0, 0, 0, 0.35);
}

.warning-card {
  position: relative;
  width: 100%;
  min-height: clamp(111px, 21.7vw, 150px);
  padding: 0 clamp(76px, 20.7vw, 106px) 0 clamp(7px, 1.4vw, 12px);
}

.warning-copy h1 {
  display: flex;
  align-items: center;
  gap: 3px;
  margin: 1px 0 5px;
  color: #f00000;
  font-size: clamp(19px, 4.5vw, 23px);
  line-height: 1;
  font-weight: 800;
}

.warning-badge {
  width: 17px;
  height: 17px;
  display: inline-block;
  background: #f00000;
  clip-path: polygon(25% 0, 75% 0, 100% 50%, 75% 100%, 25% 100%, 0 50%);
}

.warning-copy p {
  margin: 0 0 17px;
  font-size: clamp(12px, 2.75vw, 14px);
  line-height: 1.2;
  font-weight: 700;
}

.warning-copy p:last-child {
  margin-bottom: 0;
}

.warning-arrow {
  position: absolute;
  top: 10%;
  right: 3%;
  width: clamp(64px, 17vw, 88px);
  height: auto;
  overflow: visible;
}

.warning-arrow line,
.warning-arrow path {
  fill: none;
  stroke: #f00000;
  stroke-linecap: square;
  stroke-linejoin: miter;
  stroke-width: 16;
}

.view3 h2 {
  margin: 19px 0 7px 7px;
  color: #a9a9a9;
  font-size: clamp(12px, 2.75vw, 14px);
  line-height: 1;
  font-weight: 800;
}

.actions-card {
  width: 100%;
  padding: 6px 9px 15px;
}

.actions-card h3 {
  margin: 0 0 5px;
  font-size: clamp(12px, 2.75vw, 14px);
  line-height: 1.1;
  font-weight: 800;
}

.actions-card ol {
  margin: 0;
  padding-left: 20px;
  font-size: clamp(8px, 1.76vw, 9px);
  line-height: 1.65;
}

.actions-card li {
  padding-left: 1px;
}

@media (max-width: 560px) {
  .warning-card {
    padding-right: 75px;
  }
}
</style>
