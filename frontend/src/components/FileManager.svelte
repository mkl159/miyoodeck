<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  const SDCARD = '/mnt/SDCARD'
  let currentPath = SDCARD
  let files = []
  let loading = false
  let error = ''
  let uploading = false
  let dragging = false
  let toastMsg = ''

  onMount(() => loadDir(SDCARD))

  async function loadDir(path) {
    loading = true; error = ''
    try {
      const res = await api.files(path)
      files = res.files || []; currentPath = res.path
    } catch (e) { error = e.message }
    loading = false
  }

  function navigate(file) { if (file.is_dir) loadDir(file.path) }

  function goUp() {
    const parts = currentPath.split('/')
    if (parts.length <= 2) return
    parts.pop()
    loadDir(parts.join('/') || '/')
  }

  async function deleteFile(file) {
    if (!confirm($t.deleteConfirm(file.name))) return
    try {
      await api.delete(file.path)
      toast($t.deleteSuccess + file.name)
      loadDir(currentPath)
    } catch (e) { error = e.message }
  }

  async function unzipFile(file) {
    try {
      const res = await api.unzip(file.path, currentPath)
      toast(res.message); loadDir(currentPath)
    } catch (e) { error = e.message }
  }

  async function handleUpload(fileList) {
    if (!fileList?.length) return
    uploading = true
    try {
      const res = await api.upload(fileList, currentPath)
      toast(res.message || $t.uploadComplete); loadDir(currentPath)
    } catch (e) { error = e.message }
    uploading = false
  }

  function onFilePick(e) { handleUpload(e.target.files); e.target.value = '' }
  function onDrop(e) { e.preventDefault(); dragging = false; handleUpload(e.dataTransfer.files) }
  function downloadSaves() { const a = document.createElement('a'); a.href = api.savesBackupUrl(); a.click() }

  function toast(msg) { toastMsg = msg; setTimeout(() => toastMsg = '', 3000) }
  function fmtSize(b) {
    if (b < 1024) return b + ' B'
    if (b < 1048576) return (b/1024).toFixed(1) + ' KB'
    return (b/1048576).toFixed(1) + ' MB'
  }
  function getIcon(name) {
    const ext = name.split('.').pop().toLowerCase()
    return { gba:'🎮', sfc:'🎮', smc:'🎮', nes:'🎮', gb:'🎮', gbc:'🎮', md:'🎮',
      gen:'🎮', pce:'🎮', n64:'🎮', z64:'🎮', iso:'💿', zip:'🗜', '7z':'🗜',
      png:'🖼', jpg:'🖼', jpeg:'🖼', cfg:'⚙', json:'⚙', txt:'📄', sh:'📜' }[ext] || '📄'
  }

  $: breadcrumbs = currentPath.replace(SDCARD, '').split('/').filter(Boolean)
</script>

