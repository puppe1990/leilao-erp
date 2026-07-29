<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let sales = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  async function cancelSale(id) {
    const ok = await askConfirm({
      title: 'Cancelar venda',
      message: 'Cancelar esta venda pendente?',
      detail: 'Todos os itens da composição voltam ao estoque.',
      confirmLabel: 'Cancelar venda',
      tone: 'warning',
      icon: 'cancel',
    })
    if (!ok) return
    router.post(`/sales/${id}/cancel`)
  }

  function paymentBadge(label, canCancel) {
    if (canCancel) return 'ahq-badge-pending'
    const l = (label || '').toLowerCase()
    if (l.includes('receb') || l.includes('pago')) return 'ahq-badge-live'
    if (l.includes('cancel')) return 'ahq-badge-error'
    return 'ahq-badge-sold'
  }
</script>

<AppShell {companyName} active="sales">
  <div class="flex items-start justify-between gap-3 mb-section-padding">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Vendas</h1>
      <p class="text-on-surface-variant text-body-md mt-1">Canais direto e marketplace.</p>
    </div>
    <a href="/sales/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm shrink-0">
      <span class="material-symbols-outlined text-[18px] mr-1">add</span>
      Nova
    </a>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if sales.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">sell</span>
      <p class="text-on-surface-variant mb-4">Nenhuma venda registrada.</p>
      <a href="/sales/new" use:inertia class="ahq-btn-primary">Registrar primeira venda</a>
    </div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each sales as sale}
        <div class="ahq-card p-4">
          <div class="flex justify-between items-start gap-2 mb-2">
            <div>
              <a href={`/sales/${sale.id}`} use:inertia class="font-semibold text-primary hover:underline">
                {sale.itemTitle || `#${sale.itemId}`}
              </a>
              <p class="text-on-surface-variant text-sm">
                {sale.soldAt?.slice?.(0, 10) || sale.soldAt} · {sale.channelLabel}
              </p>
            </div>
            <span class={paymentBadge(sale.paymentLabel, sale.canCancel)}>{sale.paymentLabel}</span>
          </div>
          <div class="bg-surface-container-low rounded p-3 grid grid-cols-3 gap-2 text-center">
            <div>
              <span class="ahq-label text-[10px]">Bruto</span>
              <p class="font-mono text-sm font-semibold">{sale.gross}</p>
            </div>
            <div>
              <span class="ahq-label text-[10px]">Taxa</span>
              <p class="font-mono text-sm">{sale.fee}</p>
            </div>
            <div>
              <span class="ahq-label text-[10px]">Líquido</span>
              <p class="font-mono text-sm font-semibold text-secondary">{sale.net}</p>
            </div>
          </div>
          {#if sale.canCancel}
            <div class="mt-3 flex gap-3">
              <a href={`/sales/${sale.id}/edit`} use:inertia class="text-sm text-secondary font-medium">Editar</a>
              <button type="button" class="text-sm text-error font-medium" on:click={() => cancelSale(sale.id)}>
                Excluir
              </button>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</AppShell>
