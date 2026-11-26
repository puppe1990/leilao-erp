<script>
  import { useForm, inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'
  import {
    filterAndSortLotItems,
    loadLotItemSorts,
    saveLotItemSorts,
    stackLotItemSort,
  } from '@/lib/lotItemsTable.js'

  export let lot = {}
  export let items = []
  export let costs = []
  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let canDelete = false
  export let companyName = 'AuctionHQ'

  let costForm = useForm({
    cost_label: 'Frete',
    cost_amount: '',
    already_paid: false,
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
  })

  let query = ''
  /** @type {'all'|'in_stock'|'reserved'|'sold'} */
  let filterStatus = 'all'
  /** @type {{ key: string, dir: 'asc'|'desc' }[]} */
  let sorts = loadLotItemSorts()
  let sortsReady = true

  $: if (sortsReady) saveLotItemSorts(sorts)

  const filterSelectClass =
    'ahq-select h-9 text-sm min-w-[8rem] w-full text-left flex items-center justify-between gap-2'

  $: cashAccountOptions = (Array.isArray(cashAccounts) ? cashAccounts : []).map((acc) => ({
    value: String(acc.id),
    label: acc.name,
  }))

  $: statusOptions = [
    { value: 'all', label: 'Todos' },
    { value: 'in_stock', label: 'Em estoque' },
    { value: 'reserved', label: 'Reservado' },
    { value: 'sold', label: 'Vendido' },
  ]

  $: filtered = filterAndSortLotItems(items, { query, filterStatus, sorts })
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
    sorts = [{ key: 'id', dir: 'asc' }]
    saveLotItemSorts(sorts)
  }

  function toggleSort(key) {
    sorts = stackLotItemSort(sorts, key)
  }

  function iconForDir(dir) {
    if (dir === 'asc') return 'arrow_upward'
    if (dir === 'desc') return 'arrow_downward'
    return 'unfold_more'
  }

  function itemBadge(status) {
    const s = (status || '').toLowerCase()
    if (s.includes('in_stock') || s.includes('estoque')) return 'ahq-badge-open'
    if (s.includes('reserved') || s.includes('reserv')) return 'ahq-badge-pending'
    return 'ahq-badge-sold'
  }

  function saveLotItem(item) {
    router.post(`/lots/${lot.id}/items/${item.id}`, {
      title: item._title ?? item.title,
      sku: item._sku ?? item.sku ?? '',
      sale_price_hint: item._sale ?? item.salePriceRaw ?? '',
      unit_cost: item._cost ?? item.unitCostRaw ?? '',
    })
  }

  function submitCost() {
    costForm.post(`/lots/${lot.id}/costs`)
  }

  async function destroyLot() {
    const ok = await askConfirm({
      title: 'Excluir lote',
      message: `Excluir o lote “${lot.name}” e os itens ainda em estoque?`,
      detail: 'Itens vendidos não são removidos. Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir lote',
      tone: 'danger',
      icon: 'delete_forever',
    })
    if (!ok) return
    router.post(`/lots/${lot.id}/delete`)
  }

  const itemColumns = [
    { key: 'id', label: '#', thClass: 'px-1 py-1 w-16', btnClass: 'items-center' },
    { key: 'title', label: 'Item', thClass: 'px-1 py-1', btnClass: 'items-center' },
    { key: 'sku', label: 'SKU', thClass: 'px-1 py-1 w-28', btnClass: 'items-center' },
    { key: 'unitCost', label: 'Custo', thClass: 'px-1 py-1 w-24', btnClass: 'items-center justify-end' },
    { key: 'salePrice', label: 'Venda', thClass: 'px-1 py-1 w-24', btnClass: 'items-center justify-end' },
    { key: 'margin', label: 'Margem', thClass: 'px-1 py-1 w-24', btnClass: 'items-center justify-end' },
    { key: 'status', label: 'Status', thClass: 'px-1 py-1 w-36', btnClass: 'items-center' },
  ]
</script>

