<script>
  import { useForm, inertia } from '@inertiajs/svelte'
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

<div class="max-w-5xl mx-auto p-6">
  <h1 class="text-2xl font-semibold text-stone-800 mb-6">Caixa</h1>

  {#if errors.form}
    <p class="mb-4 text-red-700 text-sm">{errors.form}</p>
  {/if}

  <!-- Saldos -->
  <section class="mb-8">
    <h2 class="text-sm font-medium text-stone-600 mb-3">Saldos</h2>
    {#if balances.length === 0}
      <p class="text-stone-500 text-sm">Nenhuma conta de caixa cadastrada.</p>
    {:else}
      <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3">
        {#each balances as b}
          <div class="border rounded p-4 bg-stone-50">
            <p class="text-sm text-stone-600">{b.name}</p>
            <p class="text-xl font-semibold text-stone-800 mt-1">{b.balance}</p>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <!-- Lançamento manual (ajuste) -->
  <section class="mb-8 border rounded p-4">
    <h2 class="text-sm font-medium text-stone-600 mb-3">Lançamento manual (ajuste)</h2>
    <form on:submit|preventDefault={submit} class="grid gap-3 sm:grid-cols-2">
      <div>
        <label class="block text-sm text-stone-600 mb-1" for="account_id">Conta</label>
        <select
          id="account_id"
          bind:value={form.account_id}
          class="block w-full border p-2 rounded"
        >
          <option value="">Selecione…</option>
          {#each cashAccounts as a}
            <option value={String(a.id)}>{a.name}</option>
          {/each}
        </select>
        {#if errors.account_id}<p class="text-red-600 text-sm mt-1">{errors.account_id}</p>{/if}
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="direction">Direção</label>
        <select
          id="direction"
          bind:value={form.direction}
          class="block w-full border p-2 rounded"
        >
          <option value="in">Entrada</option>
          <option value="out">Saída</option>
        </select>
        {#if errors.direction}<p class="text-red-600 text-sm mt-1">{errors.direction}</p>{/if}
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="amount">Valor (R$)</label>
        <input
          id="amount"
          type="text"
          bind:value={form.amount}
          placeholder="0,00"
          class="block w-full border p-2 rounded"
        />
        {#if errors.amount}<p class="text-red-600 text-sm mt-1">{errors.amount}</p>{/if}
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="occurred_at">Data</label>
        <input
          id="occurred_at"
          type="date"
          bind:value={form.occurred_at}
          class="block w-full border p-2 rounded"
        />
      </div>

      <div class="sm:col-span-2">
        <label class="block text-sm text-stone-600 mb-1" for="memo">Memo</label>
        <input
          id="memo"
          type="text"
          bind:value={form.memo}
          placeholder="Opcional"
          class="block w-full border p-2 rounded"
        />
      </div>

      <div class="sm:col-span-2">
        <button
          type="submit"
          class="px-4 py-2 bg-stone-800 text-white text-sm rounded"
          disabled={form.processing}
        >
          Registrar ajuste
        </button>
      </div>
    </form>
  </section>

  <!-- Extrato -->
  <section>
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-medium text-stone-600">Extrato</h2>
      {#if cashAccounts.length > 0}
        <form method="get" action="/cash" class="flex items-center gap-2 text-sm">
          <label for="filter_account" class="text-stone-500">Conta</label>
          <select
            id="filter_account"
            name="account_id"
            class="border p-1 rounded"
            on:change={(e) => {
              const v = e.currentTarget.value
              window.location = v ? `/cash?account_id=${v}` : '/cash'
            }}
          >
            <option value="" selected={!(filterAccountId > 0)}>Todas</option>
            {#each cashAccounts as a}
              <option value={String(a.id)} selected={filterAccountId === a.id}>{a.name}</option>
            {/each}
          </select>
        </form>
      {/if}
    </div>

    {#if entries.length === 0}
      <div class="border border-dashed border-stone-300 rounded p-8 text-center text-stone-600">
        Nenhum lançamento no caixa.
      </div>
    {:else}
      <div class="overflow-x-auto border rounded">
        <table class="w-full text-sm text-left">
          <thead class="bg-stone-100 text-stone-600">
            <tr>
              <th class="p-3 font-medium">Data</th>
              <th class="p-3 font-medium">Conta</th>
              <th class="p-3 font-medium">Direção</th>
              <th class="p-3 font-medium">Categoria</th>
              <th class="p-3 font-medium">Valor</th>
              <th class="p-3 font-medium">Memo</th>
            </tr>
          </thead>
          <tbody>
            {#each entries as e}
              <tr class="border-t hover:bg-stone-50">
                <td class="p-3">{e.occurredAt?.slice?.(0, 10) || e.occurredAt}</td>
                <td class="p-3">{e.accountName}</td>
                <td class="p-3">
                  <span class={e.direction === 'in' ? 'text-green-700' : 'text-red-700'}>
                    {e.directionLabel}
                  </span>
                </td>
                <td class="p-3">{e.categoryLabel}</td>
                <td class="p-3 font-medium">{e.amount}</td>
                <td class="p-3 text-stone-500">{e.memo || '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>

  <Nav active="cash" />
</div>
