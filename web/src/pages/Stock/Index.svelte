<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import {
    brandsFromItems,
    filterAndSortStockItems,
  } from '@/lib/stockInventoryTable.js'

  export let items = []
  export let groups = []
  export let summary = {}
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let query = ''
  /** @type {'all'|'main'|'accessory'} */
  let filterType = 'all'
  /** @type {'all'|'priced'|'unpriced'} */
  let filterPrice = 'all'
  /** @type {'all'|'with'|'without'|'unset'} */
  let filterBase = 'all'
  let filterBrand = 'all'

  /** @type {string} */
  let sortKey = 'title'
  /** @type {'asc'|'desc'} */
  let sortDir = 'asc'

  let editingId = null
  let editTitle = ''
  let editSale = ''

  $: brands = brandsFromItems(items)
  $: typeOptions = [
    { value: 'all', label: 'Todos' },
    { value: 'main', label: 'Principal' },
    { value: 'accessory', label: 'Acessório' },
  ]
  $: priceOptions = [
    { value: 'all', label: 'Todos' },
    { value: 'priced', label: 'Com preço' },
    { value: 'unpriced', label: 'Sem preço' },
  ]
  $: baseOptions = [
    { value: 'all', label: 'Todas' },
    { value: 'with', label: 'Com base' },
    { value: 'without', label: 'Sem base' },
    { value: 'unset', label: 'Não informado' },
  ]
  $: brandOptions = [
    { value: 'all', label: 'Todas' },
    ...brands.map((b) => ({ value: b, label: b })),
  ]
  const filterSelectClass =
    'ahq-select h-9 text-sm min-w-[8rem] w-full text-left flex items-center justify-between gap-2'
  $: q = query.trim()
  $: filtered = filterAndSortStockItems(items, {
    query,
    filterType,
    filterPrice,
    filterBase,
    filterBrand,
    sortKey,
    sortDir,
  })
  $: filteredUnits = filtered.reduce((n, it) => n + (Number(it.qty) || 1), 0)
  $: hasActiveFilters =
    filterType !== 'all' ||
    filterPrice !== 'all' ||
    filterBase !== 'all' ||
    filterBrand !== 'all' ||
    !!q

  function clearFilters() {
    query = ''
    filterType = 'all'
    filterPrice = 'all'
    filterBase = 'all'
    filterBrand = 'all'
  }

  function toggleSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc'
      return
    }
    sortKey = key
    if (key === 'title' || key === 'type') {
      sortDir = 'asc'
    } else {
      sortDir = 'desc'
    }
  }

  function sortIcon(key) {
    if (sortKey !== key) return 'unfold_more'
    return sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward'
  }

  function startEdit(item) {
    editingId = item.id
    editTitle = item.title || ''
    editSale = item.salePriceRaw || ''
  }

  function cancelEdit() {
    editingId = null
  }

  function saveItem(item) {
    const productId = Number(item.productId) || 0
    if (productId > 0) {
      router.post(
        `/products/${productId}`,
        {
          name: editTitle || item.title,
          sale_price_hint: editSale,
          return_to: '/stock',
        },
        {
          onSuccess: () => {
            editingId = null
          },
        },
      )
      return
    }
    // Fallback: unit-level edit via lot
    router.post(
      `/lots/${item.lotId}/items/${item.sampleItemId || item.id}`,
      {
        title: editTitle || item.title,
        sku: '',
        sale_price_hint: editSale,
        return_to: '/stock',
      },
      {
        onSuccess: () => {
          editingId = null
        },
      },
    )
  }

  function sellHref(item) {
    const unitId = item.sampleItemId || item.id
    return `/sales/new?item_id=${unitId}`
  }

  /** Export currently filtered product rows as CSV (client-side). */
  function exportFilteredCSV() {
    const rows = filtered || []
    const sep = ';'
    const header = [
      'produto_id',
      'titulo',
      'qtd',
      'tipo',
      'custo_un',
      'preco_venda',
      'margem_un',
      'margem_total',
    ]
    const esc = (v) => {
      const s = v == null ? '' : String(v)
      if (/[;"\n\r]/.test(s)) return `"${s.replace(/"/g, '""')}"`
      return s
    }
    const lines = [header.join(sep)]
    for (const it of rows) {
      lines.push(
        [
          it.productId || '',
          it.title,
          it.qty ?? 1,
          it.isAccessory ? 'acessorio' : 'principal',
          it.unitCost || '',
          it.salePriceHint || '',
          it.marginHint || '',
          it.marginTotal || '',
        ]
          .map(esc)
          .join(sep),
      )
    }
    const blob = new Blob(['\uFEFF' + lines.join('\n')], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const day = new Date().toISOString().slice(0, 10)
    a.href = url
    a.download = `estoque-produtos-${day}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }
</script>

<AppShell {companyName} active="stock">
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Estoque</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Produtos iguais agrupados com quantidade.
      </p>
    </div>
    <div class="flex flex-wrap gap-2 shrink-0">
      <a href="/products" use:inertia class="ahq-btn-ghost h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">category</span>
        Catálogo
      </a>
      <a href="/stock/export.csv" class="ahq-btn-ghost h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">download</span>
        CSV (unidades)
      </a>
      <button
        type="button"
        class="ahq-btn-ghost h-10 px-4 text-sm inline-flex items-center"
        on:click={exportFilteredCSV}
        disabled={filtered.length === 0}
      >
        <span class="material-symbols-outlined text-[18px] mr-1">table_view</span>
        CSV (produtos)
      </button>
      <a href="/sales/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">sell</span>
        Vender
      </a>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="grid grid-cols-2 md:grid-cols-5 gap-stack-gap mb-section-padding">
    <div class="ahq-card p-4">
      <span class="ahq-label">Unidades</span>
      <p class="ahq-value text-primary">{summary.count ?? 0}</p>
      <p class="text-[10px] text-on-surface-variant mt-1">
        {summary.productCount ?? items.length} produtos
      </p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Custo estoque</span>
      <p class="ahq-value font-mono text-sm">{summary.totalCost || 'R$ 0,00'}</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Bruto potencial</span>
      <p class="ahq-value font-mono text-sm text-secondary">{summary.potentialGross || 'R$ 0,00'}</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Multiplicador</span>
      <p class="ahq-value font-mono text-sm text-secondary">{summary.multiplier || '—'}</p>
      <p class="text-[10px] text-on-surface-variant mt-1">bruto ÷ custo</p>
    </div>
    <div class="ahq-card p-4">
      <span class="ahq-label">Margem potencial</span>
      <p class="ahq-value font-mono text-sm text-secondary">{summary.potentialMargin || 'R$ 0,00'}</p>
    </div>
  </section>

  {#if items.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">inventory_2</span>
      <p class="text-on-surface-variant mb-4">Nenhum item em estoque.</p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar lote</a>
    </div>
  {:else}
    <div class="ahq-card overflow-hidden">
      <div class="p-3 border-b border-outline-variant space-y-3">
        <div class="flex flex-col sm:flex-row gap-2 sm:items-center sm:justify-between">
          <div class="relative flex-1 max-w-md">
            <span
              class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-[20px]"
            >
              search
            </span>
            <input
              type="search"
              class="ahq-input h-10 pl-10 w-full"
              placeholder="Buscar por modelo, marca…"
              bind:value={query}
              aria-label="Buscar no estoque"
            />
          </div>
          <p class="text-sm text-on-surface-variant shrink-0">
            {filtered.length}
            {filtered.length === 1 ? 'produto' : 'produtos'}
            · {filteredUnits}
            {filteredUnits === 1 ? 'un.' : 'un.'}
            {#if hasActiveFilters}
              <span class="text-on-surface-variant/80"> de {items.length}</span>
            {/if}
          </p>
        </div>

        <div class="flex flex-wrap gap-2 items-end">
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="f-type">Tipo</label>
            <SearchableSelect
              id="f-type"
              bind:value={filterType}
              options={typeOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="f-price">Preço venda</label>
            <SearchableSelect
              id="f-price"
              bind:value={filterPrice}
              options={priceOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="f-base">Base</label>
            <SearchableSelect
              id="f-base"
              bind:value={filterBase}
              options={baseOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="f-brand">Marca</label>
            <SearchableSelect
              id="f-brand"
              bind:value={filterBrand}
              options={brandOptions}
              searchPlaceholder="Buscar marca…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          {#if hasActiveFilters}
            <button type="button" class="ahq-btn-ghost h-9 px-3 text-sm" on:click={clearFilters}>
              Limpar filtros
            </button>
          {/if}
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[720px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              <th class="px-1 py-1">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('title')}
                >
                  Produto <span class="material-symbols-outlined text-[14px]">{sortIcon('title')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-20">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('qty')}
                >
                  Qtd <span class="material-symbols-outlined text-[14px]">{sortIcon('qty')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-28">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('type')}
                >
                  Tipo <span class="material-symbols-outlined text-[14px]">{sortIcon('type')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-28">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('unitCost')}
                >
                  Custo un. <span class="material-symbols-outlined text-[14px]">{sortIcon('unitCost')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-28">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('salePrice')}
                >
                  Preço <span class="material-symbols-outlined text-[14px]">{sortIcon('salePrice')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-28">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('margin')}
                >
                  Margem un. <span class="material-symbols-outlined text-[14px]">{sortIcon('margin')}</span>
                </button>
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-36">
                Ações
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each filtered as item (item.id)}
              {#if editingId === item.id}
                <tr class="bg-secondary-container/20">
                  <td class="px-3 py-2 align-top">
                    <input class="ahq-input h-9 text-sm w-full" bind:value={editTitle} />
                    {#if (item.qty || 1) > 1}
                      <p class="text-[11px] text-on-surface-variant mt-1">
                        Renomeia as {item.qty} unidades deste produto.
                      </p>
                    {/if}
                  </td>
                  <td class="px-3 py-2 font-mono text-right align-top font-semibold">{item.qty ?? 1}</td>
                  <td class="px-3 py-2 align-top">
                    {#if item.isAccessory}
                      <span class="ahq-badge-sold text-[10px]">Acessório</span>
                    {:else}
                      <span class="ahq-badge-live text-[10px]">Principal</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2 font-mono text-right align-top">{item.unitCost}</td>
                  <td class="px-3 py-2 align-top">
                    <input
                      class="ahq-input h-9 text-sm font-mono w-full text-right"
                      placeholder="0,00"
                      bind:value={editSale}
                    />
                  </td>
                  <td class="px-3 py-2 font-mono text-right text-secondary align-top">
                    {item.marginHint || '—'}
                  </td>
                  <td class="px-3 py-2 text-right align-top whitespace-nowrap">
                    <button
                      type="button"
                      class="text-secondary font-medium text-sm mr-2"
                      on:click={() => saveItem(item)}
                    >
                      Salvar
                    </button>
                    <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>
                      Cancelar
                    </button>
                  </td>
                </tr>
              {:else}
                <tr class="hover:bg-surface-container-low/80 transition-colors">
                  <td class="px-3 py-2.5">
                    {#if item.productId}
                      <a
                        href={`/products/${item.productId}`}
                        use:inertia
                        class="font-medium text-primary leading-snug hover:underline hover:text-secondary"
                        title="Ver produto"
                      >
                        {item.title}
                      </a>
                    {:else}
                      <p class="font-medium text-primary leading-snug">{item.title}</p>
                    {/if}
                    {#if item.lotId}
                      <p class="text-[11px] font-mono text-on-surface-variant mt-0.5">
                        lote
                        <a
                          href={`/lots/${item.lotId}`}
                          use:inertia
                          class="hover:text-secondary hover:underline"
                        >
                          {item.lotId}
                        </a>
                        {#if item.productId}
                          ·
                          <a
                            href={`/products/${item.productId}`}
                            use:inertia
                            class="hover:text-secondary hover:underline"
                          >
                            prod #{item.productId}
                          </a>
                        {/if}
                      </p>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 text-right">
                    <span
                      class="inline-flex min-w-[2rem] justify-center px-2 py-0.5 rounded-full
                        bg-secondary-container text-on-secondary-container font-mono font-semibold text-sm"
                    >
                      {item.qty ?? 1}
                    </span>
                  </td>
                  <td class="px-3 py-2.5">
                    {#if item.isAccessory}
                      <span class="ahq-badge-sold text-[10px]">Acessório</span>
                    {:else}
                      <span class="ahq-badge-live text-[10px]">Principal</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 font-mono text-right whitespace-nowrap">{item.unitCost}</td>
                  <td class="px-3 py-2.5 font-mono text-right font-semibold whitespace-nowrap">
                    {item.salePriceHint || '—'}
                  </td>
                  <td class="px-3 py-2.5 font-mono text-right text-secondary whitespace-nowrap">
                    {item.marginHint || '—'}
                    {#if item.marginTotal && (item.qty || 1) > 1}
                      <span class="block text-[10px] text-on-surface-variant mt-0.5">
                        tot. {item.marginTotal}
                      </span>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-on-surface-variant hover:text-secondary text-sm font-medium mr-2"
                      on:click={() => startEdit(item)}
                    >
                      Editar
                    </button>
                    <a href={sellHref(item)} use:inertia class="text-secondary font-medium text-sm">
                      Vender
                    </a>
                  </td>
                </tr>
              {/if}
            {:else}
              <tr>
                <td colspan="7" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum produto com esses filtros.
                  <button
                    type="button"
                    class="block mx-auto mt-2 text-secondary text-sm"
                    on:click={clearFilters}
                  >
                    Limpar filtros
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>

    {#if groups.length > 0 && !hasActiveFilters}
      <p class="mt-4 text-xs text-on-surface-variant">
        Nomes reutilizáveis no
        <a href="/products" use:inertia class="text-secondary hover:underline">catálogo de produtos</a>.
      </p>
    {/if}
  {/if}
</AppShell>
