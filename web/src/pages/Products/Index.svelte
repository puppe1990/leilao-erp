<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import {
    brandsFromProducts,
    filterAndSortProducts,
  } from '@/lib/productsTable.js'

  export let products = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let query = ''
  /** @type {'all'|'main'|'accessory'} */
  let filterType = 'all'
  /** @type {'all'|'priced'|'unpriced'} */
  let filterPrice = 'all'
  /** @type {'all'|'in_stock'|'out'} */
  let filterStock = 'all'
  /** @type {'all'|'photo'|'video'|'any'|'none'} */
  let filterMedia = 'all'
  /** @type {'all'|'with'|'without'|'unset'} */
  let filterBase = 'all'
  let filterBrand = 'all'
  /** @type {string} */
  let sortKey = 'name'
  /** @type {'asc'|'desc'} */
  let sortDir = 'asc'

  const filterSelectClass =
    'ahq-select h-9 text-sm min-w-[8rem] w-full text-left flex items-center justify-between gap-2'

  $: brands = brandsFromProducts(products)
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
  $: stockOptions = [
    { value: 'all', label: 'Todos' },
    { value: 'in_stock', label: 'Em estoque' },
    { value: 'out', label: 'Sem estoque' },
  ]
  $: mediaOptions = [
    { value: 'all', label: 'Todas' },
    { value: 'any', label: 'Com mídia' },
    { value: 'photo', label: 'Com foto' },
    { value: 'video', label: 'Com vídeo' },
    { value: 'none', label: 'Sem mídia' },
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

  $: filtered = filterAndSortProducts(products, {
    query,
    filterType,
    filterPrice,
    filterStock,
    filterMedia,
    filterBase,
    filterBrand,
    sortKey,
    sortDir,
  })
  $: hasActiveFilters =
    filterType !== 'all' ||
    filterPrice !== 'all' ||
    filterStock !== 'all' ||
    filterMedia !== 'all' ||
    filterBase !== 'all' ||
    filterBrand !== 'all' ||
    !!query.trim()

  function clearFilters() {
    query = ''
    filterType = 'all'
    filterPrice = 'all'
    filterStock = 'all'
    filterMedia = 'all'
    filterBase = 'all'
    filterBrand = 'all'
  }

  function toggleSort(key) {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc'
      return
    }
    sortKey = key
    sortDir = key === 'name' || key === 'type' ? 'asc' : 'desc'
  }

  function sortIcon(key) {
    if (sortKey !== key) return 'unfold_more'
    return sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward'
  }
</script>

<AppShell {companyName} active="products">
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Produtos</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Listagem do catálogo. Clique no nome para ver, ou em Editar.
      </p>
    </div>
    <div class="flex flex-wrap gap-2">
      <a href="/stock" use:inertia class="ahq-btn-ghost h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">inventory_2</span>
        Estoque
      </a>
      <a href="/lots/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm inline-flex items-center">
        <span class="material-symbols-outlined text-[18px] mr-1">add</span>
        Novo lote
      </a>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if products.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">category</span>
      <p class="text-on-surface-variant mb-4">
        Nenhum produto ainda. Ao cadastrar lotes, os nomes viram produtos automaticamente.
      </p>
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
              placeholder="Buscar nome, marca, ficha…"
              bind:value={query}
              aria-label="Buscar produtos"
            />
          </div>
          <p class="text-sm text-on-surface-variant shrink-0">
            {filtered.length}
            {filtered.length === 1 ? 'produto' : 'produtos'}
            {#if hasActiveFilters}
              <span class="text-on-surface-variant/80"> de {(products || []).length}</span>
            {/if}
          </p>
        </div>

        <div class="flex flex-wrap gap-2 items-end">
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-type">Tipo</label>
            <SearchableSelect
              id="pf-type"
              bind:value={filterType}
              options={typeOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-price">Preço</label>
            <SearchableSelect
              id="pf-price"
              bind:value={filterPrice}
              options={priceOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-stock">Estoque</label>
            <SearchableSelect
              id="pf-stock"
              bind:value={filterStock}
              options={stockOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-media">Mídia</label>
            <SearchableSelect
              id="pf-media"
              bind:value={filterMedia}
              options={mediaOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-base">Base</label>
            <SearchableSelect
              id="pf-base"
              bind:value={filterBase}
              options={baseOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="pf-brand">Marca</label>
            <SearchableSelect
              id="pf-brand"
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
                  on:click={() => toggleSort('name')}
                >
                  Nome <span class="material-symbols-outlined text-[14px]">{sortIcon('name')}</span>
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
              <th class="px-1 py-1 w-20">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('qty')}
                >
                  Estoque <span class="material-symbols-outlined text-[14px]">{sortIcon('qty')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-24">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-center gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('media')}
                >
                  Mídia <span class="material-symbols-outlined text-[14px]">{sortIcon('media')}</span>
                </button>
              </th>
              <th class="px-1 py-1 w-32">
                <button
                  type="button"
                  class="w-full px-2 py-1.5 flex items-center justify-end gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                  on:click={() => toggleSort('salePrice')}
                >
                  Preço <span class="material-symbols-outlined text-[14px]">{sortIcon('salePrice')}</span>
                </button>
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-36">
                Ações
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each filtered as p (p.id)}
              <tr class="hover:bg-surface-container-low/80">
                <td class="px-3 py-2.5">
                  <a
                    href={`/products/${p.id}`}
                    use:inertia
                    class="font-medium text-primary hover:underline hover:text-secondary"
                    title="Ver produto"
                  >
                    {p.name}
                  </a>
                </td>
                <td class="px-3 py-2.5">
                  {#if p.kind === 'accessory'}
                    <span class="ahq-badge-sold text-[10px]">Acessório</span>
                  {:else}
                    <span class="ahq-badge-live text-[10px]">Principal</span>
                  {/if}
                </td>
                <td class="px-3 py-2.5 font-mono text-right">
                  {#if p.qtyInStock > 0}
                    <span
                      class="inline-flex min-w-[2rem] justify-center px-2 py-0.5 rounded-full
                        bg-secondary-container text-on-secondary-container font-semibold"
                    >
                      {p.qtyInStock}
                    </span>
                  {:else}
                    <span class="text-on-surface-variant">0</span>
                  {/if}
                </td>
                <td class="px-3 py-2.5 text-center">
                  <span class="text-xs text-on-surface-variant font-mono">
                    <span class="material-symbols-outlined text-[16px] align-middle">image</span>
                    {p.photoCount || 0}
                    <span class="material-symbols-outlined text-[16px] align-middle ml-1">movie</span>
                    {p.videoCount || 0}
                  </span>
                </td>
                <td class="px-3 py-2.5 font-mono text-right font-semibold">
                  {p.salePriceHint || '—'}
                </td>
                <td class="px-3 py-2.5 text-right whitespace-nowrap">
                  <a
                    href={`/products/${p.id}`}
                    use:inertia
                    class="text-secondary text-sm font-medium mr-2"
                  >
                    Ver
                  </a>
                  <a
                    href={`/products/${p.id}/edit`}
                    use:inertia
                    class="text-on-surface-variant hover:text-secondary text-sm font-medium"
                  >
                    Editar
                  </a>
                </td>
              </tr>
            {:else}
              <tr>
                <td colspan="6" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum produto com esses filtros.
                  {#if hasActiveFilters}
                    <button
                      type="button"
                      class="block mx-auto mt-2 text-secondary text-sm"
                      on:click={clearFilters}
                    >
                      Limpar filtros
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</AppShell>
