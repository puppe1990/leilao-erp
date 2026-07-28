<script>
  import { useForm, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'

  export let balances = []
  export let entries = []
  export let cashAccounts = []
  export let filterAccountId = 0
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let form = useForm({
    account_id: cashAccounts[0]?.id?.toString() || '',
    direction: 'out',
    amount: '',
    memo: '',
    occurred_at: new Date().toISOString().slice(0, 10),
  })

  let accountForm = useForm({
    name: '',
    kind: 'pix',
    opening_balance: '0,00',
  })

  let filterAccount = filterAccountId > 0 ? String(filterAccountId) : ''

  $: cashAccountOptions = (Array.isArray(cashAccounts) ? cashAccounts : []).map((a) => ({
    value: String(a.id),
    label: a.name,
  }))
  $: filterAccountOptions = [
    { value: '', label: 'Todas as contas' },
    ...cashAccountOptions,
  ]
  const kindOptions = [
    { value: 'pix', label: 'PIX' },
    { value: 'bank', label: 'Banco' },
    { value: 'cash', label: 'Dinheiro' },
    { value: 'other', label: 'Outro' },
  ]
  const directionOptions = [
    { value: 'in', label: 'Entrada' },
    { value: 'out', label: 'Saída' },
  ]

  function submitEntry() {
    form.post('/cash/entries')
  }

  function submitAccount() {
    accountForm.post('/cash/accounts', {
      onSuccess: () => accountForm.reset('name', 'opening_balance'),
    })
  }

  function deleteEntry(id) {
    if (!confirm('Excluir este ajuste manual?')) return
    router.post(`/cash/entries/${id}/delete`)
  }

  function deleteAccount(id) {
    if (!confirm('Excluir esta conta? Só funciona se não houver lançamentos.')) return
    router.post(`/cash/accounts/${id}/delete`)
  }

  function onFilterAccount(v) {
    filterAccount = v
    window.location = v ? `/cash?account_id=${v}` : '/cash'
  }
</script>

<AppShell {companyName} active="cash">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Caixa</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Contas, extrato e ajustes.</p>
    <div class="mt-3"><Nav active="cash" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <!-- Contas -->
  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Contas</h2>
    <div class="grid gap-3 sm:grid-cols-2 mb-4">
      {#each balances as b}
        <div class="ahq-card p-4">
          <div class="flex justify-between gap-2">
            <div>
              <p class="ahq-label">{b.kind}</p>
              <p class="font-semibold text-primary">{b.name}</p>
              <p class="ahq-value mt-1">{b.balance}</p>
              {#if b.opening}
                <p class="text-[10px] text-on-surface-variant mt-1">Abertura: {b.opening}</p>
              {/if}
            </div>
            <button
              type="button"
              class="text-error text-sm self-start"
              on:click={() => deleteAccount(b.id)}
              title="Excluir conta"
            >
              <span class="material-symbols-outlined">delete</span>
            </button>
          </div>
        </div>
      {/each}
    </div>

    <form on:submit|preventDefault={submitAccount} class="ahq-card p-4 grid gap-3 sm:grid-cols-3">
      <div>
        <label class="ahq-label block mb-1" for="acc_name">Nova conta</label>
        <input id="acc_name" class="ahq-input h-10" bind:value={accountForm.name} placeholder="Nome" />
        {#if errors.name}<p class="text-error text-xs mt-1">{errors.name}</p>{/if}
      </div>
      <div>
        <label class="ahq-label block mb-1" for="acc_kind">Tipo</label>
        <SearchableSelect
          id="acc_kind"
          bind:value={accountForm.kind}
          options={kindOptions}
          placeholder="Tipo"
          searchPlaceholder="Buscar tipo…"
          allowClear={false}
          buttonClass="ahq-select h-10 w-full text-left flex items-center justify-between gap-2"
        />
      </div>
      <div>
        <label class="ahq-label block mb-1" for="acc_open">Saldo inicial</label>
        <input id="acc_open" class="ahq-input h-10 font-mono" bind:value={accountForm.opening_balance} />
      </div>
      <button type="submit" class="sm:col-span-3 ahq-btn-primary h-10" disabled={accountForm.processing}>
        Criar conta
      </button>
    </form>
  </section>

  <!-- Lançamento manual -->
  <section class="ahq-card p-5 mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-3">Lançamento manual (ajuste)</h2>
    <form on:submit|preventDefault={submitEntry} class="grid gap-3 sm:grid-cols-2">
      <div>
        <label class="ahq-label block mb-1.5" for="account_id">Conta</label>
        <SearchableSelect
          id="account_id"
          bind:value={form.account_id}
          options={cashAccountOptions}
          placeholder="Selecione…"
          searchPlaceholder="Buscar conta…"
        />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="direction">Direção</label>
        <SearchableSelect
          id="direction"
          bind:value={form.direction}
          options={directionOptions}
          placeholder="Direção"
          searchPlaceholder="Buscar…"
          allowClear={false}
        />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="amount">Valor (R$)</label>
        <input id="amount" type="text" bind:value={form.amount} placeholder="0,00" class="ahq-input font-mono" />
      </div>
      <div>
        <label class="ahq-label block mb-1.5" for="occurred_at">Data</label>
        <input id="occurred_at" type="date" bind:value={form.occurred_at} class="ahq-input font-mono" />
      </div>
      <div class="sm:col-span-2">
        <label class="ahq-label block mb-1.5" for="memo">Memo</label>
        <input id="memo" type="text" bind:value={form.memo} class="ahq-input" />
      </div>
      <div class="sm:col-span-2">
        <button type="submit" class="ahq-btn-primary" disabled={form.processing}>Registrar ajuste</button>
      </div>
    </form>
  </section>

  <!-- Extrato -->
  <section>
    <div class="flex items-center justify-between mb-stack-gap gap-2 flex-wrap">
      <h2 class="font-headline-md text-headline-md text-primary">Extrato</h2>
      {#if cashAccounts.length > 0}
        <div class="min-w-[12rem]">
          <SearchableSelect
            id="filter_account"
            bind:value={filterAccount}
            options={filterAccountOptions}
            placeholder="Todas as contas"
            searchPlaceholder="Buscar conta…"
            allowClear={false}
            buttonClass="ahq-select h-10 w-full text-left flex items-center justify-between gap-2"
            onChange={onFilterAccount}
          />
        </div>
      {/if}
    </div>

    {#if entries.length === 0}
      <div class="ahq-card p-8 text-center text-on-surface-variant border-dashed">Nenhum lançamento.</div>
    {:else}
      <div class="ahq-card divide-y divide-outline-variant">
        {#each entries as e}
          <div class="p-4 flex gap-3 items-start">
            <div
              class="w-10 h-10 rounded-full flex items-center justify-center shrink-0
                {e.direction === 'in' ? 'bg-tertiary-fixed/20' : 'bg-error-container/60'}"
            >
              <span class="material-symbols-outlined text-[20px] {e.direction === 'in' ? 'text-on-tertiary-container' : 'text-error'}">
                {e.direction === 'in' ? 'arrow_downward' : 'arrow_upward'}
              </span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex justify-between gap-2">
                <p class="font-semibold text-primary truncate">{e.categoryLabel}</p>
                <p class="font-mono font-semibold shrink-0 {e.direction === 'in' ? 'text-on-tertiary-container' : 'text-error'}">
                  {e.direction === 'in' ? '+' : '−'}{e.amount}
                </p>
              </div>
              <p class="text-on-surface-variant text-sm">
                {e.occurredAt?.slice?.(0, 10) || e.occurredAt} · {e.accountName}
                {#if e.memo}<span> · {e.memo}</span>{/if}
              </p>
              {#if e.canDelete}
                <button type="button" class="text-xs text-error mt-1" on:click={() => deleteEntry(e.id)}>Excluir ajuste</button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
</AppShell>
