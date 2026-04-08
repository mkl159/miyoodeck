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

  onMount(loadSystems)

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
      await api.launch(rom.path, selectedSystem.name)
      setTimeout(() => { launching = null }, 2500)
    } catch (e) {
      error = e.message; launching = null
    }
  }

  $: filteredRoms = roms.filter(r => r.name.toLowerCase().includes(search.toLowerCase()))
</script>

<div class="launcher">
  <div class="systems-panel">
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
    {#if !selectedSystem}
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
