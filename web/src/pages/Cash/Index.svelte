<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'

  export let balances = []
  export let entries = []
  export let cashAccounts = []
  export let filterAccountId = 0
  export let errors = {}
  export let site = {}

  let form = useForm({
    account_id: cashAccounts[0]?.id?.toString() || '',
    direction: 'out',
    amount: '',
    memo: '',
    occurred_at: new Date().toISOString().slice(0, 10),
  })

  function submit() {
    form.post('/cash/entries')
  }
</script>

<AppShell active="cash">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Relatórios Financeiros</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Caixa, extrato e ajustes.</p>
    <div class="mt-3">
      <Nav active="cash" />
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Saldos</h2>
    {#if balances.length === 0}
      <p class="text-on-surface-variant text-sm">Nenhuma conta de caixa cadastrada.</p>
    {:else}
      <div class="grid gap-3 grid-cols-1 sm:grid-cols-2 md:grid-cols-3">
        {#each balances as b}
          <div class="ahq-card p-4">
            <p class="ahq-label">{b.name}</p>
            <p class="ahq-value mt-1">{b.balance}</p>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <section class="ahq-card p-5 mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-3">Lançamento manual</h2>
    <p class="text-on-surface-variant text-sm mb-4">Categoria fixa: ajuste.</p>
    <form on:submit|preventDefault={submit} class="grid gap-3 sm:grid-cols-2">
      <div>
        <label class="ahq-label block mb-1.5" for="account_id">Conta</label>
        <select id="account_id" bind:value={form.account_id} class="ahq-select">
          <option value="">Selecione…</option>
          {#each cashAccounts as a}
            <option value={String(a.id)}>{a.name}</option>
          {/each}
        </select>
        {#if errors.account_id}<p class="text-error text-sm mt-1">{errors.account_id}</p>{/if}
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="direction">Direção</label>
        <select id="direction" bind:value={form.direction} class="ahq-select">
          <option value="in">Entrada</option>
          <option value="out">Saída</option>
        </select>
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="amount">Valor (R$)</label>
        <input
          id="amount"
          type="text"
          bind:value={form.amount}
          placeholder="0,00"
          class="ahq-input font-mono"
        />
        {#if errors.amount}<p class="text-error text-sm mt-1">{errors.amount}</p>{/if}
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="occurred_at">Data</label>
        <input id="occurred_at" type="date" bind:value={form.occurred_at} class="ahq-input font-mono" />
      </div>
      <div class="sm:col-span-2">
        <label class="ahq-label block mb-1.5" for="memo">Memo</label>
        <input id="memo" type="text" bind:value={form.memo} placeholder="Opcional" class="ahq-input" />
      </div>
      <div class="sm:col-span-2">
        <button type="submit" class="ahq-btn-primary" disabled={form.processing}>Registrar ajuste</button>
      </div>
    </form>
  </section>

  <section>
    <div class="flex items-center justify-between mb-stack-gap gap-2 flex-wrap">
      <h2 class="font-headline-md text-headline-md text-primary">Extrato</h2>
      {#if cashAccounts.length > 0}
        <select
          class="ahq-select h-10 w-auto min-w-[10rem]"
          on:change={(e) => {
            const v = e.currentTarget.value
            window.location = v ? `/cash?account_id=${v}` : '/cash'
          }}
        >
          <option value="" selected={!(filterAccountId > 0)}>Todas as contas</option>
          {#each cashAccounts as a}
            <option value={String(a.id)} selected={filterAccountId === a.id}>{a.name}</option>
          {/each}
        </select>
      {/if}
    </div>

    {#if entries.length === 0}
      <div class="ahq-card p-8 text-center text-on-surface-variant border-dashed">
        Nenhum lançamento no caixa.
      </div>
    {:else}
      <div class="ahq-card divide-y divide-outline-variant">
        {#each entries as e}
          <div class="p-4 flex gap-3 items-start">
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center shrink-0
                {e.direction === 'in' ? 'bg-tertiary-fixed/20' : 'bg-error-container/60'}"
            >
              <span
                class="material-symbols-outlined text-[20px]
                  {e.direction === 'in' ? 'text-on-tertiary-container' : 'text-error'}"
              >
                {e.direction === 'in' ? 'arrow_downward' : 'arrow_upward'}
              </span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex justify-between gap-2">
                <p class="font-semibold text-primary truncate">{e.categoryLabel}</p>
                <p
                  class="font-mono font-semibold shrink-0
                    {e.direction === 'in' ? 'text-on-tertiary-container' : 'text-error'}"
                >
                  {e.direction === 'in' ? '+' : '−'}{e.amount}
                </p>
              </div>
              <p class="text-on-surface-variant text-sm">
                {e.occurredAt?.slice?.(0, 10) || e.occurredAt} · {e.accountName}
                {#if e.memo}<span> · {e.memo}</span>{/if}
              </p>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
</AppShell>