<div class="fm">
  <div class="toolbar">
    <div class="breadcrumbs">
      <button class="crumb root" on:click={() => loadDir(SDCARD)}>SD</button>
      {#each breadcrumbs as crumb, i}
        <span class="sep">›</span>
        <button class="crumb" on:click={() => loadDir(SDCARD + '/' + breadcrumbs.slice(0,i+1).join('/'))}>
          {crumb}
        </button>
      {/each}
    </div>
    <div class="actions">
      <button class="act" on:click={goUp} disabled={currentPath === SDCARD}>↑</button>
      <label class="act">
        {$t.upload}
        <input type="file" multiple on:change={onFilePick} hidden />
      </label>
      <button class="act green" on:click={downloadSaves}>{$t.savesBackup}</button>
    </div>
  </div>

  {#if error}
    <div class="errbar">{error} <button on:click={() => error=''}>✕</button></div>
  {/if}
  {#if uploading}
    <div class="uploadbar">⬆ {$t.upload}...</div>
  {/if}

  <div
    class="list-wrap" class:dragging
    on:dragover|preventDefault={() => dragging = true}
    on:dragleave={() => dragging = false}
    on:drop={onDrop}
    role="region"
    aria-label="File drop zone"
  >
    {#if dragging}
      <div class="drop-overlay">
        <span>{$t.dropToUpload}<br/>{currentPath}</span>
      </div>
    {/if}

    {#if loading}
      <p class="state">...</p>
    {:else}
      <div class="list">
        {#each files as file}
          <div class="row">
            <button class="row-main" on:click={() => navigate(file)}>
              <span class="icon">{file.is_dir ? '📁' : getIcon(file.name)}</span>
              <span class="name">{file.name}</span>
              {#if !file.is_dir}<span class="meta">{fmtSize(file.size)}</span>{/if}
              <span class="date">{file.mod_time}</span>
            </button>
            <div class="row-actions">
              {#if !file.is_dir && file.name.endsWith('.zip')}
                <button class="fa" title={$t.extract} on:click|stopPropagation={() => unzipFile(file)}>📦</button>
              {/if}
              <button class="fa del" on:click|stopPropagation={() => deleteFile(file)}>✕</button>
            </div>
          </div>
        {:else}
          <p class="state">{$t.emptyDir}</p>
        {/each}
      </div>
    {/if}
  </div>

  {#if toastMsg}
    <div class="toast">{toastMsg}</div>
  {/if}
</div>

<style>
  .fm { display: flex; flex-direction: column; height: 100%; overflow: hidden; }

  .toolbar {
    display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
    padding: 8px 14px; background: #080808; border-bottom: 1px solid #111;
  }
  .breadcrumbs { display: flex; align-items: center; gap: 3px; flex: 1; overflow: hidden; }
  .crumb {
    background: none; border: none; color: #555; cursor: pointer;
    font-size: 0.78rem; padding: 2px 4px; border-radius: 4px;
    white-space: nowrap; max-width: 90px; overflow: hidden; text-overflow: ellipsis;
  }
  .crumb:hover { color: #e8488a; }
  .crumb.root { color: #e8488a66; font-weight: 700; }
  .sep { color: #2a2a2a; font-size: 0.7rem; }
  .actions { display: flex; gap: 6px; }
  .act {
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 6px;
    color: #666; padding: 4px 10px; font-size: 0.75rem; cursor: pointer;
    transition: all 0.12s; white-space: nowrap;
  }
  .act:hover:not(:disabled) { border-color: #e8488a44; color: #e8488a; }
  .act:disabled { opacity: 0.3; cursor: default; }
  .act.green:hover { border-color: #6bff9d44; color: #6bff9d; }

  .errbar {
    background: #ff6b6b0d; border-bottom: 1px solid #ff6b6b22;
    color: #ff6b6b; padding: 7px 14px; font-size: 0.78rem;
    display: flex; justify-content: space-between;
  }
  .errbar button { background: none; border: none; color: inherit; cursor: pointer; }
  .uploadbar {
    background: #e8488a0d; border-bottom: 1px solid #e8488a22;
    color: #e8488a; padding: 7px 14px; font-size: 0.78rem;
  }

  .list-wrap { flex: 1; overflow: hidden; position: relative; }
  .list { height: 100%; overflow-y: auto; }
  .dragging .list { opacity: 0.3; }
  .drop-overlay {
    position: absolute; inset: 0; z-index: 10;
    background: #e8488a0d; border: 2px dashed #e8488a66;
    display: flex; align-items: center; justify-content: center;
    text-align: center; color: #e8488a; font-size: 0.9rem;
    white-space: pre-line; pointer-events: none;
  }

  .row { display: flex; align-items: center; border-bottom: 1px solid #0d0d0d; }
  .row:hover { background: #0d0d0d; }
  .row-main {
    flex: 1; display: flex; align-items: center; gap: 8px;
    padding: 8px 14px; background: none; border: none;
    color: #bbb; cursor: pointer; text-align: left; overflow: hidden;
  }
  .icon { flex-shrink: 0; font-size: 0.85rem; }
  .name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.82rem; }
  .meta { font-size: 0.68rem; color: #444; white-space: nowrap; padding: 0 6px; }
  .date { font-size: 0.65rem; color: #333; white-space: nowrap; }
  .row-actions { display: flex; gap: 2px; padding-right: 8px; }
  .fa {
    background: none; border: none; color: #333; cursor: pointer;
    padding: 3px 5px; border-radius: 3px; font-size: 0.75rem; transition: all 0.12s;
  }
  .fa:hover { background: #111; color: #bbb; }
  .fa.del:hover { color: #ff6b6b; }
  .state { color: #444; padding: 16px; text-align: center; font-size: 0.82rem; }

  .toast {
    position: fixed; bottom: 20px; left: 50%; transform: translateX(-50%);
    background: #111; border: 1px solid #e8488a33;
    color: #e8488a; padding: 8px 18px; border-radius: 20px;
    font-size: 0.8rem; z-index: 100; white-space: nowrap;
    animation: fadeup 0.2s ease;
  }
  @keyframes fadeup { from { opacity: 0; transform: translateX(-50%) translateY(8px); } }
</style>
