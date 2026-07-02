<script>
  import { onMount, onDestroy } from 'svelte'
  import { api } from '../api.js'
  import { t } from '../i18n.js'

  // Track which buttons are currently held (for visual feedback)
  let held = {}

  function pointerDown(btn, e) {
    e.preventDefault()
    held[btn] = true
    held = held
    api.pressButton(btn, 'press').catch(() => {})
  }

  function pointerUp(btn, e) {
    e.preventDefault()
    if (!held[btn]) return
    delete held[btn]
    held = held
    api.pressButton(btn, 'release').catch(() => {})
  }

  // ── Keyboard control ──────────────────────────────────────────────
  // Play with a real keyboard instead of clicking on-screen buttons.
  let kbdEnabled = true
  const keyMap = {
    ArrowUp: 'up', ArrowDown: 'down', ArrowLeft: 'left', ArrowRight: 'right',
    x: 'a', z: 'b', s: 'x', a: 'y',
    q: 'l1', e: 'r1', w: 'l2', r: 'r2',
    Enter: 'start', Shift: 'select', m: 'menu',
  }

  function keyToBtn(e) {
    return keyMap[e.key] ?? keyMap[e.key.toLowerCase()]
  }

  // Never steal keys while the user is typing somewhere.
  function isTyping(e) {
    const tag = e.target?.tagName
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || e.target?.isContentEditable
  }

  function onKeyDown(e) {
    if (!kbdEnabled || e.repeat || isTyping(e) || e.ctrlKey || e.metaKey || e.altKey) return
    const btn = keyToBtn(e)
    if (!btn || held[btn]) return
    e.preventDefault()
    held[btn] = true
    held = held
    api.pressButton(btn, 'press').catch(() => {})
  }

  function onKeyUp(e) {
    if (isTyping(e)) return
    const btn = keyToBtn(e)
    if (!btn || !held[btn]) return
    e.preventDefault()
    delete held[btn]
    held = held
    api.pressButton(btn, 'release').catch(() => {})
  }

  onMount(() => {
    window.addEventListener('keydown', onKeyDown)
    window.addEventListener('keyup', onKeyUp)
  })
  onDestroy(() => {
    window.removeEventListener('keydown', onKeyDown)
    window.removeEventListener('keyup', onKeyUp)
  })

  // For buttons that shouldn't be held (menu, volume)
  function tap(btn, e) {
    e.preventDefault()
    api.pressButton(btn, 'tap').catch(() => {})
  }

  // Server-side macro: the Konami code, played as a tap sequence.
  let konamiPlaying = false
  function konami() {
    const seq = ['up', 'up', 'down', 'down', 'left', 'right', 'left', 'right', 'b', 'a']
    konamiPlaying = true
    api.macro(seq.map(b => ({ button: b, action: 'tap', delay_ms: 120 })))
      .catch(() => {})
      .finally(() => setTimeout(() => konamiPlaying = false, seq.length * 130))
  }

  function mkBtn(btn) {
    return {
      onpointerdown: (e) => pointerDown(btn, e),
      onpointerup:   (e) => pointerUp(btn, e),
      onpointerleave:(e) => pointerUp(btn, e),
    }
  }
</script>

