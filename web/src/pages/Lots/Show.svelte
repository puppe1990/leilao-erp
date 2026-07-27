<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let lot = {}
  export let items = []
  export let costs = []
  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let costForm = useForm({
    cost_label: 'Frete',
    cost_amount: '',
    already_paid: false,
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
  })

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
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">{lot.name}</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Compra em {lot.purchasedAt} · {lot.statusLabel} · Custo total
      <span class="font-mono font-semibold text-primary">{lot.totalCost}</span>
    </p>
  </div>

  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Itens</h2>
    {#if items.length === 0}
      <p class="text-on-surface-variant text-sm">Nenhum item.</p>
    {:else}
      <div class="ahq-card overflow-hidden">
        <div class="divide-y divide-outline-variant">
          {#each items as item}
            <div class="p-4 flex justify-between items-center gap-3">
              <div>
                <p class="font-semibold text-primary">{item.title}</p>
                <p class="text-[10px] font-mono text-on-surface-variant uppercase">#{item.id}</p>
              </div>
              <div class="text-right">
                <p class="ahq-value text-sm">{item.unitCost}</p>
                <p class="text-[10px] text-on-surface-variant uppercase mt-0.5">{item.statusLabel}</p>
              </div>
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
          <select id="cash_account_id" bind:value={costForm.cash_account_id} class="ahq-select">
            <option value="">Selecione…</option>
            {#each cashAccounts as acc}
              <option value={String(acc.id)}>{acc.name}</option>
            {/each}
          </select>
        </div>
      {/if}
      <button type="submit" class="ahq-btn-primary w-full" disabled={costForm.processing}>
        Adicionar custo
      </button>
    </form>
  </section>
</AppShell>
