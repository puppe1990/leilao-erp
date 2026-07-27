<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let items = []
  export let groups = []
  export let summary = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  function saveItem(item) {
    router.post(`/lots/${item.lotId}/items/${item.id}`, {
      title: item._title ?? item.title,
      sku: item._sku ?? item.sku ?? '',
      sale_price_hint: item._sale ?? item.salePriceRaw ?? '',
      return_to: '/stock',
    })
  }
</script>

<AppShell {companyName} active="stock">
  <div class="flex items-start justify-between gap-3 mb-section-padding">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Estoque</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Custo, preço sugerido de venda e margem potencial.
      </p>
    </div>
    <a href="/sales/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm shrink-0">
      <span class="material-symbols-outlined text-[18px] mr-1">sell</span>
      Vender
    </a>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="grid grid-cols-2 md:grid-cols-4 gap-stack-gap mb-section-padding">
    <div class="ahq-card p-4">
      <span class="ahq-label">Itens</span>
      <p class="ahq-value text-primary">{summary.count ?? 0}</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Custo estoque</span>
      <p class="ahq-value font-mono text-sm">{summary.totalCost || 'R$ 0,00'}</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Bruto potencial</span>
      <p class="ahq-value font-mono text-sm text-secondary">{summary.potentialGross || 'R$ 0,00'}</p>
      <p class="text-[10px] text-on-surface-variant mt-1">{summary.pricedCount || 0} com preço</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Margem potencial</span>
      <p class="ahq-value font-mono text-sm text-secondary">{summary.potentialMargin || 'R$ 0,00'}</p>
    </div>
  </section>

  {#if groups.length > 0}
    <section class="mb-section-padding">
      <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Por modelo</h2>
      <div class="flex flex-col gap-stack-gap">
        {#each groups as g}
          <div class="ahq-card p-4">
            <div class="flex justify-between items-start gap-2 mb-2">
              <div>
                <p class="font-semibold text-primary">{g.title}</p>
                <p class="text-on-surface-variant text-sm">{g.count} un. · custo {g.unitCost}</p>
              </div>
              {#if g.isAccessory}
                <span class="ahq-badge-sold text-[10px]">Acessório</span>
              {:else}
                <span class="ahq-badge-live text-[10px]">Principal</span>
              {/if}
            </div>
            <div class="bg-surface-container-low rounded p-3 grid grid-cols-3 gap-2 text-center">
              <div>
                <span class="ahq-label text-[10px]">Preço venda</span>
                <p class="font-mono text-sm font-semibold">{g.salePriceHint}</p>
              </div>
              <div>
                <span class="ahq-label text-[10px]">Bruto se vender tudo</span>
                <p class="font-mono text-sm">{g.potentialGross}</p>
              </div>
              <div>
                <span class="ahq-label text-[10px]">Margem pot.</span>
                <p class="font-mono text-sm font-semibold text-secondary">{g.potentialMargin}</p>
              </div>
            </div>
          </div>
        {/each}
      </div>
    </section>
  {/if}

  {#if items.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">inventory_2</span>
      <p class="text-on-surface-variant mb-4">Nenhum item em estoque.</p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar lote</a>
    </div>
  {:else}
    <section>
      <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Unidades</h2>
      <div class="flex flex-col gap-stack-gap">
        {#each items as item}
          <div class="ahq-card p-4 space-y-3">
            <div class="flex justify-between items-start gap-2">
              <div>
                <p class="font-semibold text-primary">{item.title}</p>
                <p class="text-[10px] font-mono text-on-surface-variant uppercase">
                  #{item.id} · lote {item.lotId}{item.sku ? ` · ${item.sku}` : ''}
                </p>
              </div>
              <div class="text-right">
                <p class="text-[10px] text-on-surface-variant uppercase">Custo</p>
                <p class="font-mono font-semibold text-sm">{item.unitCost}</p>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2 text-sm">
              <div class="bg-surface-container-low rounded p-2">
                <span class="ahq-label text-[10px]">Preço sugerido</span>
                <p class="font-mono font-semibold">{item.salePriceHint || '—'}</p>
              </div>
              <div class="bg-surface-container-low rounded p-2">
                <span class="ahq-label text-[10px]">Margem se vender</span>
                <p class="font-mono font-semibold text-secondary">{item.marginHint || '—'}</p>
              </div>
            </div>
            <form
              class="grid grid-cols-2 gap-2"
              on:submit|preventDefault={() => saveItem(item)}
            >
              <input
                class="ahq-input h-10 text-sm col-span-2"
                placeholder="Título"
                value={item.title}
                on:input={(e) => (item._title = e.currentTarget.value)}
              />
              <input
                class="ahq-input h-10 text-sm font-mono"
                placeholder="Preço venda (ex: 399,00)"
                value={item.salePriceRaw || ''}
                on:input={(e) => (item._sale = e.currentTarget.value)}
              />
              <input
                class="ahq-input h-10 text-sm font-mono"
                placeholder="SKU"
                value={item.sku || ''}
                on:input={(e) => (item._sku = e.currentTarget.value)}
              />
              <div class="col-span-2 flex gap-2">
                <button type="submit" class="ahq-btn-ghost h-10 text-sm flex-1">Salvar</button>
                <a
                  href={`/sales/new?item_id=${item.id}`}
                  use:inertia
                  class="ahq-btn-primary h-10 text-sm flex-1 text-center leading-10"
                >
                  Vender
                </a>
              </div>
            </form>
          </div>
        {/each}
      </div>
    </section>
  {/if}
</AppShell>
