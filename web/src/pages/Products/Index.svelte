<script>
  import { inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let products = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let query = ''
  let editingId = null
  let editName = ''
  let editSale = ''

  $: q = query.trim().toLowerCase()
  $: filtered = (products || []).filter((p) => {
    if (!q) return true
    return String(p.name || '')
      .toLowerCase()
      .includes(q)
  })

  function startEdit(p) {
    editingId = p.id
    editName = p.name || ''
    editSale = p.salePriceRaw || ''
  }

  function cancelEdit() {
    editingId = null
  }

  function save(p) {
    router.post(
      `/products/${p.id}`,
      {
        name: editName || p.name,
        sale_price_hint: editSale,
        return_to: '/products',
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
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Produtos</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Catálogo de nomes reutilizados no estoque. Editar aqui atualiza as unidades em estoque.
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
      <div class="p-3 border-b border-outline-variant flex flex-col sm:flex-row gap-2 sm:items-center sm:justify-between">
        <div class="relative flex-1 max-w-md">
          <span
            class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-[20px]"
          >
            search
          </span>
          <input
            type="search"
            class="ahq-input h-10 pl-10 w-full"
            placeholder="Buscar nome…"
            bind:value={query}
            aria-label="Buscar produtos"
          />
        </div>
        <p class="text-sm text-on-surface-variant">
          {filtered.length}
          {filtered.length === 1 ? 'produto' : 'produtos'}
        </p>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[640px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide">Nome</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide w-28">Tipo</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-20">
                Estoque
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-32">
                Preço venda
              </th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase tracking-wide text-right w-28">
                Ações
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each filtered as p (p.id)}
              {#if editingId === p.id}
                <tr class="bg-secondary-container/20">
                  <td class="px-3 py-2">
                    <input class="ahq-input h-9 text-sm w-full" bind:value={editName} />
                  </td>
                  <td class="px-3 py-2 text-on-surface-variant">{p.kindLabel}</td>
                  <td class="px-3 py-2 font-mono text-right">{p.qtyInStock}</td>
                  <td class="px-3 py-2">
                    <input
                      class="ahq-input h-9 text-sm font-mono w-full text-right"
                      placeholder="0,00"
                      bind:value={editSale}
                    />
                  </td>
                  <td class="px-3 py-2 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-secondary font-medium text-sm mr-2"
                      on:click={() => save(p)}
                    >
                      Salvar
                    </button>
                    <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>
                      Cancelar
                    </button>
                  </td>
                </tr>
              {:else}
                <tr class="hover:bg-surface-container-low/80">
                  <td class="px-3 py-2.5 font-medium text-primary">{p.name}</td>
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
                  <td class="px-3 py-2.5 font-mono text-right font-semibold">
                    {p.salePriceHint || '—'}
                  </td>
                  <td class="px-3 py-2.5 text-right">
                    <button
                      type="button"
                      class="text-on-surface-variant hover:text-secondary text-sm font-medium"
                      on:click={() => startEdit(p)}
                    >
                      Editar
                    </button>
                  </td>
                </tr>
              {/if}
            {:else}
              <tr>
                <td colspan="5" class="px-3 py-10 text-center text-on-surface-variant">
                  Nenhum produto com essa busca.
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</AppShell>
