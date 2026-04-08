<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  let configs = []
  let selected = null
  let content = ''
  let original = ''
  let loading = false
  let saving = false
  let error = ''
  let saveMsg = ''
  let modified = false

  onMount(load)

  async function load() {
    loading = true
    try { configs = await api.configList() }
    catch (e) { error = e.message }
    loading = false
  }

  async function open(cfg) {
    if (modified && !confirm($t.unsaved + ' — OK?')) return
    loading = true; error = ''; saveMsg = ''
    try {
      const res = await api.configRead(cfg.path)
      content = original = res.content; selected = cfg; modified = false
    } catch (e) { error = e.message }
    loading = false
  }

  async function save() {
    if (!selected) return
    saving = true; error = ''
    try {
      const res = await api.configWrite(selected.path, content)
      original = content; modified = false
      saveMsg = $t.savedMsg + ' (' + res.backup.split('/').pop() + ')'
      setTimeout(() => saveMsg = '', 4000)
    } catch (e) { error = e.message }
    saving = false
  }

  function onInput() { modified = content !== original }

  function onKeydown(e) {
    if (e.key === 'Tab') {
      e.preventDefault()
      const s = e.target.selectionStart
      content = content.substring(0, s) + '  ' + content.substring(e.target.selectionEnd)
      setTimeout(() => { e.target.selectionStart = e.target.selectionEnd = s + 2 }, 0)
      modified = true
    }
    if ((e.ctrlKey || e.metaKey) && e.key === 's') { e.preventDefault(); save() }
  }
</script>

<div class="editor">
  <div class="sidebar">
    <div class="sidebar-hdr">{$t.configFiles}</div>
    {#if loading && !configs.length}
      <p class="state">...</p>
    {:else if configs.length === 0}
      <p class="state">—</p>
    {:else}
      {#each configs as cfg}
        <button
          class="cfg-btn"
          class:active={selected?.path === cfg.path}
          on:click={() => open(cfg)}
        >
          <span class="cfg-name">{cfg.name}</span>
          <span class="cfg-type">{cfg.type}</span>
        </button>
      {/each}
    {/if}
  </div>

  <div class="panel">
    {#if !selected}
      <div class="empty-state">
        <span class="big">⚙️</span>
        <p>{$t.selectConfig}</p>
        <small>{$t.backupNote}</small>
      </div>
    {:else}
      <div class="toolbar">
        <div class="file-info">
          <span class="path">{selected.path.replace('/mnt/SDCARD/', '')}</span>
          {#if modified}<span class="dot" title={$t.unsaved}>●</span>{/if}
        </div>
        <div class="toolbar-right">
          {#if saveMsg}<span class="save-ok">{saveMsg}</span>{/if}
          <button class="btn-save" on:click={save} disabled={saving || !modified}>
            {saving ? $t.saving : $t.save}
          </button>
        </div>
      </div>

      {#if error}
        <div class="errbar">{error}</div>
      {/if}

      <textarea
        class="code lang-{selected.type}"
        bind:value={content}
        on:input={onInput}
        on:keydown={onKeydown}
        spellcheck="false"
        autocomplete="off"
        autocapitalize="off"
      ></textarea>
    {/if}
  </div>
</div>

<style>
  .editor { display: flex; height: 100%; overflow: hidden; }

  .sidebar {
    width: 190px; flex-shrink: 0;
    background: #080808; border-right: 1px solid #111; overflow-y: auto;
  }
  .sidebar-hdr {
    padding: 10px 12px; font-size: 0.65rem;
    text-transform: uppercase; letter-spacing: 1.5px; color: #333;
    border-bottom: 1px solid #111;
  }
  .cfg-btn {
    display: flex; flex-direction: column; align-items: flex-start; width: 100%;
    padding: 9px 12px; background: none; border: none; border-left: 2px solid transparent;
    color: #666; cursor: pointer; text-align: left; transition: all 0.12s;
  }
  .cfg-btn:hover { background: #0d0d0d; color: #ccc; }
  .cfg-btn.active { color: #e8488a; border-left-color: #e8488a; background: #e8488a08; }
  .cfg-name { font-size: 0.8rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; width: 100%; }
  .cfg-type { font-size: 0.62rem; color: #333; margin-top: 2px; text-transform: uppercase; }
  .cfg-btn.active .cfg-type { color: #e8488a55; }

  .panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
  .toolbar {
    display: flex; align-items: center; justify-content: space-between;
    padding: 8px 14px; background: #080808; border-bottom: 1px solid #111; gap: 10px;
  }
  .file-info { display: flex; align-items: center; gap: 6px; overflow: hidden; }
  .path { font-size: 0.75rem; color: #555; font-family: monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .dot { color: #e8488a; font-size: 0.55rem; }
  .toolbar-right { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
  .save-ok { font-size: 0.72rem; color: #6bff9d; white-space: nowrap; }
  .btn-save {
    background: linear-gradient(135deg, #e8488a, #c42d6e);
    border: none; border-radius: 6px; color: #fff;
    padding: 5px 12px; font-size: 0.75rem; font-weight: 700;
    cursor: pointer; transition: all 0.15s; white-space: nowrap;
  }
  .btn-save:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 2px 12px #e8488a44; }
  .btn-save:disabled { opacity: 0.3; cursor: default; transform: none; box-shadow: none; }

  .errbar {
    background: #ff6b6b0d; color: #ff6b6b; padding: 6px 14px;
    font-size: 0.78rem; border-bottom: 1px solid #ff6b6b1a;
  }

  .code {
    flex: 1; resize: none; outline: none;
    background: #050505; color: #ccc;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 0.82rem; line-height: 1.65;
    padding: 14px; border: none; tab-size: 2;
  }
  .lang-json { color: #9cdcfe; }
  .lang-cfg  { color: #ce9178; }
  .lang-sh   { color: #4ec9b0; }

  .empty-state {
    flex: 1; display: flex; flex-direction: column; align-items: center;
    justify-content: center; gap: 8px; color: #333;
  }
  .big { font-size: 2.2rem; }
  p { font-size: 0.82rem; }
  small { font-size: 0.68rem; color: #2a2a2a; }
  .state { color: #333; padding: 14px; text-align: center; font-size: 0.8rem; }
</style>
