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
        try {
          await api.system()
          authenticated = true
        } catch {
          if (!pinConfigured) {
            const res = await api.login('')
            setToken(res.token)
            authenticated = true
          }
        }
      }
    } catch (e) {
      console.error('Init error:', e)
    }
    loading = false
    if (authenticated) startWS()
  })

  function startWS() {
    createWS((msg) => {
      wsConnected = true
      if (msg.type === 'stats') stats = msg.data
      if (msg.type === 'screenshot') screenshot = msg.data
    })
  }

  function onLogin(token) {
    setToken(token)
    authenticated = true
    startWS()
  }
</script>

<div class="app">
  {#if loading}
    <div class="splash">
      <div class="logo">
        <div class="onion-ring r1"></div>
        <div class="onion-ring r2"></div>
        <div class="onion-ring r3"></div>
      </div>
      <span class="splash-title">MiyooDeck</span>
      <span class="splash-sub">{$t.connecting}</span>
    </div>

  {:else if !authenticated}
    <Login {pinConfigured} on:login={(e) => onLogin(e.detail)} />

  {:else}
    <header class="header">
      <div class="logo-sm">
        <div class="onion-ring r1 sm"></div>
        <div class="onion-ring r2 sm"></div>
      </div>
      <span class="brand">MiyooDeck</span>

      <div class="header-right">
        {#if stats}
          <span class="badge" class:charging={stats.battery.charging}>
            {stats.battery.charging ? '⚡' : '🔋'}
            {stats.battery.percent >= 0 ? stats.battery.percent + '%' : 'N/A'}
          </span>
          <span class="badge cpu">CPU {stats.cpu_percent.toFixed(0)}%</span>
        {/if}

        <!-- Lang switcher -->
        <div class="lang-switch">
          {#each availableLangs as l}
            <button
              class="lang-btn"
              class:active={$lang === l.code}
              on:click={() => setLang(l.code)}
              title={l.label}
            >{l.code.toUpperCase()}</button>
          {/each}
        </div>

        <span class="ws-dot" class:connected={wsConnected}
          title={wsConnected ? 'Live' : $t.connecting}></span>
      </div>
    </header>

    <nav class="nav">
      {#each tabs as t}
        <button
          class="nav-btn"
          class:active={tab === t.id}
          on:click={() => tab = t.id}
        >
          <span class="nav-icon">{t.icon}</span>
          <span class="nav-label">{t.label}</span>
        </button>
      {/each}
    </nav>

    <main class="main">
      {#if tab === 'dashboard'}
        <Dashboard {stats} {screenshot} />
      {:else if tab === 'games'}
        <GameLauncher />
      {:else if tab === 'files'}
        <FileManager />
      {:else if tab === 'config'}
        <ConfigEditor />
      {/if}
    </main>
  {/if}
</div>

<style>
  :global(*) { box-sizing: border-box; margin: 0; padding: 0; }
  :global(body) {
    background: #0a0a0a;
    color: #e0e0e0;
    font-family: 'Segoe UI', system-ui, sans-serif;
    min-height: 100vh;
  }
  :global(::-webkit-scrollbar) { width: 6px; }
  :global(::-webkit-scrollbar-track) { background: #111; }
  :global(::-webkit-scrollbar-thumb) { background: #e8488a44; border-radius: 3px; }

  .app { display: flex; flex-direction: column; min-height: 100vh; }

  /* Splash */
  .splash {
    display: flex; flex-direction: column; align-items: center;
    justify-content: center; height: 100vh; gap: 20px;
  }
  .logo { position: relative; width: 80px; height: 80px; }
  .onion-ring {
    position: absolute; border-radius: 50%; border: 3px solid;
    top: 50%; left: 50%; transform: translate(-50%, -50%);
    animation: pulse 2s ease-in-out infinite;
  }
  .r1 { width: 80px; height: 80px; border-color: #e8488a; }
  .r2 { width: 52px; height: 52px; border-color: #e8488a66; animation-delay: 0.3s; }
  .r3 { width: 26px; height: 26px; border-color: #e8488a33; animation-delay: 0.6s; }
  .r1.sm { width: 28px; height: 28px; border-width: 2px; }
  .r2.sm { width: 16px; height: 16px; border-width: 2px; }
  @keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.3} }
  .splash-title { font-size: 2rem; font-weight: 800; color: #e8488a; letter-spacing: 3px; }
  .splash-sub { font-size: 0.8rem; color: #444; }

  /* Header */
  .header {
    display: flex; align-items: center; gap: 12px;
    padding: 10px 16px;
    background: #0d0d0d;
    border-bottom: 1px solid #e8488a22;
  }
  .logo-sm { position: relative; width: 28px; height: 28px; flex-shrink: 0; }
  .brand { font-weight: 800; color: #e8488a; font-size: 1.1rem; flex: 1; letter-spacing: 1px; }
  .header-right { display: flex; align-items: center; gap: 8px; }
  .badge {
    font-size: 0.72rem; padding: 2px 8px;
    background: #1a1a1a; border: 1px solid #2a2a2a;
    border-radius: 10px; color: #888;
  }
  .badge.charging { border-color: #6bff9d33; color: #6bff9d; }
  .badge.cpu { border-color: #e8488a22; }

  /* Lang switch */
  .lang-switch { display: flex; gap: 2px; }
  .lang-btn {
    background: none; border: 1px solid #222; border-radius: 4px;
    color: #555; font-size: 0.65rem; padding: 2px 5px;
    cursor: pointer; font-weight: 600; transition: all 0.15s;
  }
  .lang-btn:hover { border-color: #e8488a44; color: #e8488a88; }
  .lang-btn.active { border-color: #e8488a; color: #e8488a; background: #e8488a11; }

  .ws-dot {
    width: 8px; height: 8px; border-radius: 50%;
    background: #333; transition: all 0.4s;
    flex-shrink: 0;
  }
  .ws-dot.connected { background: #6bff9d; box-shadow: 0 0 6px #6bff9d88; }

  /* Nav */
  .nav {
    display: flex;
    background: #0a0a0a;
    border-bottom: 1px solid #111;
  }
  .nav-btn {
    flex: 1; display: flex; flex-direction: column; align-items: center;
    gap: 2px; padding: 9px 4px;
    background: none; border: none;
    color: #444; cursor: pointer; font-size: 0.68rem;
    border-bottom: 2px solid transparent;
    transition: all 0.15s;
  }
  .nav-btn:hover { color: #e8488a66; }
  .nav-btn.active {
    color: #e8488a;
    border-bottom-color: #e8488a;
    background: #e8488a08;
  }
  .nav-icon { font-size: 1rem; }
  .nav-label { font-size: 0.65rem; }

  .main { flex: 1; overflow: hidden; display: flex; flex-direction: column; }
</style>
