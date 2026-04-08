<script>
  import { api } from '../api.js'
  import { t } from '../i18n.js'
  import { onMount, onDestroy } from 'svelte'

  export let stats = null
  export let screenshot = null

  let screenshotTs = Date.now()
  let refreshInterval = 3
  let timer = null
  let downloading = false

  onMount(() => { startPolling() })
  onDestroy(() => stopPolling())

  function startPolling() {
    stopPolling()
    timer = setInterval(() => { screenshotTs = Date.now() }, refreshInterval * 1000)
  }
  function stopPolling() { clearInterval(timer) }

  async function downloadScreenshot() {
    downloading = true
    const a = document.createElement('a')
    a.href = api.screenshotUrl()
    a.download = `miyoo_${Date.now()}.png`
    a.click()
    setTimeout(() => downloading = false, 1000)
  }

  function fmtMb(mb) {
    if (mb >= 1024) return (mb/1024).toFixed(1) + ' GB'
    return mb + ' MB'
  }

  $: ramPct = stats ? Math.round((stats.ram.used / stats.ram.total) * 100) : 0
  $: cpuPct = stats ? Math.round(stats.cpu_percent) : 0
  $: batPct = stats?.battery?.percent ?? -1
  $: screenshotSrc = screenshot || api.screenshotUrl() + '&ts=' + screenshotTs
</script>

<div class="dashboard">
  <div class="stats-row">
    <!-- CPU -->
    <div class="stat-card">
      <div class="stat-label">{$t.cpu}</div>
      <div class="stat-value">{cpuPct}<span class="unit">%</span></div>
      <div class="bar"><div class="bar-fill cpu" style="width:{cpuPct}%"></div></div>
      {#if stats}<div class="stat-sub">{stats.cpu_freq_mhz} MHz</div>{/if}
    </div>

    <!-- RAM -->
    <div class="stat-card">
      <div class="stat-label">{$t.ram}</div>
      <div class="stat-value">{stats ? fmtMb(stats.ram.used) : '—'}</div>
      <div class="bar"><div class="bar-fill ram" style="width:{ramPct}%"></div></div>
      {#if stats}<div class="stat-sub">{fmtMb(stats.ram.available)} {$lang === 'fr' ? 'libre' : 'free'}</div>{/if}
    </div>

    <!-- Battery -->
    <div class="stat-card">
      <div class="stat-label">{$t.battery}</div>
      <div class="stat-value">
        {#if stats && batPct >= 0}
          {batPct}<span class="unit">%</span>
          {#if stats.battery.charging}<span class="green">⚡</span>{/if}
        {:else}—{/if}
      </div>
      {#if stats && batPct >= 0}
        <div class="bar">
          <div class="bar-fill bat" class:low={batPct < 20} style="width:{batPct}%"></div>
        </div>
        <div class="stat-sub">{stats.battery.voltage}</div>
      {/if}
    </div>

    <!-- Network -->
    <div class="stat-card">
      <div class="stat-label">{$t.network}</div>
      <div class="stat-value ip">{stats ? stats.ip : '—'}</div>
      {#if stats}<div class="stat-sub">↑ {stats.uptime}</div>{/if}
    </div>
  </div>

  <!-- Live screen -->
  <div class="screen-section">
    <div class="section-header">
      <h2>{$t.liveScreen}</h2>
      <div class="controls">
        <label class="select-wrap">
          {$t.refresh}
          <select bind:value={refreshInterval} on:change={startPolling}>
            <option value={1}>1s</option>
            <option value={2}>2s</option>
            <option value={3}>3s</option>
            <option value={5}>5s</option>
            <option value={10}>10s</option>
          </select>
        </label>
        <button class="btn-sm" on:click={downloadScreenshot} disabled={downloading}>
          {$t.savePng}
        </button>
      </div>
    </div>

    <div class="screen-wrap">
      <img
        src={screenshotSrc}
        alt="Miyoo screen"
        class="screen"
        on:error={(e) => e.target.style.display='none'}
        on:load={(e) => e.target.style.display='block'}
      />
      <div class="screen-fallback">
        <span>{$t.screenUnavailable}</span>
        <small>{$t.startGame}</small>
      </div>
    </div>
  </div>
</div>

<script context="module">
  import { get } from 'svelte/store'
  import { lang } from '../i18n.js'
  // helper to read lang store in non-reactive context
</script>

<style>
  .dashboard { padding: 14px; display: flex; flex-direction: column; gap: 14px; overflow-y: auto; height: 100%; }

  .stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap: 10px; }
  .stat-card {
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 14px; padding: 14px;
    display: flex; flex-direction: column; gap: 3px;
  }
  .stat-label { font-size: 0.68rem; text-transform: uppercase; letter-spacing: 1.5px; color: #444; }
  .stat-value { font-size: 1.5rem; font-weight: 700; color: #e0e0e0; line-height: 1.1; }
  .unit { font-size: 0.85rem; color: #666; }
  .green { color: #6bff9d; font-size: 0.85rem; margin-left: 4px; }
  .stat-sub { font-size: 0.7rem; color: #444; margin-top: 2px; }
  .ip { font-size: 0.9rem !important; font-family: monospace; color: #6bff9d !important; }

  .bar { height: 3px; background: #161616; border-radius: 2px; overflow: hidden; margin: 6px 0 2px; }
  .bar-fill { height: 100%; border-radius: 2px; transition: width 0.6s ease; }
  .bar-fill.cpu { background: linear-gradient(90deg, #e8488a, #ff8ab4); }
  .bar-fill.ram { background: linear-gradient(90deg, #6b9dff, #a0c0ff); }
  .bar-fill.bat { background: linear-gradient(90deg, #6bff9d, #a0ffcc); }
  .bar-fill.bat.low { background: #ff6b6b; }

  .screen-section {
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 14px; overflow: hidden;
    flex: 1;
  }
  .section-header {
    display: flex; align-items: center; justify-content: space-between;
    padding: 10px 14px; border-bottom: 1px solid #111;
  }
  h2 { font-size: 0.72rem; color: #555; text-transform: uppercase; letter-spacing: 1.5px; }
  .controls { display: flex; align-items: center; gap: 10px; }
  .select-wrap { display: flex; align-items: center; gap: 6px; font-size: 0.72rem; color: #444; }
  select {
    background: #111; border: 1px solid #1e1e1e; border-radius: 6px;
    color: #888; padding: 2px 6px; font-size: 0.72rem; cursor: pointer;
  }
  .btn-sm {
    background: #111; border: 1px solid #1e1e1e; border-radius: 6px;
    color: #777; padding: 4px 10px; font-size: 0.72rem; cursor: pointer;
    transition: all 0.15s;
  }
  .btn-sm:hover:not(:disabled) { border-color: #e8488a55; color: #e8488a; }
  .btn-sm:disabled { opacity: 0.4; }

  .screen-wrap {
    position: relative; display: flex; justify-content: center; align-items: center;
    padding: 14px; background: #060606; min-height: 180px;
  }
  .screen {
    max-width: 100%; max-height: 340px;
    border-radius: 4px; image-rendering: pixelated;
    border: 1px solid #1a1a1a;
  }
  .screen-fallback {
    position: absolute; display: flex; flex-direction: column;
    align-items: center; gap: 6px; color: #333; font-size: 0.82rem;
    text-align: center; pointer-events: none;
  }
  .screen:not([style*="none"]) ~ .screen-fallback { display: none; }
  small { color: #2a2a2a; font-size: 0.7rem; }
</style>
