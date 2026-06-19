<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  let systems = []
  let selectedSystem = null
  let roms = []
  let loading = false
  let launching = null
  let error = ''
  let search = ''

  // Global cross-system search + random launcher
  let globalQuery = ''
  let globalResults = []
  let searching = false
  let gTimer = null
  let surprising = false
  $: showGlobal = globalQuery.trim().length >= 2

  onMount(loadSystems)

  function onGlobalInput() {
    clearTimeout(gTimer)
    if (globalQuery.trim().length < 2) { globalResults = []; return }
    gTimer = setTimeout(doGlobalSearch, 250)
  }

  async function doGlobalSearch() {
    searching = true; error = ''
    try { globalResults = await api.searchRoms(globalQuery.trim()) }
    catch (e) { error = e.message; globalResults = [] }
    searching = false
  }

  async function surprise() {
    surprising = true; error = ''
    try {
      const rom = await api.randomRom(selectedSystem?.name)
      await launchRom(rom)
    } catch (e) { error = e.message }
    surprising = false
  }

  async function loadSystems() {
    loading = true; error = ''
    try { systems = await api.systems() }
    catch (e) { error = e.message }
    loading = false
  }

  async function selectSystem(sys) {
    selectedSystem = sys; loading = true; error = ''; search = ''
    try { roms = await api.roms(sys.name) }
    catch (e) { error = e.message; roms = [] }
    loading = false
  }

  async function launchRom(rom) {
    launching = rom.path; error = ''
    try {
      await api.launch(rom.path, rom.system || selectedSystem?.name)
      setTimeout(() => { launching = null }, 2500)
    } catch (e) {
      error = e.message; launching = null
    }
  }

  $: filteredRoms = roms.filter(r => r.name.toLowerCase().includes(search.toLowerCase()))
</script>

