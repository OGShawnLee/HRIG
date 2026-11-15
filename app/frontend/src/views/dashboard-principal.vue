<script setup lang="ts">
import { watch } from "vue"
import { useRoute } from "vue-router"
import { useTelemetryNavigation } from "@/composables/useTelemetryNavigation"

const route = useRoute()
const { selectedAstronautName, setSelectedAstronaut, goToCrewOverview } =
  useTelemetryNavigation()

watch(
  () => route.query.astronaut,
  (astronaut) => {
    if (typeof astronaut === "string") {
      setSelectedAstronaut(astronaut)
    }
  },
  { immediate: true },
)
</script>

<template>
  <main class="dashboard" aria-label="Dashboard principal">
    <header class="dashboard-header">
      <button type="button" class="back-btn" @click="goToCrewOverview">← Crew Overview</button>
      <h1>PRIMARY VITALS</h1>
      <div class="search-box">
        <span class="search-icon" aria-hidden="true">⌕</span>
        <span>Search</span>
      </div>
    </header>

    <div class="dashboard-layout">
      <section class="metrics" aria-label="Primary vital signs">
        <article class="metric-row">
          <div class="metric-visual">
            <h2>CARDIAC RHYTHM</h2>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" aria-hidden="true">
              <path d="M0 38h18l7-2 5 3 6-4 6 2 8-1 4-3 5 13 7-46 7 37 8 1 4-7 5 7 9 0 5-2 5 1 5-4 5 3 5-3 5 4 8-1 5-2 5 3 6-3 5 4 7-2 6 2 5-4 5 3 7-1 8 2 5-4 5 2 7-3 8 4 8-2 5 2 5-5 5 4 7-2 7 2 6-2 7 2 8-2 8 2 6-3 6 3 6-2 7 2 5-3 8 2 6-2 10 2" />
            </svg>
          </div>
          <div class="metric-value"><strong>72</strong><span>bpm</span><small>HRV 25 ms</small></div>
          <span class="help">?</span>
        </article>

        <article class="metric-row">
          <div class="metric-visual">
            <h2>OXYGENATION</h2>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" aria-hidden="true">
              <path class="cyan-line" d="M0 31c12 4 16 20 24 18 8-2 8-48 18-43 10 5 10 38 22 37 12-1 10-20 19-20 9 0 13 27 24 25 11-2 11-40 21-38 10 2 11 35 22 34 11-1 10-24 20-24 10 0 13 29 24 28 11-1 10-40 20-38 10 2 11 36 22 34 11-2 10-25 20-24 10 1 13 27 24 26 11-1 11-40 20-38 9 2 10 36 21 34 11-2 11-24 20-23 9 1 13 23 24 22" />
            </svg>
          </div>
          <div class="metric-value"><strong>98%</strong><span>SpO2</span><small>02: 20.9%</small></div>
          <span class="help">?</span>
        </article>

        <article class="metric-row">
          <div class="metric-visual">
            <h2 class="green-label">PRESSURE</h2>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" aria-hidden="true">
              <path class="white-line" d="M0 54c8-1 12-39 20-38 8 1 10 17 18 14 8-3 10-8 17-5 7 3 8 5 15 3 7-2 8-26 15-28 7-2 8 44 17 43 9-1 10-22 18-21 8 1 10 12 18 9 8-3 8-35 16-35 8 0 10 46 19 45 9-1 10-26 18-24 8 2 11 8 18 4 7-4 8-27 16-27 8 0 10 41 18 40 8-1 11-14 19-13 8 1 11 6 18 2 7-4 8-26 16-26 8 0 11 37 19 37 8 0 12-12 20-11 8 1 11 7 20 2" />
            </svg>
          </div>
          <div class="metric-value"><span class="unit-top">BP <em>mmHg</em></span><strong>118/77</strong><small>CABIN: 101.3 kPa</small></div>
          <span class="help">?</span>
        </article>

        <article class="metric-row">
          <div class="metric-visual">
            <h2 class="green-label">RADIATION <b>SENSOR</b></h2>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" aria-hidden="true">
              <path class="yellow-line" d="M0 11L29 38 29 62 55 62 55 13 84 39 84 62 110 62 110 12 139 41 139 62 166 62 166 12 194 39 194 62 220 62 220 12 249 39 249 62 275 62 300 38" />
              <path class="yellow-fill" d="M0 11L29 38 29 62 55 62 55 13 84 39 84 62 110 62 110 12 139 41 139 62 166 62 166 12 194 39 194 62 220 62 220 12 249 39 249 62 275 62 300 38V64H0Z" />
            </svg>
          </div>
          <div class="metric-value"><strong>32</strong><span>uSv/h</span><small>TOTAL 1.2 mSv</small></div>
          <span class="help">?</span>
        </article>

        <article class="metric-row">
          <div class="metric-visual">
            <h2 class="green-label">TEMPERATURE <b>BODY</b></h2>
            <svg viewBox="0 0 300 64" preserveAspectRatio="none" aria-hidden="true"><path class="white-line" d="M0 60H300" /></svg>
          </div>
          <div class="metric-value"><strong>36.6</strong><span>°C</span><small>CABIN 22.0°C</small></div>
          <span class="help">?</span>
        </article>
      </section>

      <aside class="astronaut-panel">
        <svg class="astronaut" viewBox="0 0 160 360" aria-label="Astronaut silhouette">
          <g class="astronaut-fill">
            <path d="M68 12h30l10 10 4 29-10 22-8 8 19 13 16 8 4 57 13 29-9 20-11-6-5-26-8-23 4 55-4 41-12 68-11 4-5-7 4-64-7-34-7 34 4 64-6 7-11-4-12-68-4-41 5-55-9 23-5 26-11 6-9-20 13-29 4-57 16-8 19-13-8-8-10-22 4-29 10-10Z" />
            <path d="M68 12h30l10 10-2 24-7 25-21 10-21-10-7-25-2-24 10-10Z" />
          </g>
          <g class="astronaut-lines">
            <path d="M60 81l-15 18-8 58-17 29 8 20 13-6 7-27 11-24M100 81l15 18 8 58 17 29-8 20-13-6-7-27-11-24M51 181l8 54-4 41 12 68M109 181l-8 54 4 41-12 68M53 98h54M50 154h60M53 205h54M58 236h44" />
          </g>
        </svg>
        <h2>{{ selectedAstronautName }}</h2>
        <div class="legend"><span><i class="caution" />Caution</span><span><i class="warning" />Warning</span></div>
      </aside>
    </div>
  </main>
