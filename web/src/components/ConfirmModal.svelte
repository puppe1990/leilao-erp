<script>
  import { confirmState, closeConfirm } from '@/lib/confirmDialog.js'

  /** @type {HTMLElement | null} */
  let panelEl = null

  $: open = !!$confirmState
  $: state = $confirmState

  $: if (open && panelEl) {
    queueMicrotask(() => {
      const btn = panelEl?.querySelector('[data-confirm-cancel]')
      if (btn instanceof HTMLElement) btn.focus()
    })
  }

  function onKeydown(e) {
    if (!open) return
    if (e.key === 'Escape') {
      e.preventDefault()
      closeConfirm(false)
    }
  }

  $: tone = state?.tone || 'danger'
  $: icon = state?.icon || 'delete'
  $: confirmClass =
    tone === 'primary'
      ? 'ahq-btn-primary'
      : tone === 'warning'
        ? 'inline-flex items-center justify-center h-12 px-5 bg-pending text-on-primary font-bold rounded-full uppercase tracking-wide text-[11px] hover:opacity-90 active:scale-[0.98] transition-all'
        : 'inline-flex items-center justify-center h-12 px-5 bg-error text-on-error font-bold rounded-full uppercase tracking-wide text-[11px] hover:opacity-90 active:scale-[0.98] transition-all'

  $: iconWrap =
    tone === 'primary'
      ? 'bg-secondary/15 text-secondary'
      : tone === 'warning'
        ? 'bg-pending/15 text-pending'
        : 'bg-error-container text-on-error-container'
</script>

<svelte:window on:keydown={onKeydown} />

{#if open && state}
  <div
    class="fixed inset-0 z-[200] flex items-end sm:items-center justify-center p-4 sm:p-6
      animate-[confirm-fade_0.15s_ease-out]"
  >
    <!-- Backdrop as real button for a11y -->
    <button
      type="button"
      class="absolute inset-0 bg-black/55 backdrop-blur-[3px] cursor-default border-0 p-0"
      aria-label="Fechar diálogo"
      on:click={() => closeConfirm(false)}
    ></button>

    <div
      bind:this={panelEl}
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-desc"
      tabindex="-1"
      class="relative z-10 w-full max-w-md rounded-2xl border border-outline-variant bg-surface-container-lowest
        shadow-float overflow-hidden animate-[confirm-pop_0.18s_cubic-bezier(0.16,1,0.3,1)]"
    >
      <div class="p-6 sm:p-7">
        <div class="flex gap-4 items-start">
          <div
            class="w-12 h-12 rounded-full flex items-center justify-center shrink-0 {iconWrap}"
            aria-hidden="true"
          >
            <span class="material-symbols-outlined text-[26px]">{icon}</span>
          </div>
          <div class="min-w-0 flex-1 pt-0.5">
            <h2
              id="confirm-title"
              class="font-headline-md text-headline-md font-black text-on-surface"
            >
              {state.title}
            </h2>
            <p
              id="confirm-desc"
              class="mt-1.5 text-body-md text-on-surface-variant leading-relaxed"
            >
              {state.message}
            </p>
            {#if state.detail}
              <p
                class="mt-2 text-xs text-on-surface-variant/90 leading-relaxed border-l-2 border-outline-variant pl-3"
              >
                {state.detail}
              </p>
            {/if}
          </div>
        </div>

        <div class="mt-6 flex flex-col-reverse sm:flex-row gap-2.5 sm:justify-end">
          <button
            type="button"
            data-confirm-cancel
            class="ahq-btn-ghost w-full sm:w-auto"
            on:click={() => closeConfirm(false)}
          >
            {state.cancelLabel}
          </button>
          <button
            type="button"
            class="{confirmClass} w-full sm:w-auto"
            on:click={() => closeConfirm(true)}
          >
            {state.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
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