<div class="launcher">
  <div class="systems-panel">
    <div class="global-bar">
      <input type="search" class="global-search" placeholder={$t.searchAll}
        bind:value={globalQuery} on:input={onGlobalInput} />
      <button class="surprise-btn" on:click={surprise} disabled={surprising || launching !== null}
        title={$t.surpriseMe}>{$t.surpriseMe}</button>
    </div>
    <div class="panel-header">{$t.systems}</div>
    {#if loading && !selectedSystem}
      <p class="state-msg">...</p>
    {:else if systems.length === 0}
      <p class="state-msg">{$t.noSystemsFound}</p>
    {:else}
      {#each systems as sys}
        <button
          class="sys-btn"
          class:active={selectedSystem?.name === sys.name}
          on:click={() => selectSystem(sys)}
        >
          <span class="sys-name">{sys.name}</span>
          <span class="sys-count">{sys.rom_count}</span>
        </button>
      {/each}
    {/if}
  </div>

  <div class="roms-panel">
    {#if showGlobal}
      <div class="roms-header">
        <h2>🔍 {$t.searchResults}</h2>
        <span class="result-count">{globalResults.length}</span>
      </div>
      {#if error}<div class="error-bar">{error}</div>{/if}
      {#if searching}
        <p class="state-msg">...</p>
      {:else}
        <div class="roms-list">
          {#each globalResults as rom (rom.path)}
            <button class="rom-btn" class:launching={launching === rom.path}
              on:click={() => launchRom(rom)} disabled={launching !== null}>
              <span class="rom-icon">{launching === rom.path ? '▶' : '◉'}</span>
              <span class="rom-name">{rom.name}</span>
              <span class="sys-tag">{rom.system}</span>
            </button>
          {:else}
            <p class="state-msg">"{globalQuery}" — {$t.noRoms}</p>
          {/each}
        </div>
      {/if}
    {:else if !selectedSystem}
      <div class="empty-state">
        <span class="big-icon">🎮</span>
        <p>{$t.selectSystem}</p>
        <small>{$t.selectSystemSub}</small>
      </div>
    {:else}
      <div class="roms-header">
        <h2>{selectedSystem.name}</h2>
        <input type="search" placeholder={$t.searchRoms} bind:value={search} />
      </div>

      {#if error}
        <div class="error-bar">{error}</div>
      {/if}

      {#if loading}
        <p class="state-msg">...</p>
      {:else}
        <div class="roms-list">
          {#each filteredRoms as rom}
            <button
              class="rom-btn"
              class:launching={launching === rom.path}
              on:click={() => launchRom(rom)}
              disabled={launching !== null}
            >
              <span class="rom-icon">{launching === rom.path ? '▶' : '◉'}</span>
              <span class="rom-name">{rom.name}</span>
              <span class="rom-size">{(rom.size / 1048576).toFixed(1)} MB</span>
              {#if launching === rom.path}
                <span class="badge-launch">{$t.launching}</span>
              {/if}
            </button>
          {:else}
            <p class="state-msg">{search ? `"${search}" — ${$t.noRoms}` : $t.noRoms}</p>
          {/each}
        </div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .launcher { display: flex; height: 100%; overflow: hidden; }

  .systems-panel {
    width: 150px; flex-shrink: 0;
    background: #080808; border-right: 1px solid #111;
    overflow-y: auto; display: flex; flex-direction: column;
  }
  .panel-header {
    padding: 10px 12px; font-size: 0.65rem;
    text-transform: uppercase; letter-spacing: 1.5px; color: #333;
    border-bottom: 1px solid #111;
  }
  .global-bar { display: flex; flex-direction: column; gap: 6px; padding: 10px; border-bottom: 1px solid #111; }
  .global-search {
    width: 100%; background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 8px;
    color: #ccc; padding: 6px 9px; font-size: 0.75rem; outline: none;
  }
  .global-search:focus { border-color: #e8488a33; }
  .surprise-btn {
    width: 100%; background: linear-gradient(135deg,#e8488a,#c42d6e); border: none;
    border-radius: 8px; color: #fff; padding: 6px; font-size: 0.75rem; font-weight: 700;
    cursor: pointer; transition: all .15s;
  }
  .surprise-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 3px 14px #e8488a44; }
  .surprise-btn:disabled { opacity: .4; cursor: default; }
  .result-count {
    font-size: 0.65rem; color: #444; background: #111; padding: 1px 7px; border-radius: 8px;
  }
  .sys-tag {
    font-size: 0.62rem; color: #6b9dff; background: #6b9dff11; border: 1px solid #6b9dff22;
    padding: 1px 6px; border-radius: 6px; white-space: nowrap;
  }
  .sys-btn {
    display: flex; align-items: center; justify-content: space-between;
    padding: 9px 12px; background: none; border: none;
    color: #666; cursor: pointer; text-align: left;
    border-left: 2px solid transparent; transition: all 0.12s; font-size: 0.82rem;
  }
  .sys-btn:hover { color: #ccc; background: #0d0d0d; }
  .sys-btn.active { color: #e8488a; border-left-color: #e8488a; background: #e8488a08; }
  .sys-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sys-count {
    font-size: 0.65rem; color: #444; background: #111;
    padding: 1px 5px; border-radius: 8px; margin-left: 4px;
  }

  .roms-panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
  .roms-header {
    display: flex; align-items: center; gap: 10px; padding: 10px 14px;
    border-bottom: 1px solid #111; background: #080808;
  }
  h2 { font-size: 0.95rem; color: #e8488a; white-space: nowrap; font-weight: 700; }
  input[type=search] {
    flex: 1; background: #0d0d0d; border: 1px solid #1a1a1a;
    border-radius: 8px; color: #ccc; padding: 5px 10px;
    font-size: 0.82rem; outline: none;
  }
  input[type=search]:focus { border-color: #e8488a33; }

  .roms-list { flex: 1; overflow-y: auto; padding: 6px; display: flex; flex-direction: column; gap: 2px; }
  .rom-btn {
    display: flex; align-items: center; gap: 8px; padding: 9px 12px;
    background: #0d0d0d; border: 1px solid #111; border-radius: 8px;
    color: #bbb; cursor: pointer; text-align: left; transition: all 0.12s;
  }
  .rom-btn:hover:not(:disabled) { border-color: #e8488a22; background: #e8488a06; color: #e0e0e0; }
  .rom-btn.launching { border-color: #e8488a55; background: #e8488a0d; color: #e8488a; }
  .rom-btn:disabled { opacity: 0.5; cursor: default; }
  .rom-icon { font-size: 0.65rem; color: #333; flex-shrink: 0; }
  .rom-btn.launching .rom-icon { color: #e8488a; }
  .rom-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.85rem; }
  .rom-size { font-size: 0.68rem; color: #444; white-space: nowrap; }
  .badge-launch {
    font-size: 0.65rem; color: #e8488a; background: #e8488a15;
    padding: 2px 7px; border-radius: 8px; white-space: nowrap;
  }

  .empty-state {
    flex: 1; display: flex; flex-direction: column; align-items: center;
    justify-content: center; gap: 10px; color: #333;
  }
  .big-icon { font-size: 2.5rem; }
  p { font-size: 0.85rem; color: #555; }
  small { font-size: 0.72rem; color: #333; }
  .state-msg { color: #444; padding: 16px; text-align: center; font-size: 0.82rem; }
  .error-bar {
    background: #ff6b6b0d; border-bottom: 1px solid #ff6b6b22;
    color: #ff6b6b; padding: 6px 14px; font-size: 0.78rem;
  }
</style>
