<script>
  import { onMount } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  const SDCARD = '/mnt/SDCARD'
  let currentPath = SDCARD
  let files = []
  let loading = false
  let error = ''
  // Fix #6: upload progress state
  let uploading = false
  let uploadPct = 0
  let uploadLabel = ''
  let dragging = false
  let toastMsg = ''

  onMount(() => loadDir(SDCARD))

  async function loadDir(path) {
    loading = true; error = ''
    try { const r = await api.files(path); files = r.files || []; currentPath = r.path }
    catch (e) { error = e.message }
    loading = false
  }

  function navigate(f) { if (f.is_dir) loadDir(f.path) }

  function goUp() {
    const p = currentPath.split('/')
    if (p.length <= 2) return
    p.pop(); loadDir(p.join('/') || '/')
  }

  async function del(file) {
    if (!confirm($t.deleteConfirm(file.name))) return
    try { await api.delete(file.path); toast($t.deleteSuccess + file.name); loadDir(currentPath) }
    catch (e) { error = e.message }
  }

  async function unzip(file) {
    try { const r = await api.unzip(file.path, currentPath); toast(r.message); loadDir(currentPath) }
    catch (e) { error = e.message }
  }

  // Fix #10: individual file download
  function download(file) {
    const a = document.createElement('a')
    a.href = api.downloadUrl(file.path)
    a.download = file.name
    a.click()
  }

  // Fix #6: upload with real progress bar
  async function handleUpload(fileList) {
    if (!fileList?.length) return
    uploading = true; uploadPct = 0
    const names = Array.from(fileList).map(f => f.name).join(', ')
    uploadLabel = fileList.length === 1 ? names : `${fileList.length} fichiers`
    try {
      const res = await api.upload(fileList, currentPath, (pct) => { uploadPct = pct })
      toast(res.message || $t.uploadComplete)
      loadDir(currentPath)
    } catch (e) { error = e.message }
    uploading = false; uploadPct = 0
  }

  function onFilePick(e) { handleUpload(e.target.files); e.target.value = '' }
  function onDrop(e) { e.preventDefault(); dragging = false; handleUpload(e.dataTransfer.files) }
  function savesBackup() { window.open(api.savesBackupUrl(), '_blank') }

  function toast(msg) { toastMsg = msg; setTimeout(() => toastMsg = '', 3000) }
  function fmtSize(b) {
    if (b < 1024) return b + ' B'
    if (b < 1048576) return (b / 1024).toFixed(1) + ' KB'
    return (b / 1048576).toFixed(1) + ' MB'
  }
  function icon(name) {
    const e = name.split('.').pop().toLowerCase()
    return ({gba:'🎮',sfc:'🎮',smc:'🎮',nes:'🎮',gb:'🎮',gbc:'🎮',md:'🎮',
      gen:'🎮',pce:'🎮',n64:'🎮',z64:'🎮',iso:'💿',pbp:'💿',
      zip:'🗜','7z':'🗜',png:'🖼',jpg:'🖼',jpeg:'🖼',
      cfg:'⚙',json:'⚙',txt:'📄',sh:'📜'})[e] || '📄'
  }

  $: crumbs = currentPath.replace(SDCARD, '').split('/').filter(Boolean)
</script>

