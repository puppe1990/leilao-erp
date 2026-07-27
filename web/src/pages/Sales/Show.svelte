<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let sale = {}
  export let errors = {}
  export let companyName = 'AuctionHQ'
  export let site = {}

  function destroy() {
    if (!confirm('Excluir/cancelar esta venda pendente?')) return
    router.post(`/sales/${sale.id}/delete`)
  }
</script>

<AppShell {companyName} active="sales">
  <div class="mb-section-padding">
    <a href="/sales" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Vendas
    </a>
    <div class="flex items-start justify-between gap-3">
      <div>
        <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{sale.itemTitle || `Venda #${sale.id}`}</h1>
        <p class="text-on-surface-variant text-body-md mt-1">
          {sale.soldAt?.slice?.(0, 10) || sale.soldAt} · {sale.channelLabel}
        </p>
      </div>
      <span class="ahq-badge-pending">{sale.paymentLabel}</span>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <div class="ahq-card p-5 mb-section-padding grid grid-cols-2 gap-4">
    <div>
      <span class="ahq-label">Bruto</span>
      <p class="ahq-value">{sale.gross}</p>
    </div>
    <div>
      <span class="ahq-label">Taxa</span>
      <p class="ahq-value">{sale.fee}</p>
    </div>
    <div>
      <span class="ahq-label">Frete</span>
      <p class="ahq-value">{sale.shipping}</p>
    </div>
    <div>
      <span class="ahq-label">Líquido</span>
      <p class="ahq-value text-secondary">{sale.net}</p>
    </div>
    <div>
      <span class="ahq-label">Custo total (composição)</span>
      <p class="font-mono font-semibold">{sale.unitCost}</p>
    </div>
    <div>
      <span class="ahq-label">Margem</span>
      <p class="font-mono font-semibold text-secondary">{sale.margin || '—'}</p>
    </div>
  </div>

  {#if sale.lines?.length}
    <div class="ahq-card p-5 mb-section-padding">
      <h2 class="font-semibold text-primary mb-3">Itens da venda</h2>
      <ul class="divide-y divide-outline-variant">
        {#each sale.lines as line}
          <li class="py-2 flex justify-between gap-3 text-sm">
            <div>
              <span class="font-medium">{line.title}</span>
              <span class="text-on-surface-variant ml-2 text-xs">{line.roleLabel}</span>
            </div>
            <span class="font-mono shrink-0">{line.unitCost}</span>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <div class="flex flex-wrap gap-3">
    {#if sale.canEdit}
      <a href={`/sales/${sale.id}/edit`} use:inertia class="ahq-btn-primary">Editar</a>
    {/if}
    {#if sale.canDelete}
      <button type="button" class="ahq-btn-ghost text-error border-error" on:click={destroy}>Excluir</button>
    {/if}
    <a href="/sales" use:inertia class="ahq-btn-ghost">Voltar</a>
  </div>
</AppShell>
