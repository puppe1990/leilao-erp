<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
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
  let sortKey = 'id'
  /** @type {'asc'|'desc'} */
  let sortDir = 'asc'

  let editingId = null
  let editTitle = ''
  let editSale = ''
  let editSku = ''

  $: brands = brandsFromItems(items)
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
    if (key === 'id' || key === 'lotId' || key === 'title' || key === 'type') {
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
    editSku = item.sku || ''
  }

  function cancelEdit() {
    editingId = null
  }

  function saveItem(item) {
    router.post(
      `/lots/${item.lotId}/items/${item.id}`,
      {
        title: editTitle || item.title,
        sku: editSku,
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
</script>

<AppShell {companyName} active="stock">
  <div class="flex items-start justify-between gap-3 mb-section-padding">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Estoque</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Lista · busca, filtros e ordenação.
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
      <!-- Search + filters -->
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
              placeholder="Buscar por modelo, id, lote, SKU…"
              bind:value={query}
              aria-label="Buscar no estoque"
            />
          </div>
          <p class="text-sm text-on-surface-variant shrink-0">
            {filtered.length}
            {filtered.length === 1 ? 'item' : 'itens'}
            {#if hasActiveFilters}
              <span class="text-on-surface-variant/80"> de {items.length}</span>
            {/if}
          </p>
        </div>

        <div class="flex flex-wrap gap-2 items-end">
          <div>
            <label class="ahq-label text-[10px] block mb-1" for="f-type">Tipo</label>
            <select id="f-type" class="ahq-select h-9 text-sm min-w-[8rem]" bind:value={filterType}>
              <option value="all">Todos</option>
              <option value="main">Principal</option>
              <option value="accessory">Acessório</option>
            </select>
          </div>
          <div>
            <label class="ahq-label text-[10px] block mb-1" for="f-price">Preço venda</label>
            <select id="f-price" class="ahq-select h-9 text-sm min-w-[8rem]" bind:value={filterPrice}>
              <option value="all">Todos</option>
              <option value="priced">Com preço</option>
              <option value="unpriced">Sem preço</option>
            </select>
          </div>
          <div>
            <label class="ahq-label text-[10px] block mb-1" for="f-base">Base</label>
            <select id="f-base" class="ahq-select h-9 text-sm min-w-[8rem]" bind:value={filterBase}>
              <option value="all">Todas</option>
              <option value="with">Com base</option>
              <option value="without">Sem base</option>
              <option value="unset">Não informado</option>
            </select>
          </div>
          <div>
            <label class="ahq-label text-[10px] block mb-1" for="f-brand">Marca</label>
            <select id="f-brand" class="ahq-select h-9 text-sm min-w-[8rem]" bind:value={filterBrand}>
              <option value="all">Todas</option>
              {#each brands as b}
                <option value={b}>{b}</option>
              {/each}
            </select>
          </div>
          {#if hasActiveFilters}
            <button
              type="button"
              class="ahq-btn-ghost h-9 px-3 text-sm"
              on:click={clearFilters}
            >
              Limpar filtros
            </button>
          {/if}
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[720px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              <th class="px-1 py-1 w-14">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('id')}
                >
                  # <span class="material-symbols-outlined text-[14px]">{sortIcon('id')}</span>
                </button>
              </th>
              <th class="px-1 py-1">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('title')}
                >
                  Título <span class="material-symbols-outlined text-[14px]">{sortIcon('title')}</span>
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
                  Custo <span class="material-symbols-outlined text-[14px]">{sortIcon('unitCost')}</span>
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
                  Margem <span class="material-symbols-outlined text-[14px]">{sortIcon('margin')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-16">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('lotId')}
                >
                  Lote <span class="material-symbols-outlined text-[14px]">{sortIcon('lotId')}</span>
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
                  <td class="px-3 py-2 font-mono text-on-surface-variant align-top">{item.id}</td>
                  <td class="px-3 py-2 align-top" colspan="2">
                    <input class="ahq-input h-9 text-sm w-full mb-1.5" bind:value={editTitle} />
                    <input
                      class="ahq-input h-9 text-sm font-mono w-full"
                      placeholder="SKU"
                      bind:value={editSku}
                    />
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
                  <td class="px-3 py-2 font-mono text-on-surface-variant align-top">{item.lotId}</td>
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
                  <td class="px-3 py-2.5 font-mono text-on-surface-variant">{item.id}</td>
                  <td class="px-3 py-2.5">
                    <p class="font-medium text-primary leading-snug">{item.title}</p>
                    {#if item.sku}
                      <p class="text-[11px] font-mono text-on-surface-variant mt-0.5">{item.sku}</p>
                    {/if}
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
                  </td>
                  <td class="px-3 py-2.5 font-mono text-on-surface-variant">
                    <a
                      href={`/lots/${item.lotId}`}
                      use:inertia
                      class="hover:text-secondary hover:underline"
                    >
                      {item.lotId}
                    </a>
                  </td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-on-surface-variant hover:text-secondary text-sm font-medium mr-2"
                      on:click={() => startEdit(item)}
                    >
                      Editar
                    </button>
                    <a
                      href={`/sales/new?item_id=${item.id}`}
                      use:inertia
                      class="text-secondary font-medium text-sm"
                    >
                      Vender
                    </a>
                  </td>
                </tr>
              {/if}
            {:else}
              <tr>
                <td colspan="8" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum item com esses filtros.
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
      <details class="mt-section-padding ahq-card p-4">
        <summary class="cursor-pointer font-semibold text-primary select-none">
          Resumo por modelo ({groups.length})
        </summary>
        <div class="mt-3 overflow-x-auto">
          <table class="w-full text-sm min-w-[480px]">
            <thead>
              <tr class="text-on-surface-variant border-b border-outline-variant">
                <th class="px-2 py-2 text-left text-[11px] uppercase">Modelo</th>
                <th class="px-2 py-2 text-right text-[11px] uppercase">Qtd</th>
                <th class="px-2 py-2 text-right text-[11px] uppercase">Custo un.</th>
                <th class="px-2 py-2 text-right text-[11px] uppercase">Preço</th>
                <th class="px-2 py-2 text-right text-[11px] uppercase">Margem pot.</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              {#each groups as g}
                <tr>
                  <td class="px-2 py-2 font-medium">{g.title}</td>
                  <td class="px-2 py-2 text-right font-mono">{g.count}</td>
                  <td class="px-2 py-2 text-right font-mono">{g.unitCost}</td>
                  <td class="px-2 py-2 text-right font-mono">{g.salePriceHint}</td>
                  <td class="px-2 py-2 text-right font-mono text-secondary">{g.potentialMargin}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </details>
    {/if}
  {/if}
</AppShell>
