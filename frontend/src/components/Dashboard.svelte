<script>
  import { onMount, onDestroy } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'
  import Controller from './Controller.svelte'

  export let stats = null
  export let screenshot = null  // base64 from WebSocket

  let refreshInterval = 3
  let ts = Date.now()
  let timer = null
  let downloading = false

  // Brightness control (only shown if the device exposes a backlight)
  let bright = 0
  let brightMax = 100
  let brightSupported = false
  let brightTimer = null

  onMount(() => { startPolling(); loadBrightness() })
  onDestroy(stopPolling)

  async function loadBrightness() {
    try {
      const b = await api.brightness()
      brightSupported = !!b.supported
      if (brightSupported) { bright = b.value; brightMax = b.max || 100 }
    } catch (e) { /* ignore */ }
  }

  // Debounce slider writes so we don't spam the device while dragging.
  function onBright() {
    clearTimeout(brightTimer)
    brightTimer = setTimeout(() => api.setBrightness(bright).catch(() => {}), 150)
  }

  function startPolling() {
    stopPolling()
    // Fix #8: only poll if no WS screenshot arriving
    timer = setInterval(() => { ts = Date.now() }, refreshInterval * 1000)
  }
  function stopPolling() { clearInterval(timer) }

  // Live screen: either polled snapshots or a smooth native MJPEG stream.
  let streamMode = false
  $: screenshotSrc = streamMode ? api.streamUrl() : (screenshot || api.screenshotUrl(ts))
  // Reset the load-error flag whenever the source changes so each refresh retries
  let imgError = false
  $: screenshotSrc, (imgError = false)

  // Rolling CPU history for the sparkline (futile but pretty).
  let cpuHistory = []
  $: if (stats) cpuHistory = [...cpuHistory, stats.cpu_percent].slice(-32)
  $: spark = cpuHistory.length > 1
    ? cpuHistory.map((v, i) => `${(i / (cpuHistory.length - 1)) * 100},${100 - Math.min(Math.max(v, 0), 100)}`).join(' ')
    : ''

  // Quit the running game
  let quitting = false
  async function quitGame() {
    if (!confirm($t.quitGameConfirm)) return
    quitting = true
    try { await api.quitGame() } catch (e) { /* ignore */ }
    setTimeout(() => quitting = false, 2000)
  }

  // Server log viewer
  let showLogPanel = false
  let logText = ''
  let logLoading = false
  async function loadLogs() {
    logLoading = true
    try { const r = await api.logs(300); logText = r.log || '(vide)' }
    catch (e) { logText = e.message }
    logLoading = false
  }
  function toggleLogs() {
    showLogPanel = !showLogPanel
    if (showLogPanel) loadLogs()
  }

  let powering = ''
  async function power(action) {
    const msg = action === 'reboot' ? $t.rebootConfirm : $t.poweroffConfirm
    if (!confirm(msg)) return
    powering = action
    try { await api.power(action) } catch (e) { /* device is going down */ }
  }

  async function saveScreenshot() {
    downloading = true
    const a = document.createElement('a')
    a.href = api.screenshotUrl(Date.now())
    a.download = `miyoo_${Date.now()}.png`
    a.click()
    setTimeout(() => downloading = false, 1000)
  }

  function fmtMb(mb) {
    if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
    return mb + ' MB'
  }

  $: cpuPct  = stats ? Math.round(stats.cpu_percent) : 0
  $: ramPct  = (stats && stats.ram.total > 0) ? Math.round((stats.ram.used / stats.ram.total) * 100) : 0
  $: batPct  = stats?.battery?.percent ?? -1
  $: gameOn  = stats?.game_running ?? false
  $: ramOk   = stats && stats.ram.total > 0
  $: batOk   = batPct >= 0
  $: tempC   = stats?.temp_c ?? 0
</script>

