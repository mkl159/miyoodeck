<script>
  import { onMount } from 'svelte'
  import { api, setToken, getToken, createWS } from './api.js'
  import { t, lang, setLang, availableLangs } from './i18n.js'
  import Dashboard from './components/Dashboard.svelte'
  import FileManager from './components/FileManager.svelte'
  import GameLauncher from './components/GameLauncher.svelte'
  import ConfigEditor from './components/ConfigEditor.svelte'
  import Login from './components/Login.svelte'

  let tab = 'dashboard'
  let authenticated = false
  let pinConfigured = false
  let loading = true
  let stats = null
  let screenshot = null
  let wsConnected = false
  // Fix #13: track connection error for helpful message
  let connectionError = false
  let wsControls = null

  $: tabs = [
    { id: 'dashboard', label: $t.navDashboard, icon: '◉' },
    { id: 'games',     label: $t.navGames,     icon: '▶' },
    { id: 'files',     label: $t.navFiles,     icon: '⊞' },
    { id: 'config',    label: $t.navConfig,    icon: '≡' },
  ]

  onMount(async () => {
    try {
      const status = await api.authStatus()
      pinConfigured = status.pin_configured
      if (!pinConfigured || getToken()) {
        try { await api.system(); authenticated = true }
        catch (e) {
          if (!pinConfigured) {
            const r = await api.login('')
            setToken(r.token); authenticated = true
          }
        }
      }
    } catch (e) {
      // Fix #13: if we can't reach the server at all, show a helpful error
      if (e.message.includes('Failed to fetch') || e.message.includes('NetworkError')) {
        connectionError = true
      }
    }
    loading = false
    if (authenticated) startWS()
  })

  function startWS() {
    wsControls = createWS(
      (msg) => {
        wsConnected = true
        if (msg.type === 'stats') stats = msg.data
        if (msg.type === 'screenshot') screenshot = msg.data
      },
      (connected) => { wsConnected = connected }  // Fix #12: status callback
    )
  }

  function onLogin(token) {
    setToken(token)
    authenticated = true
    connectionError = false
    startWS()
  }

  // Fix #12: stop WS cleanly on logout
  async function logout() {
    wsControls?.stop()
    wsControls = null
    await api.logout()
    setToken('')
    authenticated = false
    stats = null; screenshot = null; wsConnected = false
  }
</script>