</template>

<style scoped>
.dashboard {
  box-sizing: border-box;
  min-height: 100dvh;
  padding: 20px clamp(14px, 4vw, 41px) 26px;
  overflow: auto;
  color: #eee;
  background: #000;
  border: 2px solid #5c6d89;
  font-family: Arial, Helvetica, sans-serif;
}

.dashboard-header {
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 650px;
  margin: 0 auto 16px;
}

.dashboard-header h1 {
  margin: 0;
  flex: 0 0 auto;
  font-size: clamp(17px, 3vw, 20px);
}

.back-btn {
  flex: 0 0 auto;
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

.search-box {
  position: relative;
  flex: 1;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #eee;
  border-radius: 5px;
  color: #aaa;
  font-size: 13px;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: #00a7ff;
  font-size: 30px;
  line-height: 1;
  transform: rotate(-20deg);
}

.dashboard-layout {
  display: grid;
  grid-template-columns: minmax(0, 426px) minmax(150px, 1fr);
  gap: clamp(22px, 7vw, 76px);
  max-width: 650px;
  margin: 0 auto;
}

.metrics { min-width: 0; }

.metric-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 124px 20px;
  min-height: 75px;
  border: 1px solid #777;
  background: #020202;
}

.metric-visual {
  min-width: 0;
  overflow: hidden;
  border-right: 1px solid #777;
  background: #03080b;
}

.metric-visual h2 {
  height: 18px;
  margin: 0;
  padding: 3px 11px 0;
  color: #eee;
  font-size: 11px;
  font-weight: 800;
}

.metric-visual h2.green-label { color: #378c2a; }
.metric-visual h2 b { float: right; color: #eee; }
.metric-visual svg { display: block; width: 100%; height: 57px; padding: 0 4px; }
.metric-visual path { fill: none; stroke: #62ad3d; stroke-width: 2; }
.metric-visual path.cyan-line { stroke: #36b3b7; }
.metric-visual path.white-line { stroke: #eee; }
.metric-visual path.yellow-line { stroke: #c9bc60; }
.metric-visual path.yellow-fill { fill: rgba(188, 167, 45, .28); stroke: none; }

.metric-value {
  position: relative;
  padding: 16px 7px 4px;
  color: #438a32;
}

.metric-value strong {
  display: block;
  font-size: clamp(23px, 4vw, 29px);
  line-height: 1;
  font-weight: 800;
}

.metric-value span { position: absolute; top: 8px; right: 7px; font-size: 11px; font-weight: 700; }
.metric-value span.unit-top { top: 8px; left: 7px; right: 7px; }
.metric-value span em { float: right; font-style: normal; }
.metric-value small { display: block; margin-top: 7px; font-size: 11px; font-weight: 700; white-space: nowrap; }
.help { align-self: end; padding: 0 5px 3px 0; font-size: 11px; }

.astronaut-panel { position: relative; min-width: 0; text-align: center; }
.astronaut { width: min(100%, 155px); height: auto; margin-top: 0; }
.astronaut-fill { fill: #70a9e8; stroke: #0064c9; stroke-width: 3; }
.astronaut-lines { fill: none; stroke: #075ab3; stroke-width: 2; }
.astronaut-panel h2 { margin: 1px 0 0; font-size: 11px; }
.legend { margin: 5px auto 0; width: 66px; padding: 3px; border: 1px solid #aaa; text-align: left; font-size: 9px; line-height: 1.25; }
.legend span { display: block; white-space: nowrap; }
.legend i { display: inline-block; width: 0; height: 0; margin-right: 3px; border-left: 5px solid transparent; border-right: 5px solid transparent; border-bottom: 8px solid #fff000; }
.legend i.warning { border-bottom-color: #f00000; }

@media (max-width: 660px) {
  .dashboard-header { flex-direction: column; align-items: stretch; }
  .dashboard-header h1 { text-align: center; }
  .dashboard-layout { grid-template-columns: 1fr; }
  .astronaut-panel { order: -1; }
  .astronaut { width: 120px; }
  .legend { display: inline-block; margin-left: 12px; vertical-align: top; }
}

@media (max-width: 430px) {
  .metric-row { grid-template-columns: minmax(0, 1fr) 103px 16px; }
  .metric-value small { font-size: 9px; }
  .metric-value span { font-size: 9px; }
}
</style>
