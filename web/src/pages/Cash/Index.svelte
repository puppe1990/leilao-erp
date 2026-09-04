<script>
  import { useForm, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let balances = []
  export let statement = []
  export let cashAccounts = []
  export let selectedAccountId = 0
  export let selectedName = ''
  export let selectedBalance = ''
  export let categories = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  const accountList = Array.isArray(balances) ? balances : []
  const rows = Array.isArray(statement) ? statement : []

  let form = useForm({
    account_id: selectedAccountId ? String(selectedAccountId) : '',
    direction: 'in',
    category: 'ajuste',
    amount: '',
    memo: '',
    occurred_at: new Date().toISOString().slice(0, 10),
  })

  let accountForm = useForm({
    name: '',
    kind: 'pix',
    opening_balance: '0,00',
  })

  let showNewAccount = false
  let editingEntryId = null
  let editEntry = {
    account_id: '',
    direction: 'out',
    category: 'despesa',
    amount: '',
    memo: '',
    occurred_at: '',
  }

  const directionOptions = [
    { value: 'in', label: 'Entrada' },
    { value: 'out', label: 'Saída' },
  ]

  function selectAccount(id) {
    window.location = `/cash?account_id=${id}`
  }

  function submitEntry() {
    form.account_id = selectedAccountId ? String(selectedAccountId) : form.account_id
    form.post('/cash/entries', {
      onSuccess: () => form.reset('amount', 'memo'),
    })
  }

  function submitAccount() {
    accountForm.post('/cash/accounts', {
      onSuccess: () => {
        accountForm.reset('name', 'opening_balance')
        showNewAccount = false
      },
    })
  }

  function startEditEntry(e) {
    editingEntryId = e.id
    editEntry = {
      account_id: String(e.accountId || ''),
      direction: e.direction || 'out',
      category: e.category || 'despesa',
      amount: e.amountRaw || '',
      memo: e.memo || '',
      occurred_at: e.occurredDate || (e.occurredAt || '').slice(0, 10),
    }
  }

  function cancelEditEntry() {
    editingEntryId = null
  }

  function saveEntry(id) {
    router.post(`/cash/entries/${id}`, { ...editEntry }, {
      onSuccess: () => {
        editingEntryId = null
      },
    })
  }

  async function deleteEntry(id) {
    const ok = await askConfirm({
      title: 'Excluir lançamento',
      message: 'Excluir este lançamento do extrato?',
      detail: 'O saldo da conta será recalculado.',
      confirmLabel: 'Excluir',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/cash/entries/${id}/delete`)
  }

  function lineDate(line) {
    const raw = line.occurredAt || ''
    return raw.slice ? raw.slice(0, 10) : raw
  }
</script>

<AppShell {companyName} active="cash">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Contas correntes</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Extrato e saldo de cada conta, como no banco.
    </p>
    <div class="mt-3"><Nav active="cash" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <div class="md:flex md:gap-6 md:items-start">
    <aside class="md:w-56 shrink-0 mb-6 md:mb-0">
      <p class="ahq-label mb-2">Contas</p>
      <div class="flex flex-col gap-1">
        {#each accountList as b}
          <button
            type="button"
            class="text-left ahq-card p-3 transition-colors
              {String(b.id) === String(selectedAccountId)
              ? 'border-secondary'
              : 'hover:bg-surface-container-low'}"
            on:click={() => selectAccount(b.id)}
          >
            <p class="font-semibold text-primary truncate">{b.name}</p>
            <p class="font-mono text-sm mt-0.5">{b.balance}</p>
          </button>
        {:else}
          <p class="text-sm text-on-surface-variant">Nenhuma conta ainda.</p>
        {/each}
      </div>
      <button
        type="button"
        class="mt-3 text-sm text-secondary font-medium"
        on:click={() => (showNewAccount = !showNewAccount)}
      >
        {showNewAccount ? 'Cancelar' : '+ Nova conta'}
      </button>
      {#if showNewAccount}
        <form on:submit|preventDefault={submitAccount} class="ahq-card p-3 mt-2 space-y-2">
          <input class="ahq-input h-9 text-sm" bind:value={accountForm.name} placeholder="Nome (ex: Nubank)" />
          {#if errors.name}<p class="text-error text-xs">{errors.name}</p>{/if}
          <input
            class="ahq-input h-9 text-sm font-mono"
            bind:value={accountForm.opening_balance}
            placeholder="Saldo inicial"
          />
          <button type="submit" class="ahq-btn-primary h-9 w-full text-xs" disabled={accountForm.processing}>
            Criar
          </button>
        </form>
      {/if}
    </aside>

    <div class="flex-1 min-w-0">
      {#if !selectedAccountId}
        <div class="ahq-card p-8 text-center text-on-surface-variant border-dashed">
          Crie uma conta corrente para ver o extrato.
        </div>
      {:else}
        <div class="ahq-card p-5 mb-section-padding">
          <p class="ahq-label">Saldo atual</p>
          <h2 class="font-headline-md text-headline-md text-primary mt-1">{selectedName}</h2>
          <p class="ahq-value mt-2">{selectedBalance}</p>
        </div>

        <section class="ahq-card p-5 mb-section-padding">
          <h3 class="font-semibold text-primary mb-3">Novo lançamento</h3>
          <form on:submit|preventDefault={submitEntry} class="grid gap-3 sm:grid-cols-2">
            <div>
              <label class="ahq-label block mb-1" for="direction">Tipo</label>
              <SearchableSelect
                id="direction"
                bind:value={form.direction}
                options={directionOptions}
                allowClear={false}
              />
            </div>
            <div>
              <label class="ahq-label block mb-1" for="amount">Valor (R$)</label>
              <input id="amount" class="ahq-input font-mono" bind:value={form.amount} placeholder="0,00" />
              {#if errors.amount}<p class="text-error text-xs mt-1">{errors.amount}</p>{/if}
            </div>
            <div>
              <label class="ahq-label block mb-1" for="occurred_at">Data</label>
              <input id="occurred_at" type="date" class="ahq-input font-mono" bind:value={form.occurred_at} />
            </div>
            <div>
              <label class="ahq-label block mb-1" for="memo">Histórico</label>
              <input id="memo" class="ahq-input" bind:value={form.memo} placeholder="PIX recebido, Uber…" />
            </div>
            <div class="sm:col-span-2">
              <button type="submit" class="ahq-btn-primary" disabled={form.processing}>Lançar no extrato</button>
            </div>
          </form>
        </section>

        <section>
          <h3 class="font-headline-md text-headline-md text-primary mb-3">Extrato</h3>
          <div class="ahq-card overflow-x-auto">
            <table class="w-full text-sm min-w-[640px]">
              <thead>
                <tr class="text-left border-b border-outline-variant">
                  <th class="ahq-label px-3 py-2">Data</th>
                  <th class="ahq-label px-3 py-2">Histórico</th>
                  <th class="ahq-label px-3 py-2 text-right">Entrada</th>
                  <th class="ahq-label px-3 py-2 text-right">Saída</th>
                  <th class="ahq-label px-3 py-2 text-right">Saldo</th>
                  <th class="px-3 py-2"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-outline-variant">
                {#each rows as line (line.kind + '-' + line.entryId)}
                  <tr class="align-top">
                    {#if editingEntryId && line.entry && editingEntryId === line.entry.id}
                      <td colspan="6" class="p-3">
                        <div class="grid gap-2 sm:grid-cols-2">
                          <SearchableSelect
                            bind:value={editEntry.direction}
                            options={directionOptions}
                            allowClear={false}
                          />
                          <input class="ahq-input h-9 font-mono text-sm" bind:value={editEntry.amount} />
                          <input type="date" class="ahq-input h-9 font-mono text-sm" bind:value={editEntry.occurred_at} />
                          <input class="ahq-input h-9 text-sm" bind:value={editEntry.memo} placeholder="Histórico" />
                          <div class="sm:col-span-2 flex gap-3 justify-end">
                            <button
                              type="button"
                              class="text-secondary text-sm font-medium"
                              on:click={() => saveEntry(line.entry.id)}
                            >
                              Salvar
                            </button>
                            <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEditEntry}>
                              Cancelar
                            </button>
                          </div>
                        </div>
                      </td>
                    {:else}
                      <td class="px-3 py-2 font-mono text-on-surface-variant whitespace-nowrap">
                        {line.kind === 'opening' ? '—' : lineDate(line)}
                      </td>
                      <td class="px-3 py-2 text-primary">{line.description}</td>
                      <td class="px-3 py-2 text-right font-mono text-secondary">{line.credit || ''}</td>
                      <td class="px-3 py-2 text-right font-mono text-error">{line.debit || ''}</td>
                      <td class="px-3 py-2 text-right font-mono font-semibold">{line.balance}</td>
                      <td class="px-3 py-2 text-right whitespace-nowrap">
                        {#if line.canEdit}
                          <button
                            type="button"
                            class="text-xs text-secondary font-medium mr-2"
                            on:click={() => startEditEntry(line.entry)}
                          >
                            Editar
                          </button>
                        {/if}
                        {#if line.canDelete}
                          <button
                            type="button"
                            class="text-xs text-error"
                            on:click={() => deleteEntry(line.entry.id)}
                          >
                            Excluir
                          </button>
                        {/if}
                      </td>
                    {/if}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
      {/if}
    </div>
  </div>
</AppShell>