<div class="controller">
  <!-- Shoulder buttons -->
  <div class="shoulders">
    <div class="shoulder-group left">
      <button class="btn shoulder" class:active={held['l1']}
        on:pointerdown={(e) => pointerDown('l1', e)}
        on:pointerup={(e) => pointerUp('l1', e)}
        on:pointerleave={(e) => pointerUp('l1', e)}>L1</button>
      <button class="btn shoulder" class:active={held['l2']}
        on:pointerdown={(e) => pointerDown('l2', e)}
        on:pointerup={(e) => pointerUp('l2', e)}
        on:pointerleave={(e) => pointerUp('l2', e)}>L2</button>
    </div>
    <div class="shoulder-group center">
      <button class="btn menu-btn" on:pointerdown|preventDefault={(e) => tap('menu', e)}>MENU</button>
    </div>
    <div class="shoulder-group right">
      <button class="btn shoulder" class:active={held['r1']}
        on:pointerdown={(e) => pointerDown('r1', e)}
        on:pointerup={(e) => pointerUp('r1', e)}
        on:pointerleave={(e) => pointerUp('r1', e)}>R1</button>
      <button class="btn shoulder" class:active={held['r2']}
        on:pointerdown={(e) => pointerDown('r2', e)}
        on:pointerup={(e) => pointerUp('r2', e)}
        on:pointerleave={(e) => pointerUp('r2', e)}>R2</button>
    </div>
  </div>

  <!-- Main controls -->
  <div class="main-row">
    <!-- D-Pad -->
    <div class="dpad">
      <button class="btn dpad-btn up" class:active={held['up']}
        on:pointerdown={(e) => pointerDown('up', e)}
        on:pointerup={(e) => pointerUp('up', e)}
        on:pointerleave={(e) => pointerUp('up', e)}>▲</button>
      <div class="dpad-mid">
        <button class="btn dpad-btn left" class:active={held['left']}
          on:pointerdown={(e) => pointerDown('left', e)}
          on:pointerup={(e) => pointerUp('left', e)}
          on:pointerleave={(e) => pointerUp('left', e)}>◀</button>
        <div class="dpad-center"></div>
        <button class="btn dpad-btn right" class:active={held['right']}
          on:pointerdown={(e) => pointerDown('right', e)}
          on:pointerup={(e) => pointerUp('right', e)}
          on:pointerleave={(e) => pointerUp('right', e)}>▶</button>
      </div>
      <button class="btn dpad-btn down" class:active={held['down']}
        on:pointerdown={(e) => pointerDown('down', e)}
        on:pointerup={(e) => pointerUp('down', e)}
        on:pointerleave={(e) => pointerUp('down', e)}>▼</button>
    </div>

    <!-- Select / Start -->
    <div class="mid-btns">
      <button class="btn mid-btn"
        on:pointerdown={(e) => pointerDown('select', e)}
        on:pointerup={(e) => pointerUp('select', e)}
        on:pointerleave={(e) => pointerUp('select', e)}
        class:active={held['select']}>SELECT</button>
      <button class="btn mid-btn"
        on:pointerdown={(e) => pointerDown('start', e)}
        on:pointerup={(e) => pointerUp('start', e)}
        on:pointerleave={(e) => pointerUp('start', e)}
        class:active={held['start']}>START</button>
    </div>

    <!-- Face buttons (ABXY) -->
    <div class="face">
      <button class="btn face-btn x" class:active={held['x']}
        on:pointerdown={(e) => pointerDown('x', e)}
        on:pointerup={(e) => pointerUp('x', e)}
        on:pointerleave={(e) => pointerUp('x', e)}>X</button>
      <div class="face-mid">
        <button class="btn face-btn y" class:active={held['y']}
          on:pointerdown={(e) => pointerDown('y', e)}
          on:pointerup={(e) => pointerUp('y', e)}
          on:pointerleave={(e) => pointerUp('y', e)}>Y</button>
        <div class="face-center"></div>
        <button class="btn face-btn a" class:active={held['a']}
          on:pointerdown={(e) => pointerDown('a', e)}
          on:pointerup={(e) => pointerUp('a', e)}
          on:pointerleave={(e) => pointerUp('a', e)}>A</button>
      </div>
      <button class="btn face-btn b" class:active={held['b']}
        on:pointerdown={(e) => pointerDown('b', e)}
        on:pointerup={(e) => pointerUp('b', e)}
        on:pointerleave={(e) => pointerUp('b', e)}>B</button>
    </div>
  </div>

  <!-- Volume + macro + keyboard toggle -->
  <div class="vol-row">
    <button class="btn vol-btn" title="Volume −" on:pointerdown|preventDefault={(e) => tap('volume_dn', e)}>🔉</button>
    <button class="btn macro-btn" class:active={konamiPlaying} on:click={konami}>🎮 Konami</button>
    <button class="btn macro-btn" class:active={kbdEnabled} title={$t.kbdHint}
      on:click={() => kbdEnabled = !kbdEnabled}>{$t.kbdControl}</button>
    <button class="btn vol-btn" title="Volume +" on:pointerdown|preventDefault={(e) => tap('volume_up', e)}>🔊</button>
  </div>

  {#if kbdEnabled}
    <p class="kbd-hint">{$t.kbdHint}</p>
  {/if}
</div>

<style>
  .controller {
    background: #0d0d0d;
    border: 1px solid #1a1a1a;
    border-radius: 14px;
    padding: 12px 14px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    user-select: none;
    -webkit-user-select: none;
    touch-action: none;
  }

  .btn {
    background: #111;
    border: 1px solid #222;
    border-radius: 8px;
    color: #666;
    cursor: pointer;
    font-size: 0.72rem;
    font-weight: 700;
    transition: background .08s, color .08s, border-color .08s, transform .08s;
    display: flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
    padding: 0;
  }
  .btn:active, .btn.active {
    background: #e8488a22;
    border-color: #e8488a;
    color: #e8488a;
    transform: scale(0.93);
  }

  /* Shoulders */
  .shoulders {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 6px;
  }
  .shoulder-group { display: flex; gap: 5px; }
  .shoulder-group.center { flex: 1; justify-content: center; }
  .shoulder {
    width: 44px;
    height: 28px;
    border-radius: 6px;
    font-size: 0.65rem;
  }
  .menu-btn {
    width: 64px;
    height: 24px;
    border-radius: 12px;
    font-size: 0.6rem;
    letter-spacing: 1px;
  }

  /* Main row */
  .main-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  /* D-Pad */
  .dpad {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }
  .dpad-mid {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .dpad-btn {
    width: 42px;
    height: 42px;
    border-radius: 6px;
    font-size: 0.9rem;
  }
  .dpad-center {
    width: 42px;
    height: 42px;
    background: #0a0a0a;
    border-radius: 4px;
  }

  /* Mid buttons */
  .mid-btns {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: center;
  }
  .mid-btn {
    width: 64px;
    height: 26px;
    border-radius: 12px;
    font-size: 0.6rem;
    letter-spacing: 1px;
  }

  /* Face buttons */
  .face {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }
  .face-mid {
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .face-btn {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    font-size: 0.75rem;
  }
  .face-center {
    width: 42px;
    height: 42px;
  }
  .face-btn.a { border-color: #c8692233; color: #c86922; }
  .face-btn.a:active, .face-btn.a.active { background: #c8692222; border-color: #c86922; color: #c86922; }
  .face-btn.b { border-color: #e8488a33; color: #e8488a; }
  .face-btn.b:active, .face-btn.b.active { background: #e8488a22; border-color: #e8488a; color: #e8488a; }
  .face-btn.x { border-color: #6b9dff33; color: #6b9dff; }
  .face-btn.x:active, .face-btn.x.active { background: #6b9dff22; border-color: #6b9dff; color: #6b9dff; }
  .face-btn.y { border-color: #6bff9d33; color: #6bff9d; }
  .face-btn.y:active, .face-btn.y.active { background: #6bff9d22; border-color: #6bff9d; color: #6bff9d; }

  /* Volume */
  .vol-row {
    display: flex;
    justify-content: center;
    gap: 10px;
  }
  .vol-btn {
    width: 44px;
    height: 28px;
    font-size: 0.9rem;
    border-radius: 8px;
  }
  .macro-btn {
    height: 28px;
    padding: 0 12px;
    font-size: 0.62rem;
    letter-spacing: 0.5px;
    border-radius: 8px;
  }
  .kbd-hint {
    font-size: 0.6rem;
    color: #333;
    text-align: center;
    line-height: 1.5;
  }
</style>
