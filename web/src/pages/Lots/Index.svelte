<script>
  import { inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let lots = []
  export let site = {}
  export let companyName = 'AuctionHQ'

  function statusBadge(status) {
    const s = (status || '').toLowerCase()
    if (s.includes('open') || s.includes('aberto')) return 'ahq-badge-open'
    if (s.includes('partial') || s.includes('parcial')) return 'ahq-badge-pending'
    if (s.includes('sold') || s.includes('vend')) return 'ahq-badge-sold'
    return 'ahq-badge-sold'
  }
</script>

<AppShell {companyName} active="lots">
  <div class="flex items-start justify-between gap-3 mb-section-padding">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Gestão de Lotes</h1>
      <p class="text-on-surface-variant text-body-md mt-1">Compras de leilão e rateio de custos.</p>
    </div>
    <a href="/lots/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm shrink-0">
      <span class="material-symbols-outlined text-[18px] mr-1">add</span>
      Novo
    </a>
  </div>

  {#if lots.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">inventory_2</span>
      <p class="text-on-surface-variant mb-4">Nenhum lote cadastrado.</p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar primeira compra de leilão</a>
    </div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each lots as lot}
        <a
          href={`/lots/${lot.id}`}
          use:inertia
          class="ahq-card lot-card overflow-hidden active:scale-[0.99] transition-transform block"
        >
          <div class="p-4 flex gap-4">
            <div
              class="w-14 h-14 rounded-lg bg-surface-container flex items-center justify-center shrink-0"
            >
              <span class="material-symbols-outlined text-primary text-2xl">gavel</span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between gap-2">
                <h2 class="font-semibold text-body-lg text-primary truncate">{lot.name}</h2>
                <span class={statusBadge(lot.status || lot.statusLabel)}>{lot.statusLabel || lot.status}</span>
              </div>
              <p class="text-on-surface-variant text-body-md mt-0.5">
                {lot.purchasedAt} · {lot.itemCount}
                {lot.itemCount === 1 ? 'item' : 'itens'}
              </p>
              <div class="mt-3 bg-surface-container-low rounded p-3 flex justify-between items-end">
                <div>
                  <span class="ahq-label text-[10px]">Custo total</span>
                  <p class="ahq-value text-primary">{lot.totalCost}</p>
                </div>
                <span class="material-symbols-outlined text-on-surface-variant">chevron_right</span>
              </div>
            </div>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</AppShell>
