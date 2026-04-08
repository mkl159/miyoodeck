// API client - auto-detects base URL from current host
const BASE = window.location.origin

let token = localStorage.getItem('webdeck_token') || ''

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
  const res = await request(path, options)
  return res.json()
}

export const api = {
  // Auth
  authStatus: () => json('/api/auth/status'),
  login: (pin) => json('/api/auth/login', { method: 'POST', body: JSON.stringify({ pin }) }),
  logout: () => json('/api/auth/logout'),
  setupPin: (pin) => json('/api/auth/setup', { method: 'POST', body: JSON.stringify({ pin }) }),

  // System
  system: () => json('/api/system'),

  // Games
  systems: () => json('/api/systems'),
  roms: (system) => json(`/api/roms?system=${encodeURIComponent(system)}`),
  launch: (rom_path, system) => json('/api/launch', {
    method: 'POST',
    body: JSON.stringify({ rom_path, system })
  }),

  // Files
  files: (path) => json(`/api/files?path=${encodeURIComponent(path)}`),
  delete: (path) => json(`/api/delete?path=${encodeURIComponent(path)}`, { method: 'DELETE' }),
  upload: (files, destPath) => {
    const form = new FormData()
    form.append('path', destPath)
    for (const f of files) form.append('files', f)
    const headers = {}
    if (token) headers['Authorization'] = `Bearer ${token}`
    return fetch(BASE + '/api/upload', { method: 'POST', headers, body: form })
      .then(r => r.json())
  },
  unzip: (zip_path, dest_dir) => json('/api/unzip', {
    method: 'POST',
    body: JSON.stringify({ zip_path, dest_dir })
  }),
  savesBackupUrl: () => {
    const t = token ? `?token=${token}` : ''
    return BASE + '/api/saves/backup' + t
  },

  // Screenshot
  screenshotUrl: () => BASE + '/api/screenshot?' + Date.now(),

  // Config
  configList: () => json('/api/config/list'),
  configRead: (path) => json(`/api/config?path=${encodeURIComponent(path)}`),
  configWrite: (path, content) => json('/api/config', {
    method: 'POST',
    body: JSON.stringify({ path, content })
  }),
}

// WebSocket connection
export function createWS(onMessage) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws`
  const ws = new WebSocket(wsUrl)

  ws.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data)
      onMessage(msg)
    } catch {}
  }

  ws.onerror = () => {}
  ws.onclose = () => {
    // Reconnect after 3s
    setTimeout(() => createWS(onMessage), 3000)
  }

  return ws
}
