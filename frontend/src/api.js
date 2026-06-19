// MiyooDeck API client
const BASE = window.location.origin

let token = localStorage.getItem('webdeck_token') || ''
let wsInstance = null
let wsShouldReconnect = true // Fix #12: stop reconnecting after logout

export function setToken(t) {
  token = t
  if (t) localStorage.setItem('webdeck_token', t)
  else localStorage.removeItem('webdeck_token')
}

export function getToken() { return token }

async function request(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...options.headers }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(BASE + path, { ...options, headers })
  if (res.status === 401) {
    setToken('')
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  return res
}

async function json(path, options = {}) {
  return (await request(path, options)).json()
}

export const api = {
  // Auth
  authStatus: () => json('/api/auth/status'),
  login: (pin) => json('/api/auth/login', { method: 'POST', body: JSON.stringify({ pin }) }),
  logout: async () => {
    wsShouldReconnect = false // Fix #12: stop WS reconnect on logout
    if (wsInstance) { wsInstance.close(); wsInstance = null }
    return json('/api/auth/logout')
  },
  setupPin: (pin) => json('/api/auth/setup', { method: 'POST', body: JSON.stringify({ pin }) }),

  // System
  system: () => json('/api/system'),
  power: (action) => json('/api/system/power', {
    method: 'POST',
    body: JSON.stringify({ action }),
  }),
  brightness: () => json('/api/system/brightness'),
  setBrightness: (value) => json('/api/system/brightness', {
    method: 'POST',
    body: JSON.stringify({ value }),
  }),

  // Games
  systems: () => json('/api/systems'),
  roms: (system) => json(`/api/roms?system=${encodeURIComponent(system)}`),
  searchRoms: (q) => json(`/api/search?q=${encodeURIComponent(q)}`),
  randomRom: (system) => json(`/api/random${system ? `?system=${encodeURIComponent(system)}` : ''}`),
  launch: (rom_path, system) => json('/api/launch', {
    method: 'POST',
    body: JSON.stringify({ rom_path, system }),
  }),

  // Files
  files: (path) => json(`/api/files?path=${encodeURIComponent(path)}`),
  delete: (path) => json(`/api/delete?path=${encodeURIComponent(path)}`, { method: 'DELETE' }),

  // Fix #6: upload with progress callback
  upload: (files, destPath, onProgress) => new Promise((resolve, reject) => {
    const form = new FormData()
    form.append('path', destPath)
    for (const f of files) form.append('files', f)

    const xhr = new XMLHttpRequest()
    xhr.open('POST', BASE + '/api/upload')
    if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`)

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    }
    xhr.onload = () => {
      try { resolve(JSON.parse(xhr.responseText)) }
      catch { reject(new Error('Upload failed')) }
    }
    xhr.onerror = () => reject(new Error('Network error'))
    xhr.send(form)
  }),

  unzip: (zip_path, dest_dir) => json('/api/unzip', {
    method: 'POST',
    body: JSON.stringify({ zip_path, dest_dir }),
  }),

  // Fix #2: token passed as query param for direct browser download links
  savesBackupUrl: () => {
    const t = token ? `?token=${encodeURIComponent(token)}` : ''
    return `${BASE}/api/saves/backup${t}`
  },

  // Fix #10: individual file download
  downloadUrl: (path) => {
    const t = token ? `&token=${encodeURIComponent(token)}` : ''
    return `${BASE}/api/download?path=${encodeURIComponent(path)}${t}`
  },

  // Fix #8: screenshotUrl no longer embeds timestamp — caller handles cache-busting
  screenshotUrl: (ts) => {
    const t = token ? `&token=${encodeURIComponent(token)}` : ''
    return `${BASE}/api/screenshot?ts=${ts || Date.now()}${t}`
  },

  // Input (gamepad buttons)
  pressButton: (button, action = 'tap') => json('/api/input/press', {
    method: 'POST',
    body: JSON.stringify({ button, action }),
  }),

  // Config
  configList: () => json('/api/config/list'),
  configRead: (path) => json(`/api/config?path=${encodeURIComponent(path)}`),
  configWrite: (path, content) => json('/api/config', {
    method: 'POST',
    body: JSON.stringify({ path, content }),
  }),
}

// Fix #12: WebSocket with controlled reconnect lifecycle
export function createWS(onMessage, onStatusChange) {
  wsShouldReconnect = true

  function connect() {
    if (!wsShouldReconnect) return

    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${proto}//${location.host}/ws`
    const ws = new WebSocket(url)
    wsInstance = ws

    ws.onopen = () => onStatusChange?.(true)
    ws.onclose = () => {
      onStatusChange?.(false)
      wsInstance = null
      if (wsShouldReconnect) setTimeout(connect, 3000)
    }
    ws.onerror = () => {}
    ws.onmessage = (e) => {
      try { onMessage(JSON.parse(e.data)) } catch {}
    }
  }

  connect()
  return { stop: () => { wsShouldReconnect = false; wsInstance?.close() } }
}
