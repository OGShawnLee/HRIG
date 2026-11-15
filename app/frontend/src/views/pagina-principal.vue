<template>
  <main class="main-page" aria-label="HRIG main page">
    <div class="search-box" role="search">
      <span class="search-icon" aria-hidden="true">⌕</span>
      <span>Search</span>
    </div>

    <section class="crew-grid" aria-label="Crew members">
      <article class="crew-card">
        <div class="crew-heading">
          <div class="crew-portrait portrait-sam" aria-hidden="true">
            <span class="flag-stripe flag-blue" /><span class="flag-stripe flag-red" /><span class="flag-stripe flag-blue" />
          </div>
          <h2>Sam<br />Sonia</h2>
        </div>
        <dl class="vitals">
          <div><dt>Oxygenation:</dt><dd>98%</dd></div>
          <div><dt>Blood Pressure:</dt><dd>120/80 mm/Hg</dd></div>
          <div><dt>Rad. Exposure:</dt><dd>100 mSv</dd></div>
          <div><dt>Acc. Radiation:</dt><dd>200 mSv</dd></div>
          <div><dt>Heart Beat:</dt><dd>80 BPM</dd></div>
          <div><dt>Temperature:</dt><dd>36° C</dd></div>
        </dl>
      </article>

      <article class="crew-card">
        <div class="crew-heading">
          <div class="crew-portrait portrait-patrick" aria-hidden="true">
            <span class="flag-stripe flag-blue" /><span class="flag-stripe flag-red" /><span class="flag-stripe flag-blue" />
          </div>
          <h2>Patrick<br />Cuckburn</h2>
        </div>
        <dl class="vitals">
          <div><dt>Oxygenation:</dt><dd>98%</dd></div>
          <div><dt>Blood Pressure:</dt><dd>70/80 mm/Hg</dd></div>
          <div><dt>Rad. Exposure:</dt><dd>100 mSv</dd></div>
          <div><dt>Acc. Radiation:</dt><dd>200 mSv</dd></div>
          <div><dt>Heart Beat:</dt><dd>80 BPM</dd></div>
          <div><dt>Temperature:</dt><dd>3° C</dd></div>
        </dl>
      </article>
    </section>

    <section class="bottom-grid">
      <article class="manual-card">
        <svg viewBox="0 0 80 80" aria-hidden="true">
          <path d="M16 17c12-5 21-3 24 2 3-5 12-7 24-2v42c-12-5-21-3-24 2-3-5-12-7-24-2V17Z" />
          <path d="M40 19v42M21 22c7-2 13-1 17 2M59 22c-7-2-13-1-17 2M21 53c7-2 13-1 17 2M59 53c-7-2-13-1-17 2" />
          <path class="medical-star" d="M40 29v14M33 36h14" />
        </svg>
        <span>Medical<br />Manual</span>
      </article>

      <article class="environment-card">
        <h2>Environment</h2>
        <dl class="environment-values">
          <div><dt>O₂:</dt><dd>98%</dd></div>
          <div><dt>Cabin Pressure:</dt><dd>101.3 kPa</dd></div>
          <div><dt>Radiation Exposure:</dt><dd>400 mSv</dd></div>
          <div><dt>Temperature:</dt><dd>22° C</dd></div>
        </dl>
      </article>
    </section>
  </main>
</template>

<style scoped>
.main-page {
  box-sizing: border-box;
  min-height: 100dvh;
  padding: clamp(28px, 7vw, 78px) clamp(18px, 7.5vw, 106px);
  color: #ededed;
  background: #000;
  border: 2px solid #566985;
  font-family: Arial, Helvetica, sans-serif;
}

.search-box {
  width: min(100%, 355px);
  height: 38px;
  margin: 0 auto clamp(22px, 4.4vw, 44px);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border: 1px solid #d9d9d9;
  border-radius: 5px;
  color: #aaa;
  font-size: 13px;
}

.search-icon {
  position: absolute;
  margin-left: -290px;
  color: #00a7ff;
  font-size: 31px;
  line-height: 1;
  transform: rotate(-20deg);
}

.crew-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: clamp(26px, 8vw, 80px);
  max-width: 484px;
  margin: 0 auto;
}

.crew-card {
  min-width: 0;
  min-height: 220px;
  padding: 17px 10px 12px;
  border: 2px solid #ddd;
  border-radius: 10px;
  background: #050505;
}

.crew-heading {
  display: flex;
  align-items: center;
  gap: clamp(12px, 3vw, 24px);
  margin-bottom: 13px;
}

.crew-heading h2 {
  margin: 0;
  font-size: clamp(14px, 2vw, 16px);
  line-height: 1.12;
  font-weight: 400;
}

.crew-portrait {
  position: relative;
  flex: 0 0 59px;
  width: 59px;
  height: 60px;
  overflow: hidden;
  background: linear-gradient(145deg, #b9a283 12%, #54402e 13% 28%, #c4a786 29% 58%, #715134 59%);
  box-shadow: inset 0 0 0 3px rgba(20, 45, 75, .75);
}

.portrait-patrick {
  background: linear-gradient(145deg, #c38b63 14%, #392b22 15% 29%, #b67b56 30% 62%, #4f3527 63%);
}

.flag-stripe {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 18px;
  opacity: .75;
}

.flag-blue { left: 0; background: #173b74; }
.flag-red { left: 18px; background: #b51c24; }
.flag-blue:last-child { left: 36px; }

.vitals,
.environment-values {
  margin: 0;
  font-size: clamp(10px, 1.65vw, 12px);
  line-height: 1.75;
}

.vitals div,
.environment-values div {
  display: flex;
  gap: 5px;
  white-space: nowrap;
}

.vitals dt,
.vitals dd,
.environment-values dt,
.environment-values dd { margin: 0; }

.vitals dt { font-weight: 700; }
.vitals dd,
.environment-values dd { color: #63ff00; }

.bottom-grid {
  display: grid;
  grid-template-columns: 98px minmax(0, 196px);
  align-items: start;
  justify-content: center;
  gap: clamp(70px, 14vw, 124px);
  margin-top: clamp(42px, 7vw, 64px);
}

.manual-card {
  min-height: 96px;
  padding: 9px 7px 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  border: 2px solid #ddd;
  border-radius: 7px;
  font-size: 12px;
  line-height: 1.1;
  text-align: center;
}

.manual-card svg {
  width: 52px;
  height: 52px;
  fill: #eaf6ff;
  stroke: #008fd2;
  stroke-width: 2;
}

.manual-card .medical-star {
  stroke: #117bcc;
  stroke-width: 5;
}

.environment-card {
  min-height: 110px;
  padding: 9px 13px;
  border: 2px solid #ddd;
  border-radius: 7px;
}

.environment-card h2 {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 400;
}

.environment-values {
  font-size: 11px;
  line-height: 1.75;
}

.environment-values dt { font-weight: 700; }

@media (max-width: 620px) {
  .main-page { padding-inline: 12px; }
  .crew-grid { gap: 12px; }
  .crew-card { padding-inline: 8px; }
  .crew-heading { gap: 8px; }
  .crew-portrait { flex-basis: 48px; width: 48px; height: 52px; }
  .flag-stripe { width: 15px; }
  .flag-red { left: 15px; }
  .flag-blue:last-child { left: 30px; }
  .bottom-grid { grid-template-columns: 90px minmax(0, 1fr); gap: 20px; }
}

@media (max-width: 420px) {
  .crew-grid { grid-template-columns: 1fr; max-width: 260px; }
  .bottom-grid { grid-template-columns: 90px minmax(0, 1fr); }
  .vitals, .environment-values { font-size: 11px; }
}
</style>