<div class="fm">
  <!-- Toolbar -->
  <div class="toolbar">
    <div class="breadcrumbs">
      <button class="crumb root" on:click={() => loadDir(SDCARD)}>SD</button>
      {#each crumbs as c, i}
        <span class="sep">›</span>
        <button class="crumb" on:click={() => loadDir(SDCARD + '/' + crumbs.slice(0,i+1).join('/'))}>
          {c}
        </button>
      {/each}
    </div>
    <div class="actions">
      <button class="act" on:click={goUp} disabled={currentPath === SDCARD}>↑</button>
      <label class="act">{$t.upload}<input type="file" multiple on:change={onFilePick} hidden /></label>
      <button class="act green" on:click={savesBackup}>{$t.savesBackup}</button>
    </div>
  </div>

  {#if error}
    <div class="bar-err">{error} <button on:click={() => error=''}>✕</button></div>
  {/if}

  <!-- Fix #6: upload progress bar -->
  {#if uploading}
    <div class="bar-upload">
      <div class="upload-info">
        <span>⬆ {uploadLabel}</span>
        <span class="pct">{uploadPct}%</span>
      </div>
      <div class="progress-track">
        <div class="progress-fill" style="width:{uploadPct}%"></div>
      </div>
    </div>
  {/if}

  <!-- File list + drop zone -->
  <div
    class="list-wrap" class:dragging
    on:dragover|preventDefault={() => dragging = true}
    on:dragleave={() => dragging = false}
    on:drop={onDrop}
    role="region" aria-label="Zone de dépôt de fichiers"
  >
    {#if dragging}
      <div class="drop-overlay">
        <span>⬆</span>
        <p>{$t.dropToUpload}</p>
        <small>{currentPath}</small>
      </div>
    {/if}

    {#if loading}
      <p class="state">...</p>
    {:else}
      <div class="list">
        {#each files as f (f.path)}
          <div class="row">
            <button class="row-main" on:click={() => navigate(f)}>
              <span class="ico">{f.is_dir ? '📁' : icon(f.name)}</span>
              <span class="name">{f.name}</span>
              {#if !f.is_dir}<span class="meta">{fmtSize(f.size)}</span>{/if}
              <span class="date">{f.mod_time}</span>
            </button>
            <div class="row-btns">
              {#if !f.is_dir}
                <!-- Fix #10: download button -->
                <button class="rb" title="Télécharger" on:click|stopPropagation={() => download(f)}>⬇</button>
              {/if}
              {#if !f.is_dir && f.name.endsWith('.zip')}
                <button class="rb" title={$t.extract} on:click|stopPropagation={() => unzip(f)}>📦</button>
              {/if}
              <button class="rb del" on:click|stopPropagation={() => del(f)}>✕</button>
            </div>
          </div>
        {:else}
          <p class="state">{$t.emptyDir}</p>
        {/each}
      </div>
    {/if}
  </div>

  {#if toastMsg}
    <div class="toast" role="status">{toastMsg}</div>
  {/if}
</div>

<style>
  .fm { display: flex; flex-direction: column; height: 100%; overflow: hidden; }

  .toolbar {
    display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
    padding: 8px 14px; background: #080808; border-bottom: 1px solid #111;
  }
  .breadcrumbs { display: flex; align-items: center; gap: 3px; flex: 1; overflow: hidden; min-width: 0; }
  .crumb {
    background: none; border: none; color: #444; cursor: pointer;
    font-size: 0.78rem; padding: 2px 4px; border-radius: 4px;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 90px;
  }
  .crumb:hover { color: #e8488a; }
  .crumb.root { color: #e8488a55; font-weight: 700; }
  .sep { color: #222; font-size: 0.65rem; }
  .actions { display: flex; gap: 5px; flex-shrink: 0; }
  .act {
    background: #0d0d0d; border: 1px solid #1a1a1a; border-radius: 6px;
    color: #555; padding: 4px 10px; font-size: 0.75rem; cursor: pointer;
    transition: all .12s; white-space: nowrap;
  }
  .act:hover:not(:disabled) { border-color: #e8488a44; color: #e8488a; }
  .act:disabled { opacity: .3; cursor: default; }
  .act.green:hover { border-color: #6bff9d44; color: #6bff9d; }

  /* Error bar */
  .bar-err {
    background: #ff6b6b09; border-bottom: 1px solid #ff6b6b22;
    color: #ff6b6b; padding: 6px 14px; font-size: 0.78rem;
    display: flex; justify-content: space-between;
  }
  .bar-err button { background: none; border: none; color: inherit; cursor: pointer; }

  /* Fix #6: upload progress bar */
  .bar-upload {
    background: #0d0d0d; border-bottom: 1px solid #e8488a22; padding: 6px 14px 8px;
  }
  .upload-info { display: flex; justify-content: space-between; font-size: 0.75rem; color: #888; margin-bottom: 4px; }
  .pct { color: #e8488a; font-weight: 600; }
  .progress-track { height: 3px; background: #1a1a1a; border-radius: 2px; overflow: hidden; }
  .progress-fill {
    height: 100%; background: linear-gradient(90deg,#e8488a,#ff8ab4);
    border-radius: 2px; transition: width .2s ease;
  }

  /* Drop zone */
  .list-wrap { flex: 1; overflow: hidden; position: relative; }
  .list { height: 100%; overflow-y: auto; }
  .dragging .list { opacity: .25; pointer-events: none; }
  .drop-overlay {
    position: absolute; inset: 0; z-index: 10;
    background: #e8488a08; border: 2px dashed #e8488a55;
    display: flex; flex-direction: column; align-items: center;
    justify-content: center; gap: 6px; pointer-events: none;
  }
  .drop-overlay span { font-size: 2rem; color: #e8488a; }
  .drop-overlay p { color: #e8488a; font-size: 0.9rem; }
  .drop-overlay small { color: #e8488a55; font-size: 0.72rem; }

  /* File rows */
  .row { display: flex; align-items: center; border-bottom: 1px solid #0d0d0d; }
  .row:hover { background: #0c0c0c; }
  .row-main {
    flex: 1; display: flex; align-items: center; gap: 8px;
    padding: 8px 14px; background: none; border: none;
    color: #bbb; cursor: pointer; text-align: left; overflow: hidden;
  }
  .ico { flex-shrink: 0; font-size: 0.85rem; }
  .name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.82rem; }
  .meta { font-size: 0.68rem; color: #3a3a3a; white-space: nowrap; padding: 0 6px; }
  .date { font-size: 0.65rem; color: #2a2a2a; white-space: nowrap; }
  .row-btns { display: flex; gap: 1px; padding-right: 8px; }
  .rb {
    background: none; border: none; color: #2a2a2a; cursor: pointer;
    padding: 4px 6px; border-radius: 4px; font-size: 0.78rem; transition: all .12s;
  }
  .rb:hover { background: #111; color: #bbb; }
  .rb.del:hover { color: #ff6b6b; }
  .state { color: #333; padding: 16px; text-align: center; font-size: 0.82rem; }

  .toast {
    position: fixed; bottom: 20px; left: 50%; transform: translateX(-50%);
    background: #111; border: 1px solid #e8488a33; color: #e8488a;
    padding: 8px 18px; border-radius: 20px; font-size: 0.8rem; z-index: 100;
    animation: up .2s ease;
  }
  @keyframes up { from{opacity:0;transform:translateX(-50%) translateY(8px)} }
</style>
