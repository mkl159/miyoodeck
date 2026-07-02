<script>
  import { createEventDispatcher } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  export let pinConfigured = false

  const dispatch = createEventDispatcher()
  let pin = ''
  let newPin = ''
  let confirmPin = ''
  let error = ''
  let loading = false
  let setupMode = !pinConfigured

  async function login() {
    error = ''
    loading = true
    try {
      const res = await api.login(pin)
      dispatch('login', res.token)
    } catch (e) {
      // Show the server's message for lockouts ("Too many attempts…"),
      // fall back to the generic wrong-PIN text otherwise.
      error = e.message && e.message !== 'Unauthorized' ? e.message : $t.wrongPin
    }
    loading = false
  }

  // PINs are numeric — silently drop anything else as the user types.
  function digitsOnly(e) {
    const clean = e.target.value.replace(/\D/g, '')
    if (clean !== e.target.value) e.target.value = clean
    return clean
  }

  async function setupPin() {
    error = ''
    if (newPin.length < 4) { error = $t.pinTooShort; return }
    if (newPin !== confirmPin) { error = $t.pinMismatch; return }
    loading = true
    try {
      await api.setupPin(newPin)
      pin = newPin
      pinConfigured = true
      setupMode = false
      await login()
    } catch (e) {
      error = e.message
    }
    loading = false
  }

  function handleKeydown(e) {
    if (e.key === 'Enter') setupMode ? setupPin() : login()
  }
</script>

<div class="login-wrap">
  <div class="login-card">
    <div class="logo">
      <div class="ring r1"></div>
      <div class="ring r2"></div>
      <div class="ring r3"></div>
    </div>

    <h1>MiyooDeck</h1>

    {#if setupMode}
      <p class="subtitle">{$t.setupSubtitle}</p>
      <input
        type="password" inputmode="numeric" autocomplete="new-password"
        placeholder={$t.setupNewPin}
        bind:value={newPin}
        on:input={(e) => newPin = digitsOnly(e)}
        on:keydown={handleKeydown}
        maxlength="8"
      />
      <input
        type="password" inputmode="numeric" autocomplete="new-password"
        placeholder={$t.setupConfirm}
        bind:value={confirmPin}
        on:input={(e) => confirmPin = digitsOnly(e)}
        on:keydown={handleKeydown}
        maxlength="8"
      />
      <button on:click={setupPin} disabled={loading}>
        {loading ? $t.loginConnecting : $t.setupButton}
      </button>
      <button class="skip" on:click={() => { newPin=''; setupMode=false; pinConfigured=false }}>
        {$t.setupSkip}
      </button>
    {:else}
      <p class="subtitle">{$t.loginSubtitle}</p>
      <input
        type="password" inputmode="numeric" autocomplete="current-password"
        placeholder={$t.loginPin}
        bind:value={pin}
        on:input={(e) => pin = digitsOnly(e)}
        on:keydown={handleKeydown}
        maxlength="8"
        autofocus
      />
      <button on:click={login} disabled={loading}>
        {loading ? $t.loginConnecting : $t.loginButton}
      </button>
    {/if}

    {#if error}
      <p class="error">{error}</p>
    {/if}
  </div>
</div>

<style>
  .login-wrap {
    display: flex; align-items: center; justify-content: center;
    min-height: 100vh; background: #0a0a0a;
  }
  .login-card {
    display: flex; flex-direction: column; align-items: center;
    gap: 14px; padding: 40px 32px;
    background: #0d0d0d; border: 1px solid #e8488a1a;
    border-radius: 20px; width: 320px;
    box-shadow: 0 0 60px #e8488a08;
  }
  .logo { position: relative; width: 64px; height: 64px; margin-bottom: 4px; }
  .ring {
    position: absolute; border-radius: 50%;
    top: 50%; left: 50%; transform: translate(-50%,-50%);
    border: 3px solid;
  }
  .r1 { width: 64px; height: 64px; border-color: #e8488a; }
  .r2 { width: 42px; height: 42px; border-color: #e8488a55; }
  .r3 { width: 20px; height: 20px; border-color: #e8488a22; }
  h1 { font-size: 1.6rem; color: #e8488a; font-weight: 800; letter-spacing: 2px; }
  .subtitle { font-size: 0.82rem; color: #555; text-align: center; }
  input {
    width: 100%; padding: 12px 16px;
    background: #111; border: 1px solid #222;
    border-radius: 10px; color: #e0e0e0;
    font-size: 1.1rem; outline: none;
    transition: border-color 0.2s;
    text-align: center; letter-spacing: 6px;
  }
  input:focus { border-color: #e8488a55; }
  input::placeholder { letter-spacing: 0; font-size: 0.85rem; color: #444; }
  button {
    width: 100%; padding: 12px;
    background: linear-gradient(135deg, #e8488a, #c42d6e);
    border: none; border-radius: 10px; color: #fff;
    font-size: 0.95rem; font-weight: 700;
    cursor: pointer; transition: all 0.2s;
    letter-spacing: 0.5px;
  }
  button:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 20px #e8488a44; }
  button:disabled { opacity: 0.5; cursor: default; transform: none; }
  button.skip {
    background: #111; border: 1px solid #222;
    color: #555; font-weight: 400; font-size: 0.8rem;
  }
  button.skip:hover { border-color: #444; color: #888; transform: none; box-shadow: none; }
  .error { color: #ff6b6b; font-size: 0.8rem; text-align: center; }
</style>
