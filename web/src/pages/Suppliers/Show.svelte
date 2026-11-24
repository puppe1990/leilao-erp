<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let supplier = {}
  export let messages = []
  export let companyName = 'AuctionHQ'

  $: rows = Array.isArray(messages) ? messages : []
  $: name = supplier?.name || 'Fornecedor'

  async function destroy() {
    const ok = await askConfirm({
      title: 'Descartar fornecedor',
      message: `Descartar “${name}”?`,
      detail: 'A conversa salva também some. Essa ação não pode ser desfeita.',
      confirmLabel: 'Descartar',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/suppliers/${supplier.id}/delete`)
  }
</script>

<AppShell {companyName} active="suppliers">
  <div class="mb-4">
    <a href="/suppliers" use:inertia class="text-sm text-secondary font-medium hover:underline">← Fornecedores</a>
  </div>

  <div class="mb-section-padding flex flex-wrap items-start justify-between gap-3">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{name}</h1>
      <p class="text-on-surface-variant text-body-md mt-1 font-mono text-sm">{supplier.whatsappLabel}</p>
      {#if supplier.notes}
        <p class="text-on-surface-variant text-sm mt-1">{supplier.notes}</p>
      {/if}
      <p class="text-on-surface-variant text-sm mt-1">Pedido mínimo {supplier.minOrder}</p>
    </div>
    <button
      type="button"
      class="ahq-btn-ghost h-8 px-3 text-error border-error hover:border-error hover:text-error"
      on:click={destroy}
    >
      Descartar
    </button>
  </div>

  <section class="ahq-card p-4">
    <h2 class="font-semibold text-primary mb-3">Conversa</h2>
    {#if rows.length === 0}
      <p class="text-on-surface-variant text-sm">Nenhuma mensagem salva ainda.</p>
    {:else}
      <div class="grid gap-2">
        {#each rows as msg (msg.id)}
          <div class="flex {msg.fromMe ? 'justify-end' : 'justify-start'}">
            <div
              class="max-w-[85%] rounded-2xl px-3 py-2 {msg.fromMe
                ? 'bg-secondary text-on-secondary'
                : 'bg-surface-container-high text-on-surface'}"
            >
              <p class="whitespace-pre-wrap text-sm">{msg.body}</p>
              {#if msg.sentAt}
                <p class="text-[10px] mt-1 opacity-70 {msg.fromMe ? 'text-right' : ''}">{msg.sentAt}</p>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
</AppShell>
