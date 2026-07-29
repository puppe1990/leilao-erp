<script>
  import { onDestroy, onMount, tick } from 'svelte'
  import { writable } from 'svelte/store'
  import { closeConfirm, subscribeConfirm } from '@/lib/confirmDialog.js'

  /** local store so UI always reacts when bus fires */
  const openReq = writable(null)

  /** @type {HTMLElement | null} */
  let panelEl = null
  /** @type {(() => void) | null} */
  let unsub = null

  onMount(() => {
    unsub = subscribeConfirm((req) => {
      openReq.set(req)
      if (req) {
        tick().then(() => {
          const btn = panelEl?.querySelector('[data-confirm-cancel]')
          if (btn instanceof HTMLElement) btn.focus()
        })
      }
    })
  })

  onDestroy(() => {
    if (unsub) unsub()
    unsub = null
  })

  /** Append dialog to document.body so fixed stacking always works. */
  function portal(node) {
    document.body.appendChild(node)
    return {
      destroy() {
        if (node.parentNode) node.parentNode.removeChild(node)
      },
    }
  }

  function onKeydown(e) {
    if (!$openReq) return
    if (e.key === 'Escape') {
      e.preventDefault()
      closeConfirm(false)
    }
  }

  $: state = $openReq
  $: tone = state?.tone || 'danger'
  $: icon = state?.icon || 'delete'
  $: confirmBtnClass =
    tone === 'primary'
      ? 'confirm-btn confirm-btn-primary'
      : tone === 'warning'
        ? 'confirm-btn confirm-btn-warning'
        : 'confirm-btn confirm-btn-danger'
  $: iconWrapClass =
    tone === 'primary'
      ? 'confirm-icon confirm-icon-primary'
      : tone === 'warning'
        ? 'confirm-icon confirm-icon-warning'
        : 'confirm-icon confirm-icon-danger'
</script>

<svelte:window on:keydown={onKeydown} />

{#if state}
  <div use:portal class="confirm-root" role="presentation">
    <button
      type="button"
      class="confirm-backdrop"
      aria-label="Fechar diálogo"
      on:click={() => closeConfirm(false)}
    ></button>

    <div
      bind:this={panelEl}
      class="confirm-panel"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-desc"
      tabindex="-1"
    >
      <div class="confirm-body">
        <div class="confirm-row">
          <div class={iconWrapClass} aria-hidden="true">
            <span class="material-symbols-outlined">{icon}</span>
          </div>
          <div class="confirm-copy">
            <h2 id="confirm-title">{state.title}</h2>
            <p id="confirm-desc">{state.message}</p>
            {#if state.detail}
              <p class="confirm-detail">{state.detail}</p>
            {/if}
          </div>
        </div>

        <div class="confirm-actions">
          <button
            type="button"
            data-confirm-cancel
            class="confirm-btn confirm-btn-ghost"
            on:click={() => closeConfirm(false)}
          >
            {state.cancelLabel}
          </button>
          <button type="button" class={confirmBtnClass} on:click={() => closeConfirm(true)}>
            {state.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .confirm-root {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding: 1rem;
    box-sizing: border-box;
    animation: confirm-fade 0.15s ease-out;
  }
  @media (min-width: 640px) {
    .confirm-root {
      align-items: center;
      padding: 1.5rem;
    }
  }
  .confirm-backdrop {
    position: absolute;
    inset: 0;
    border: 0;
    padding: 0;
    margin: 0;
    cursor: default;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(3px);
    -webkit-backdrop-filter: blur(3px);
  }
  .confirm-panel {
    position: relative;
    z-index: 1;
    width: 100%;
    max-width: 28rem;
    border-radius: 1rem;
    border: 1px solid rgb(var(--c-border) / 1);
    background: rgb(var(--c-panel) / 1);
    color: rgb(var(--c-text) / 1);
    box-shadow: var(--c-shadow);
    overflow: hidden;
    animation: confirm-pop 0.18s cubic-bezier(0.16, 1, 0.3, 1);
    outline: none;
  }
  .confirm-body {
    padding: 1.5rem;
  }
  .confirm-row {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
  }
  .confirm-icon {
    width: 3rem;
    height: 3rem;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }
  .confirm-icon .material-symbols-outlined {
    font-size: 26px;
  }
  .confirm-icon-danger {
    background: rgb(var(--c-error-bg) / 1);
    color: rgb(var(--c-error-fg) / 1);
  }
  .confirm-icon-warning {
    background: rgb(var(--c-pending) / 0.15);
    color: rgb(var(--c-pending) / 1);
  }
  .confirm-icon-primary {
    background: rgb(var(--c-green) / 0.15);
    color: rgb(var(--c-green) / 1);
  }
  .confirm-copy {
    min-width: 0;
    flex: 1;
    padding-top: 0.125rem;
  }
  .confirm-copy h2 {
    margin: 0;
    font-size: 1.25rem;
    line-height: 1.75rem;
    font-weight: 900;
    letter-spacing: -0.01em;
    color: rgb(var(--c-text) / 1);
  }
  .confirm-copy p {
    margin: 0.375rem 0 0;
    font-size: 0.875rem;
    line-height: 1.4rem;
    color: rgb(var(--c-muted) / 1);
  }
  .confirm-detail {
    margin-top: 0.5rem !important;
    font-size: 0.75rem !important;
    border-left: 2px solid rgb(var(--c-border) / 1);
    padding-left: 0.75rem;
  }
  .confirm-actions {
    margin-top: 1.5rem;
    display: flex;
    flex-direction: column-reverse;
    gap: 0.625rem;
  }
  @media (min-width: 640px) {
    .confirm-actions {
      flex-direction: row;
      justify-content: flex-end;
    }
  }
  .confirm-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 3rem;
    padding: 0 1.25rem;
    border-radius: 9999px;
    font-weight: 700;
    font-size: 11px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    border: 1px solid transparent;
    cursor: pointer;
    transition:
      opacity 0.15s,
      transform 0.1s,
      border-color 0.15s;
    width: 100%;
  }
  @media (min-width: 640px) {
    .confirm-btn {
      width: auto;
    }
  }
  .confirm-btn:active {
    transform: scale(0.98);
  }
  .confirm-btn-ghost {
    background: transparent;
    color: rgb(var(--c-text) / 1);
    border-color: rgb(var(--c-border) / 1);
  }
  .confirm-btn-ghost:hover {
    border-color: rgb(var(--c-green) / 1);
    color: rgb(var(--c-green) / 1);
  }
  .confirm-btn-danger {
    background: rgb(var(--c-error) / 1);
    color: rgb(var(--c-on-error) / 1);
  }
  .confirm-btn-danger:hover {
    opacity: 0.9;
  }
  .confirm-btn-warning {
    background: rgb(var(--c-pending) / 1);
    color: #000;
  }
  .confirm-btn-warning:hover {
    opacity: 0.9;
  }
  .confirm-btn-primary {
    background: rgb(var(--c-green) / 1);
    color: rgb(var(--c-on-green) / 1);
  }
  .confirm-btn-primary:hover {
    opacity: 0.9;
  }
  @keyframes confirm-fade {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes confirm-pop {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