<div class="dashboard">
  <div class="stats-row">
    <!-- CPU -->
    <div class="card">
      <div class="label">{$t.cpu}</div>
      <div class="value">{cpuPct}<span class="unit">%</span></div>
      <div class="bar"><div class="fill cpu" style="width:{cpuPct}%"></div></div>
      {#if spark}
        <svg class="spark" viewBox="0 0 100 100" preserveAspectRatio="none">
          <polyline points={spark} />
        </svg>
      {/if}
      {#if stats}<small>{stats.cpu_freq_mhz} MHz</small>{/if}
    </div>

    <!-- RAM -->
    <div class="card">
      <div class="label">{$t.ram}</div>
      <div class="value">{ramOk ? fmtMb(stats.ram.used) : (stats ? '0 MB' : '—')}</div>
      <div class="bar"><div class="fill ram" style="width:{ramPct}%"></div></div>
      {#if ramOk}<small>{fmtMb(stats.ram.available)} free</small>
      {:else if stats}<small class="warn">N/A</small>{/if}
    </div>

    <!-- Battery -->
    <div class="card">
      <div class="label">{$t.battery}</div>
      <div class="value">
        {#if batOk}{batPct}<span class="unit">%</span>
          {#if stats?.battery?.charging}<span class="green">⚡</span>{/if}
        {:else if stats}<span class="warn">N/A</span>
        {:else}—{/if}
      </div>
      {#if batOk}
        <div class="bar"><div class="fill bat" class:low={batPct < 20} style="width:{batPct}%"></div></div>
        {#if stats?.battery?.voltage}<small>{stats.battery.voltage}</small>{/if}
      {/if}
    </div>

    <!-- Network -->
    <div class="card">
      <div class="label">{$t.network}</div>
      <div class="value ip">{stats?.ip ?? '—'}</div>
      {#if stats}<small>↑ {stats.uptime}</small>{/if}
      {#if gameOn}<div class="game-badge">🎮 {$t.gameRunning}</div>{/if}
    </div>

    <!-- Temperature -->
    {#if tempC > 0}
    <div class="card">
      <div class="label">{$t.temp}</div>
      <div class="value">{tempC}<span class="unit">°C</span></div>
      <div class="bar"><div class="fill temp" class:hot={tempC >= 70} style="width:{Math.min(tempC, 100)}%"></div></div>
      <small class:warn={tempC >= 70}>{tempC >= 70 ? '🔥' : 'SoC'}</small>
    </div>
    {/if}
  </div>

  <!-- Brightness -->
  {#if brightSupported}
  <div class="power-row">
    <span class="power-label">☀ {$t.brightness}</span>
    <input class="bright-slider" type="range" min="0" max={brightMax}
      bind:value={bright} on:input={onBright} />
    <span class="bright-val">{Math.round((bright / brightMax) * 100)}%</span>
  </div>
  {/if}

  <!-- Power control -->
  <div class="power-row">
    <span class="power-label">{$t.power}</span>
    <div class="power-btns">
      {#if gameOn}
        <button class="pwr quit" on:click={quitGame} disabled={quitting}>{$t.quitGame}</button>
      {/if}
      <button class="pwr" on:click={toggleLogs} class:active={showLogPanel}>{$t.showLogs}</button>
      <button class="pwr reboot" on:click={() => power('reboot')} disabled={powering !== ''}>
        ↻ {$t.reboot}
      </button>
      <button class="pwr off" on:click={() => power('poweroff')} disabled={powering !== ''}>
        ⏻ {$t.poweroff}
      </button>
    </div>
  </div>

  <!-- Server log viewer -->
  {#if showLogPanel}
  <div class="log-panel">
    <div class="log-head">
      <span class="power-label">{$t.logs}</span>
      <button class="pwr" on:click={loadLogs} disabled={logLoading}>{$t.refreshLogs}</button>
    </div>
    <pre class="log-body">{logLoading ? '…' : logText}</pre>
  </div>
  {/if}

  <!-- Gamepad controller -->
  <Controller />

  <!-- Live screen -->
  <div class="screen-card">
    <div class="card-header">
      <span class="card-title">{$t.liveScreen}</span>
      {#if gameOn}<span class="live-dot"></span>{/if}
      <div class="header-actions">
        <label class="select-row">
          {$t.refresh}
          <select bind:value={refreshInterval} on:change={startPolling}>
            <option value={1}>1s</option>
            <option value={2}>2s</option>
            <option value={3}>3s</option>
            <option value={5}>5s</option>
            <option value={10}>10s</option>
          </select>
        </label>
        <button class="btn-sm" class:active={streamMode} on:click={() => streamMode = !streamMode}>
          🎥 {$t.smoothStream}
        </button>
        <button class="btn-sm" on:click={saveScreenshot} disabled={downloading}>
          {$t.savePng}
        </button>
      </div>
    </div>

    <div class="screen-wrap">
      {#if screenshotSrc && !imgError}
        <img src={screenshotSrc} alt="Miyoo screen" class="screen"
          on:error={() => imgError = true}
        />
      {:else}
        <div class="no-screen">
          <span>{$t.screenUnavailable}</span>
          <small>{$t.startGame}</small>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .dashboard { padding: 14px; display: flex; flex-direction: column; gap: 14px; overflow-y: auto; height: 100%; }

  .stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px,1fr)); gap: 10px; }
  .card {
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 14px; padding: 14px;
    display: flex; flex-direction: column; gap: 3px;
  }
  .label { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 1.5px; color: #3a3a3a; }
  .value { font-size: 1.5rem; font-weight: 700; color: #e0e0e0; line-height: 1.1; }
  .unit { font-size: 0.8rem; color: #555; }
  .green { color: #6bff9d; font-size: 0.8rem; margin-left: 3px; }
  small { font-size: 0.68rem; color: #3a3a3a; margin-top: 2px; }
  .warn { font-size: 0.75rem; color: #555; }
  .ip { font-size: 0.88rem !important; font-family: monospace; color: #6bff9d !important; }
  .game-badge {
    margin-top: 4px; font-size: 0.65rem; color: #e8488a;
    background: #e8488a11; border: 1px solid #e8488a22;
    border-radius: 8px; padding: 2px 7px; align-self: flex-start;
  }

  .bar { height: 3px; background: #111; border-radius: 2px; overflow: hidden; margin: 6px 0 2px; }
  .fill { height: 100%; border-radius: 2px; transition: width .6s ease; }
  .fill.cpu { background: linear-gradient(90deg,#e8488a,#ff8ab4); }
  .fill.ram { background: linear-gradient(90deg,#6b9dff,#a0c0ff); }
  .fill.bat { background: linear-gradient(90deg,#6bff9d,#a0ffcc); }
  .fill.bat.low { background: #ff6b6b; }
  .fill.temp { background: linear-gradient(90deg,#6bff9d,#ffcf6b); }
  .fill.temp.hot { background: linear-gradient(90deg,#ff8a4b,#ff6b6b); }

  .bright-slider { flex: 1; margin: 0 12px; accent-color: #e8488a; height: 4px; cursor: pointer; }
  .bright-val { font-size: 0.72rem; color: #888; min-width: 34px; text-align: right; }

  .power-row {
    display: flex; align-items: center; justify-content: space-between;
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 14px; padding: 10px 14px;
  }
  .power-label { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 1.5px; color: #3a3a3a; }
  .power-btns { display: flex; gap: 8px; }
  .pwr {
    background: #111; border: 1px solid #1a1a1a; border-radius: 8px;
    color: #888; padding: 6px 14px; font-size: 0.78rem; font-weight: 600;
    cursor: pointer; transition: all .15s;
  }
  .pwr:hover:not(:disabled) { transform: translateY(-1px); }
  .pwr.reboot:hover:not(:disabled) { border-color: #6b9dff55; color: #6b9dff; }
  .pwr.off:hover:not(:disabled) { border-color: #ff6b6b55; color: #ff6b6b; }
  .pwr.quit { border-color: #ff6b6b33; color: #ff8a8a; }
  .pwr.quit:hover:not(:disabled) { border-color: #ff6b6b; color: #ff6b6b; }
  .pwr.active { border-color: #e8488a; color: #e8488a; background: #e8488a11; }
  .pwr:disabled { opacity: .4; cursor: default; }

  .log-panel { background: #0a0a0a; border: 1px solid #1a1a1a; border-radius: 14px; overflow: hidden; }
  .log-head { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; border-bottom: 1px solid #111; }
  .log-body {
    margin: 0; padding: 10px 12px; max-height: 220px; overflow: auto;
    font-family: 'Consolas','Monaco',monospace; font-size: 0.68rem; line-height: 1.5;
    color: #8a8a8a; white-space: pre-wrap; word-break: break-word;
  }

  .screen-card { background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 14px; overflow: hidden; flex: 1; }
  .card-header {
    display: flex; align-items: center; gap: 8px;
    padding: 9px 14px; border-bottom: 1px solid #111;
  }
  .card-title { font-size: 0.65rem; text-transform: uppercase; letter-spacing: 1.5px; color: #3a3a3a; }
  .live-dot {
    width: 6px; height: 6px; border-radius: 50%; background: #e8488a;
    animation: blink 1.2s ease-in-out infinite; flex-shrink: 0;
  }
  @keyframes blink { 0%,100%{opacity:1} 50%{opacity:.2} }
  .header-actions { display: flex; align-items: center; gap: 8px; margin-left: auto; }
  .select-row { display: flex; align-items: center; gap: 5px; font-size: 0.68rem; color: #3a3a3a; }
  select {
    background: #111; border: 1px solid #1a1a1a; border-radius: 5px;
    color: #777; padding: 2px 5px; font-size: 0.68rem;
  }
  .btn-sm {
    background: #111; border: 1px solid #1a1a1a; border-radius: 5px;
    color: #666; padding: 3px 9px; font-size: 0.68rem; cursor: pointer;
    transition: all .15s;
  }
  .btn-sm:hover:not(:disabled) { border-color: #e8488a44; color: #e8488a; }
  .btn-sm.active { border-color: #e8488a; color: #e8488a; background: #e8488a11; }
  .btn-sm:disabled { opacity: .3; }

  .spark { width: 100%; height: 18px; margin-top: 4px; display: block; }
  .spark polyline { fill: none; stroke: #e8488a; stroke-width: 2.5; vector-effect: non-scaling-stroke; opacity: .8; }

  .screen-wrap {
    position: relative; display: flex; justify-content: center; align-items: center;
    padding: 14px; background: #060606; min-height: 160px;
  }
  .screen {
    max-width: 100%; max-height: 320px; border-radius: 4px;
    image-rendering: pixelated; border: 1px solid #1a1a1a;
  }
  .no-screen {
    display: flex; flex-direction: column;
    align-items: center; gap: 6px; color: #2a2a2a; text-align: center; font-size: 0.8rem;
  }
  small { color: #1f1f1f; font-size: 0.68rem; }
</style>
