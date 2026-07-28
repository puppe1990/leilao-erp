<script>
  import { useForm, inertia, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'

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

  $: cashAccountOptions = (Array.isArray(cashAccounts) ? cashAccounts : []).map((acc) => ({
    value: String(acc.id),
    label: acc.name,
  }))

  function submitCost() {
    costForm.post(`/lots/${lot.id}/costs`)
  }
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
          <button type="button" class="ahq-btn-ghost h-10 px-4 text-sm text-error border-error"
            on:click={() => { if (confirm('Excluir este lote e itens em estoque?')) router.post(`/lots/${lot.id}/delete`) }}>
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
        <div class="divide-y divide-outline-variant">
          {#each items as item}
            <div class="p-4 space-y-3">
              <div class="flex justify-between items-center gap-3">
                <div>
                  <p class="font-semibold text-primary">{item.title}</p>
                  <p class="text-[10px] font-mono text-on-surface-variant uppercase">#{item.id}{item.sku ? ` · ${item.sku}` : ''}</p>
                </div>
                <div class="text-right">
                  <p class="ahq-value text-sm">{item.unitCost}</p>
                  <p class="text-[10px] text-on-surface-variant uppercase mt-0.5">{item.statusLabel}</p>
                </div>
              </div>
              {#if item.salePriceHint}
                <div class="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span class="ahq-label text-[10px]">Preço venda</span>
                    <p class="font-mono font-semibold text-secondary">{item.salePriceHint}</p>
                  </div>
                  <div>
                    <span class="ahq-label text-[10px]">Margem pot.</span>
                    <p class="font-mono font-semibold">{item.marginHint || '—'}</p>
                  </div>
                </div>
              {/if}
              {#if item.canEdit}
                <form
                  class="grid grid-cols-2 gap-2"
                  on:submit|preventDefault={() => {
                    router.post(`/lots/${lot.id}/items/${item.id}`, {
                      title: item._title ?? item.title,
                      sku: item._sku ?? item.sku ?? '',
                      sale_price_hint: item._sale ?? item.salePriceRaw ?? '',
                    })
                  }}
                >
                  <input class="ahq-input h-10 text-sm col-span-2" placeholder="Título" value={item.title}
                    on:input={(e) => (item._title = e.currentTarget.value)} />
                  <input class="ahq-input h-10 text-sm font-mono" placeholder="Preço venda" value={item.salePriceRaw || ''}
                    on:input={(e) => (item._sale = e.currentTarget.value)} />
                  <input class="ahq-input h-10 text-sm font-mono" placeholder="SKU" value={item.sku || ''}
                    on:input={(e) => (item._sku = e.currentTarget.value)} />
                  <button type="submit" class="col-span-2 ahq-btn-ghost h-10 text-sm">Salvar item</button>
                </form>
              {/if}
            </div>
          {/each}
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