<AppShell {companyName} active="lots">
  <div class="mb-section-padding">
    <a href="/lots" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Lotes
    </a>
    <div class="flex items-start justify-between gap-3">
      <div>
        <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{lot.name}</h1>
        <p class="text-on-surface-variant text-body-md mt-1">
          Compra em {lot.purchasedAt} · {lot.statusLabel} · Custo total
          <span class="font-mono font-semibold text-primary">{lot.totalCost}</span>
        </p>
      </div>
      <div class="flex flex-col gap-2 shrink-0">
        <a href={`/lots/${lot.id}/edit`} use:inertia class="ahq-btn-primary h-10 px-4 text-sm">Editar</a>
        {#if canDelete}
          <button
            type="button"
            class="ahq-btn-ghost h-10 px-4 text-sm text-error border-error"
            on:click={destroyLot}
          >
            Excluir
          </button>
        {/if}
      </div>
    </div>
  </div>

  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Itens</h2>
    {#if items.length === 0}
      <p class="text-on-surface-variant text-sm">Nenhum item.</p>
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
                placeholder="Buscar item, SKU, status…"
                bind:value={query}
                aria-label="Buscar itens do lote"
              />
            </div>
            <p class="text-sm text-on-surface-variant shrink-0">
              {filtered.length}
              {filtered.length === 1 ? 'item' : 'itens'}
              {#if hasActiveFilters}
                <span class="text-on-surface-variant/80"> de {(items || []).length}</span>
              {/if}
            </p>
          </div>

          <div class="flex flex-wrap gap-2 items-end">
            <div class="min-w-[8rem]">
              <label class="ahq-label text-[10px] block mb-1" for="li-status">Status</label>
              <SearchableSelect
                id="li-status"
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
                title="Voltar à ordem por #"
                on:click={clearSorts}
              >
                Limpar ordem ({sorts.length})
              </button>
            {/if}
          </div>
          {#if multiSort}
            <p class="text-[11px] text-on-surface-variant">
              Ordenação empilhada: última coluna clicada é a principal (#1). Clicar de novo inverte a
              direção.
            </p>
          {/if}
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-sm text-left min-w-[860px]">
            <thead>
              <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
                {#each itemColumns as col (col.key)}
                  <th class={col.thClass}>
                    <button
                      type="button"
                      class="w-full px-2 py-1.5 flex {col.btnClass} gap-0.5 font-medium text-[11px] uppercase tracking-wide hover:text-primary"
                      on:click={() => toggleSort(col.key)}
                    >
                      {col.label}
                      {#if multiSort && sortRankByKey[col.key]}
                        <span class="text-[10px] font-bold text-secondary">{sortRankByKey[col.key]}</span>
                      {/if}
                      <span class="material-symbols-outlined text-[14px]">{iconForDir(sortDirByKey[col.key])}</span>
                    </button>
                  </th>
                {/each}
                <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-20">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-outline-variant">
              {#each filtered as item (item.id)}
                <tr class="hover:bg-surface-container-low/80">
                  <td class="px-3 py-1.5 font-mono text-on-surface-variant whitespace-nowrap">
                    #{item.id}
                  </td>
                  <td class="px-2 py-1.5">
                    {#if item.canEdit}
                      <input
                        class="ahq-input h-8 text-sm w-full min-w-[12rem]"
                        placeholder="Título"
                        value={item._title ?? item.title}
                        on:input={(e) => (item._title = e.currentTarget.value)}
                      />
                    {:else}
                      <span class="font-medium text-primary">{item.title}</span>
                    {/if}
                  </td>
                  <td class="px-2 py-1.5">
                    {#if item.canEdit}
                      <input
                        class="ahq-input h-8 text-sm font-mono w-full"
                        placeholder=""
                        value={item._sku ?? item.sku ?? ''}
                        on:input={(e) => (item._sku = e.currentTarget.value)}
                      />
                    {:else}
                      <span class="font-mono text-on-surface-variant">{item.sku || '—'}</span>
                    {/if}
                  </td>
                  <td class="px-2 py-1.5">
                    {#if item.canEdit}
                      <input
                        class="ahq-input h-8 text-sm font-mono w-full text-right"
                        placeholder="0,00"
                        value={item._cost ?? item.unitCostRaw ?? ''}
                        on:input={(e) => (item._cost = e.currentTarget.value)}
                      />
                    {:else}
                      <span class="font-mono block text-right">{item.unitCost}</span>
                    {/if}
                  </td>
                  <td class="px-2 py-1.5">
                    {#if item.canEdit}
                      <input
                        class="ahq-input h-8 text-sm font-mono w-full text-right"
                        placeholder="0,00"
                        value={item._sale ?? item.salePriceRaw ?? ''}
                        on:input={(e) => (item._sale = e.currentTarget.value)}
                      />
                    {:else}
                      <span class="font-mono block text-right text-secondary">{item.salePriceHint || '—'}</span>
                    {/if}
                  </td>
                  <td class="px-3 py-1.5 font-mono text-right whitespace-nowrap">
                    {item.marginHint || '—'}
                  </td>
                  <td class="px-3 py-1.5 whitespace-nowrap">
                    <span class={itemBadge(item.status || item.statusLabel)}>{item.statusLabel || item.status}</span>
                  </td>
                  <td class="px-3 py-1.5 text-right whitespace-nowrap">
                    {#if item.canEdit}
                      <button
                        type="button"
                        class="text-secondary text-sm font-medium"
                        on:click={() => saveLotItem(item)}
                      >
                        Salvar
                      </button>
                    {:else}
                      <span class="text-on-surface-variant text-sm">—</span>
                    {/if}
                  </td>
                </tr>
              {:else}
                <tr>
                  <td colspan="8" class="px-3 py-10 text-center text-on-surface-variant">
                    Nenhum item com esses filtros.
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
  </section>

  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Custos</h2>
    {#if costs.length === 0}
      <p class="text-on-surface-variant text-sm">Nenhum custo.</p>
    {:else}
      <ul class="ahq-card divide-y divide-outline-variant">
        {#each costs as cost}
          <li class="p-4 flex justify-between text-body-md">
            <span>{cost.label}</span>
            <span class="font-mono font-semibold">{cost.amount}</span>
          </li>
        {/each}
      </ul>
    {/if}

    {#if payables.length > 0}
      <h3 class="ahq-label mt-4 mb-2">A pagar / pagos</h3>
      <ul class="space-y-2">
        {#each payables as p}
          <li class="ahq-card p-3 flex justify-between text-sm">
            <span class="text-on-surface-variant">{p.description} · {p.statusLabel}</span>
            <span class="font-mono font-medium">{p.amount}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="ahq-card p-5">
    <h2 class="font-headline-md text-headline-md text-primary mb-3">Adicionar custo</h2>
    {#if errors.form}
      <p class="mb-2 text-error text-sm">{errors.form}</p>
    {/if}
    <form on:submit|preventDefault={submitCost} class="space-y-3">
      <div>
        <label class="ahq-label block mb-1.5" for="cost_label">Rótulo</label>
        <input id="cost_label" type="text" bind:value={costForm.cost_label} class="ahq-input" />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="cost_amount">Valor (R$)</label>
        <input
          id="cost_amount"
          type="text"
          bind:value={costForm.cost_amount}
          class="ahq-input font-mono"
          placeholder="80,00"
        />
        {#if errors.cost_amount}<p class="text-error text-sm mt-1">{errors.cost_amount}</p>{/if}
      </div>
      <label class="flex items-center gap-2 text-body-md">
        <input type="checkbox" bind:checked={costForm.already_paid} class="rounded border-outline-variant text-secondary" />
        Já paguei
      </label>
      {#if costForm.already_paid}
        <div>
          <label class="ahq-label block mb-1.5" for="cash_account_id">Conta</label>
          <SearchableSelect
            id="cash_account_id"
            bind:value={costForm.cash_account_id}
            options={cashAccountOptions}
            placeholder="Selecione…"
            searchPlaceholder="Buscar conta…"
          />
        </div>
      {/if}
      <button type="submit" class="ahq-btn-primary w-full" disabled={costForm.processing}>
        Adicionar custo
      </button>
    </form>
  </section>
</AppShell>
