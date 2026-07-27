<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let errors = {}
  export let cashAccounts = []
  export let site = {}
  export let companyName = 'AuctionHQ'

  let form = useForm({
    name: '',
    purchased_at: new Date().toISOString().slice(0, 10),
    item_title: '',
    item_qty: '1',
    cost_label: 'Arremate',
    cost_amount: '',
    already_paid: false,
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
  })

  function submit() {
    form.post('/lots')
  }
</script>

<AppShell {companyName} active="lots">
  <div class="mb-section-padding">
    <a href="/lots" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Lotes
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Novo lote</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Registre a compra no leilão e o rateio dos itens.</p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 border-error/30 bg-error-container/30">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4">
    <div>
      <label class="ahq-label block mb-1.5" for="name">Nome do lote</label>
      <input
        id="name"
        type="text"
        bind:value={form.name}
        class="ahq-input"
        placeholder="Ex: Monitores — leilão Jul/2026"
      />
      {#if errors.name}<p class="text-error text-sm mt-1">{errors.name}</p>{/if}
    </div>

    <div>
      <label class="ahq-label block mb-1.5" for="purchased_at">Data da compra</label>
      <input id="purchased_at" type="date" bind:value={form.purchased_at} class="ahq-input font-mono" />
      {#if errors.purchased_at}<p class="text-error text-sm mt-1">{errors.purchased_at}</p>{/if}
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="col-span-2 sm:col-span-1">
        <label class="ahq-label block mb-1.5" for="item_title">Título do item</label>
        <input
          id="item_title"
          type="text"
          bind:value={form.item_title}
          class="ahq-input"
          placeholder="Ex: Monitor"
        />
        {#if errors.item_title}<p class="text-error text-sm mt-1">{errors.item_title}</p>{/if}
      </div>
      <div class="col-span-2 sm:col-span-1">
        <label class="ahq-label block mb-1.5" for="item_qty">Quantidade</label>
        <input
          id="item_qty"
          type="number"
          min="1"
          bind:value={form.item_qty}
          class="ahq-input font-mono"
        />
        {#if errors.item_qty}<p class="text-error text-sm mt-1">{errors.item_qty}</p>{/if}
      </div>
    </div>

    <div class="border-t border-outline-variant pt-4">
      <h2 class="font-semibold text-primary mb-3">Custo (arremate)</h2>
      <div class="space-y-3">
        <div>
          <label class="ahq-label block mb-1.5" for="cost_label">Rótulo</label>
          <input id="cost_label" type="text" bind:value={form.cost_label} class="ahq-input" />
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="cost_amount">Valor (R$)</label>
          <input
            id="cost_amount"
            type="text"
            bind:value={form.cost_amount}
            class="ahq-input font-mono"
            placeholder="603,00"
          />
          {#if errors.cost_amount}<p class="text-error text-sm mt-1">{errors.cost_amount}</p>{/if}
        </div>
        <label class="flex items-center gap-2 text-body-md text-on-surface">
          <input type="checkbox" bind:checked={form.already_paid} class="rounded border-outline-variant text-secondary" />
          Já paguei
        </label>
        {#if form.already_paid}
          <div>
            <label class="ahq-label block mb-1.5" for="cash_account_id">Conta de caixa</label>
            <select id="cash_account_id" bind:value={form.cash_account_id} class="ahq-select">
              <option value="">Selecione…</option>
              {#each cashAccounts as acc}
                <option value={String(acc.id)}>{acc.name}</option>
              {/each}
            </select>
            {#if errors.cash_account_id}
              <p class="text-error text-sm mt-1">{errors.cash_account_id}</p>
            {/if}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex flex-col sm:flex-row gap-3 pt-2">
      <button type="submit" class="ahq-btn-primary flex-1" disabled={form.processing}>Salvar lote</button>
      <a href="/lots" use:inertia class="ahq-btn-ghost flex-1 text-center">Cancelar</a>
    </div>
  </form>
</AppShell>
