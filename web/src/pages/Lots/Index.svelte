<script>
  import { inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import {
    filterAndSortLots,
    loadLotSorts,
    saveLotSorts,
    stackLotSort,
  } from '@/lib/lotsTable.js'

  export let lots = []
  export let site = {}
  export let companyName = 'Pix Eletro'

  let query = ''
  /** @type {'all'|'open'|'partial'|'sold'|'closed'} */
  let filterStatus = 'all'
  /** @type {{ key: string, dir: 'asc'|'desc' }[]} */
  let sorts = loadLotSorts()
  let sortsReady = true

  $: if (sortsReady) saveLotSorts(sorts)

  const filterSelectClass =
    'ahq-select h-9 text-sm min-w-[8rem] w-full text-left flex items-center justify-between gap-2'

  $: statusOptions = [
    { value: 'all', label: 'Todos' },
    { value: 'open', label: 'Aberto' },
    { value: 'partial', label: 'Parcial' },
    { value: 'sold', label: 'Vendido' },
    { value: 'closed', label: 'Fechado' },
  ]

  $: filtered = filterAndSortLots(lots, { query, filterStatus, sorts })
  $: hasActiveFilters = filterStatus !== 'all' || !!query.trim()
  $: multiSort = sorts.length > 1
  /** @type {Record<string, number>} */
  $: sortRankByKey = Object.fromEntries(sorts.map((s, i) => [s.key, i + 1]))
  /** @type {Record<string, 'asc'|'desc'>} */
  $: sortDirByKey = Object.fromEntries(sorts.map((s) => [s.key, s.dir]))

  function clearFilters() {
    query = ''
    filterStatus = 'all'
  }

  function clearSorts() {
    sorts = [{ key: 'purchasedAt', dir: 'desc' }]
    saveLotSorts(sorts)
  }

  function toggleSort(key) {
    sorts = stackLotSort(sorts, key)
  }

  /** Template reads sortDirByKey[key] so Svelte invalidates the icon. */
  function iconForDir(dir) {
    if (dir === 'asc') return 'arrow_upward'
    if (dir === 'desc') return 'arrow_downward'
    return 'unfold_more'
  }

  const lotColumns = [
    { key: 'name', label: 'Lote', thClass: 'px-1 py-1', btnClass: 'items-center' },
    { key: 'purchasedAt', label: 'Data', thClass: 'px-1 py-1 w-32', btnClass: 'items-center' },
    { key: 'status', label: 'Status', thClass: 'px-1 py-1 w-28', btnClass: 'items-center' },
    { key: 'itemCount', label: 'Itens', thClass: 'px-1 py-1 w-24', btnClass: 'items-center justify-end' },
    { key: 'totalCost', label: 'Custo', thClass: 'px-1 py-1 w-32', btnClass: 'items-center justify-end' },
  ]

  function statusBadge(status) {
    const s = (status || '').toLowerCase()
    if (s.includes('open') || s.includes('aberto')) return 'ahq-badge-open'
    if (s.includes('partial') || s.includes('parcial')) return 'ahq-badge-pending'
    if (s.includes('sold') || s.includes('vend')) return 'ahq-badge-sold'
    return 'ahq-badge-sold'
  }
</script>

<AppShell {companyName} active="lots">
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Lotes</h1>
      <p class="text-on-surface-variant text-body-md mt-1">Compras e rateio de custos.</p>
    </div>
    <a href="/lots/new" use:inertia class="ahq-btn-primary h-10 px-4 text-sm shrink-0 inline-flex items-center">
      <span class="material-symbols-outlined text-[18px] mr-1">add</span>
      Novo
    </a>
  </div>

  {#if lots.length === 0}
    <div class="ahq-card p-10 text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-on-surface-variant mb-3">inventory_2</span>
      <p class="text-on-surface-variant mb-4">Nenhum lote cadastrado.</p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar primeiro lote</a>
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
              placeholder="Buscar lote, data, status…"
              bind:value={query}
              aria-label="Buscar lotes"
            />
          </div>
          <p class="text-sm text-on-surface-variant shrink-0">
            {filtered.length}
            {filtered.length === 1 ? 'lote' : 'lotes'}
            {#if hasActiveFilters}
              <span class="text-on-surface-variant/80"> de {(lots || []).length}</span>
            {/if}
          </p>
        </div>

        <div class="flex flex-wrap gap-2 items-end">
          <div class="min-w-[8rem]">
            <label class="ahq-label text-[10px] block mb-1" for="lf-status">Status</label>
            <SearchableSelect
              id="lf-status"
              bind:value={filterStatus}
              options={statusOptions}
              searchPlaceholder="Buscar…"
              allowClear={false}
              buttonClass={filterSelectClass}
            />
          </div>
          {#if hasActiveFilters}
            <button type="button" class="ahq-btn-ghost h-9 px-3 text-sm" on:click={clearFilters}>
              Limpar filtros
            </button>
          {/if}
          {#if multiSort}
            <button
              type="button"
              class="ahq-btn-ghost h-9 px-3 text-sm"
              title="Voltar à ordem só por data"
              on:click={clearSorts}
            >
              Limpar ordem ({sorts.length})
            </button>
          {/if}
        </div>
        {#if multiSort}
          <p class="text-[11px] text-on-surface-variant">
            Ordenação empilhada: última coluna clicada é a principal (#1); as outras viram desempate
            (#2, #3…). Clicar de novo na mesma coluna inverte a direção.
          </p>
        {/if}
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[640px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              {#each lotColumns as col (col.key)}
                <th class={col.thClass}>
                  <button
                    type="button"
                    class="w-full px-2 py-1.5 flex {col.btnClass} gap-0.5 font-medium text-[11px] hover:text-primary"
                    on:click={() => toggleSort(col.key)}
                  >
                    {col.label}
                    {#if multiSort && sortRankByKey[col.key]}
                      <span class="text-[10px] font-bold text-secondary">{sortRankByKey[col.key]}</span>
                    {/if}
                    <span class="material-symbols-outlined text-[14px]"
                      >{iconForDir(sortDirByKey[col.key])}</span
                    >
                  </button>
                </th>
              {/each}
              <th class="px-3 py-2.5 font-medium text-[11px] text-right w-24">
                Ações
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each filtered as lot (lot.id)}
              <tr class="hover:bg-surface-container-low/80">
                <td class="px-3 py-2.5">
                  <a
                    href={`/lots/${lot.id}`}
                    use:inertia
                    class="font-medium text-primary hover:underline hover:text-secondary"
                    title="Ver lote"
                  >
                    {lot.name}
                  </a>
                </td>
                <td class="px-3 py-2.5 font-mono text-on-surface-variant whitespace-nowrap">
                  {lot.purchasedAt}
                </td>
                <td class="px-3 py-2.5">
                  <span class={statusBadge(lot.status || lot.statusLabel)}>{lot.statusLabel || lot.status}</span>
                </td>
                <td class="px-3 py-2.5 font-mono text-right">
                  {lot.itemCount}
                </td>
                <td class="px-3 py-2.5 font-mono text-right font-semibold">
                  {lot.totalCost}
                </td>
                <td class="px-3 py-2.5 text-right whitespace-nowrap">
                  <a href={`/lots/${lot.id}`} use:inertia class="text-secondary text-sm font-medium">
                    Ver
                  </a>
                </td>
              </tr>
            {:else}
              <tr>
                <td colspan="6" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum lote com esses filtros.
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
