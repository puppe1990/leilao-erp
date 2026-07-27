<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

  export let errors = {}
  export let items = []
  export let cashAccounts = []
  export let channels = []
  export let site = {}
  export let companyName = 'AuctionHQ'

  let form = useForm({
    item_id: items[0]?.id?.toString() || '',
    channel: 'direct',
    gross: '',
    fee: '0',
    shipping: '0',
    payment_status: 'received',
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
    due_on: '',
    sold_at: new Date().toISOString().slice(0, 10),
  })

  function submit() {
    form.post('/sales')
  }
</script>

<AppShell {companyName} active="sales">
  <div class="mb-section-padding">
    <a href="/sales" use:inertia class="text-sm text-secondary flex items-center gap-1 mb-3">
      <span class="material-symbols-outlined text-[18px]">arrow_back</span>
      Vendas
    </a>
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Nova venda</h1>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if items.length === 0}
    <div class="ahq-card p-6 border-dashed text-on-surface-variant text-sm">
      <p class="mb-2">Nenhum item em estoque para vender.</p>
      <a href="/lots" use:inertia class="text-secondary font-medium">Ver lotes</a>
    </div>
  {:else}
    <form on:submit|preventDefault={submit} class="ahq-card p-5 space-y-4">
      <div>
        <label class="ahq-label block mb-1.5" for="item_id">Item</label>
        <select id="item_id" bind:value={form.item_id} class="ahq-select">
          <option value="">Selecione…</option>
          {#each items as it}
            <option value={String(it.id)}>#{it.id} — {it.title} (custo {it.unitCost})</option>
          {/each}
        </select>
        {#if errors.item_id}<p class="text-error text-sm mt-1">{errors.item_id}</p>{/if}
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="ahq-label block mb-1.5" for="channel">Canal</label>
          <select id="channel" bind:value={form.channel} class="ahq-select">
            {#each channels as ch}
              <option value={ch.value}>{ch.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="sold_at">Data</label>
          <input id="sold_at" type="date" bind:value={form.sold_at} class="ahq-input font-mono" />
        </div>
      </div>

      <div>
        <label class="ahq-label block mb-1.5" for="gross">Valor bruto (R$)</label>
        <input
          id="gross"
          type="text"
          bind:value={form.gross}
          class="ahq-input font-mono"
          placeholder="150,00"
        />
        {#if errors.gross}<p class="text-error text-sm mt-1">{errors.gross}</p>{/if}
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="ahq-label block mb-1.5" for="fee">Taxa (R$)</label>
          <input id="fee" type="text" bind:value={form.fee} class="ahq-input font-mono" placeholder="0,00" />
        </div>
        <div>
          <label class="ahq-label block mb-1.5" for="shipping">Frete (R$)</label>
          <input
            id="shipping"
            type="text"
            bind:value={form.shipping}
            class="ahq-input font-mono"
            placeholder="0,00"
          />
        </div>
      </div>

      <fieldset class="border border-outline-variant rounded-lg p-4 space-y-3">
        <legend class="ahq-label px-1">Pagamento</legend>
        <label class="flex items-center gap-2 text-body-md">
          <input type="radio" bind:group={form.payment_status} value="received" class="text-secondary" />
          Recebi agora
        </label>
        <label class="flex items-center gap-2 text-body-md">
          <input type="radio" bind:group={form.payment_status} value="pending" class="text-secondary" />
          A receber
        </label>

        {#if form.payment_status === 'received'}
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

        {#if form.payment_status === 'pending'}
          <div>
            <label class="ahq-label block mb-1.5" for="due_on">Vencimento</label>
            <input id="due_on" type="date" bind:value={form.due_on} class="ahq-input font-mono" />
            {#if errors.due_on}<p class="text-error text-sm mt-1">{errors.due_on}</p>{/if}
          </div>
        {/if}
      </fieldset>

      <div class="flex flex-col sm:flex-row gap-3 pt-2">
        <button type="submit" class="ahq-btn-primary flex-1" disabled={form.processing}>Salvar venda</button>
        <a href="/sales" use:inertia class="ahq-btn-ghost flex-1 text-center">Cancelar</a>
      </div>
    </form>
  {/if}
</AppShell>