<div class="app">
  {#if loading}
    <div class="splash">
      <div class="logo">
        <div class="ring r1"></div><div class="ring r2"></div><div class="ring r3"></div>
      </div>
      <span class="title">MiyooDeck</span>
      <span class="sub">{$t.connecting}</span>
    </div>

  {:else if connectionError}
    <!-- Fix #13: friendly WiFi error page -->
    <div class="splash">
      <div class="err-icon">📡</div>
      <span class="title err">{$t.connErrTitle}</span>
      <div class="err-card">
        <p>{$t.connErrMsg} <code>:8080</code></p>
        <ol>
          <li>{$t.connErrStep1}</li>
          <li>{$t.connErrStep2}</li>
          <li>{$t.connErrStep3}</li>
          <li>{$t.connErrStep4}</li>
        </ol>
        <button class="btn-retry" on:click={() => location.reload()}>{$t.connErrRetry}</button>
      </div>
    </div>

  {:else if !authenticated}
    <Login {pinConfigured} on:login={(e) => onLogin(e.detail)} />

  {:else}
    <header class="header">
      <div class="logo-sm">
        <div class="ring r1 sm"></div><div class="ring r2 sm"></div>
      </div>
      <span class="brand">MiyooDeck</span>
      <div class="right">
        {#if stats?.game_running}
          <span class="badge game">🎮 {$t.gameRunning}</span>
        {/if}
        {#if stats}
          <span class="badge" class:charging={stats.battery.charging}>
            {stats.battery.charging ? '⚡' : '🔋'}{stats.battery.percent >= 0 ? stats.battery.percent + '%' : '?'}
          </span>
          <span class="badge cpu">CPU {stats.cpu_percent.toFixed(0)}%</span>
        {/if}
        <div class="lang-sw">
          {#each availableLangs as l}
            <button class="lbtn" class:active={$lang === l.code} on:click={() => setLang(l.code)}
              title={l.label}>{l.code.toUpperCase()}</button>
          {/each}
        </div>
        <button class="logout-btn" on:click={logout} title="Déconnexion">⏏</button>
        <span class="dot" class:on={wsConnected}></span>
      </div>
    </header>

    <nav class="nav">
      {#each tabs as tb}
        <button class="nb" class:active={tab === tb.id} on:click={() => tab = tb.id}>
          <span class="ni">{tb.icon}</span>
          <span class="nl">{tb.label}</span>
        </button>
      {/each}
    </nav>

    <main class="main">
      {#if tab === 'dashboard'}  <Dashboard {stats} {screenshot} />
      {:else if tab === 'games'} <GameLauncher />
      {:else if tab === 'files'} <FileManager />
      {:else if tab === 'config'}<ConfigEditor />
      {/if}
    </main>
  {/if}
</div>

<style>
  :global(*){box-sizing:border-box;margin:0;padding:0}
  :global(body){background:#0a0a0a;color:#e0e0e0;font-family:'Segoe UI',system-ui,sans-serif;min-height:100vh}
  :global(::-webkit-scrollbar){width:5px}
  :global(::-webkit-scrollbar-track){background:#0a0a0a}
  :global(::-webkit-scrollbar-thumb){background:#e8488a33;border-radius:3px}

  .app{display:flex;flex-direction:column;min-height:100vh}

  /* Splash / error */
  .splash{display:flex;flex-direction:column;align-items:center;justify-content:center;min-height:100vh;gap:18px}
  .logo{position:relative;width:80px;height:80px}
  .ring{position:absolute;border-radius:50%;border:3px solid;top:50%;left:50%;transform:translate(-50%,-50%);animation:pulse 2s ease-in-out infinite}
  .r1{width:80px;height:80px;border-color:#e8488a}
  .r2{width:52px;height:52px;border-color:#e8488a55;animation-delay:.3s}
  .r3{width:26px;height:26px;border-color:#e8488a22;animation-delay:.6s}
  .r1.sm{width:28px;height:28px;border-width:2px}
  .r2.sm{width:16px;height:16px;border-width:2px;animation-delay:.3s}
  @keyframes pulse{0%,100%{opacity:1}50%{opacity:.25}}
  .title{font-size:1.8rem;font-weight:800;color:#e8488a;letter-spacing:2px}
  .title.err{color:#ff6b6b;font-size:1.3rem}
  .sub{font-size:0.78rem;color:#333}
  .err-icon{font-size:3rem}
  .err-card{background:#0d0d0d;border:1px solid #ff6b6b22;border-radius:14px;padding:20px 24px;max-width:340px;display:flex;flex-direction:column;gap:12px}
  .err-card p{font-size:0.85rem;color:#888}
  .err-card code{color:#e8488a;background:#e8488a11;padding:1px 5px;border-radius:4px}
  .err-card ol{padding-left:18px;font-size:0.82rem;color:#666;display:flex;flex-direction:column;gap:5px}
  .err-card li{color:#ccc}
  .btn-retry{background:linear-gradient(135deg,#e8488a,#c42d6e);border:none;border-radius:8px;color:#fff;padding:9px;font-size:0.85rem;font-weight:700;cursor:pointer;transition:all .2s}
  .btn-retry:hover{transform:translateY(-1px);box-shadow:0 4px 16px #e8488a44}

  /* Header */
  .header{display:flex;align-items:center;gap:10px;padding:9px 16px;background:#0d0d0d;border-bottom:1px solid #e8488a1a}
  .logo-sm{position:relative;width:26px;height:26px;flex-shrink:0}
  .brand{font-weight:800;color:#e8488a;font-size:1rem;letter-spacing:1px;flex:1}
  .right{display:flex;align-items:center;gap:7px}
  .badge{font-size:0.68rem;padding:2px 7px;background:#111;border:1px solid #1e1e1e;border-radius:10px;color:#666}
  .badge.charging{border-color:#6bff9d22;color:#6bff9d}
  .badge.cpu{border-color:#e8488a1a}
  .badge.game{border-color:#e8488a44;color:#e8488a;animation:pulse 2s ease-in-out infinite}
  .lang-sw{display:flex;gap:2px}
  .lbtn{background:none;border:1px solid #1a1a1a;border-radius:4px;color:#444;font-size:0.62rem;padding:2px 5px;cursor:pointer;font-weight:700;transition:all .12s}
  .lbtn:hover{border-color:#e8488a44;color:#e8488a88}
  .lbtn.active{border-color:#e8488a;color:#e8488a;background:#e8488a0d}
  .logout-btn{background:none;border:none;color:#333;cursor:pointer;font-size:0.9rem;padding:2px 4px;transition:color .15s;line-height:1}
  .logout-btn:hover{color:#ff6b6b}
  .dot{width:7px;height:7px;border-radius:50%;background:#222;transition:all .4s;flex-shrink:0}
  .dot.on{background:#6bff9d;box-shadow:0 0 5px #6bff9d66}

  /* Nav */
  .nav{display:flex;background:#080808;border-bottom:1px solid #0d0d0d}
  .nb{flex:1;display:flex;flex-direction:column;align-items:center;gap:2px;padding:9px 4px;background:none;border:none;color:#333;cursor:pointer;border-bottom:2px solid transparent;transition:all .12s}
  .nb:hover{color:#e8488a55}
  .nb.active{color:#e8488a;border-bottom-color:#e8488a;background:#e8488a06}
  .ni{font-size:0.95rem}
  .nl{font-size:0.6rem}

  .main{flex:1;overflow:hidden;display:flex;flex-direction:column}
</style>
