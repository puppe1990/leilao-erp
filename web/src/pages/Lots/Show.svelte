<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let lot = {}
  export let items = []
  export let costs = []
  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}

  let costForm = useForm({
    cost_label: 'Frete',
    cost_amount: '',
    already_paid: false,
    cash_account_id: cashAccounts[0]?.id?.toString() || '',
  })

  function submitCost() {
    $costForm.post(`/lots/${lot.id}/costs`)
  }
</script>

<div class="max-w-4xl mx-auto p-6 space-y-8">
  <div class="flex items-start justify-between gap-4">
    <div>
      <h1 class="text-2xl font-semibold text-stone-800">{lot.name}</h1>
      <p class="text-sm text-stone-600 mt-1">
        Compra em {lot.purchasedAt} · {lot.statusLabel} · Custo total {lot.totalCost}
      </p>
    </div>
    <a href="/lots" use:inertia class="text-sm underline">← Lotes</a>
  </div>

  <section>
    <h2 class="text-lg font-medium text-stone-800 mb-3">Itens</h2>
    {#if items.length === 0}
      <p class="text-sm text-stone-500">Nenhum item.</p>
    {:else}
      <div class="overflow-x-auto border rounded">
        <table class="w-full text-sm text-left">
          <thead class="bg-stone-100 text-stone-600">
            <tr>
              <th class="p-3 font-medium">#</th>
              <th class="p-3 font-medium">Título</th>
              <th class="p-3 font-medium">Custo unitário</th>
              <th class="p-3 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {#each items as item}
              <tr class="border-t">
                <td class="p-3 text-stone-500">{item.id}</td>
                <td class="p-3">{item.title}</td>
                <td class="p-3">{item.unitCost}</td>
                <td class="p-3">{item.statusLabel}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>

  <section>
    <h2 class="text-lg font-medium text-stone-800 mb-3">Custos</h2>
    {#if costs.length === 0}
      <p class="text-sm text-stone-500">Nenhum custo.</p>
    {:else}
      <ul class="border rounded divide-y">
        {#each costs as cost}
          <li class="p-3 flex justify-between text-sm">
            <span>{cost.label}</span>
            <span class="font-medium">{cost.amount}</span>
          </li>
        {/each}
      </ul>
    {/if}

    {#if payables.length > 0}
      <h3 class="text-sm font-medium text-stone-700 mt-4 mb-2">A pagar / pagos</h3>
      <ul class="text-sm space-y-1">
        {#each payables as p}
          <li class="flex justify-between border-b py-1">
            <span>{p.description} · {p.statusLabel}</span>
            <span>{p.amount}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  <section class="border rounded p-4 bg-stone-50">
    <h2 class="text-lg font-medium text-stone-800 mb-3">Adicionar custo</h2>
    {#if errors.form}
      <p class="mb-2 text-red-700 text-sm">{errors.form}</p>
    {/if}
    <form on:submit|preventDefault={submitCost} class="space-y-3 max-w-md">
      <div>
        <label class="block text-sm text-stone-600 mb-1" for="cost_label">Rótulo</label>
        <input
          id="cost_label"
          type="text"
          bind:value={$costForm.cost_label}
          class="block w-full border p-2 rounded bg-white"
        />
      </div>
      <div>
        <label class="block text-sm text-stone-600 mb-1" for="cost_amount">Valor (R$)</label>
        <input
          id="cost_amount"
          type="text"
          bind:value={$costForm.cost_amount}
          class="block w-full border p-2 rounded bg-white"
          placeholder="50,00"
        />
        {#if errors.cost_amount}<p class="text-red-600 text-sm mt-1">{errors.cost_amount}</p>{/if}
      </div>
      <label class="flex items-center gap-2 text-sm">
        <input type="checkbox" bind:checked={$costForm.already_paid} />
        Já paguei
      </label>
      {#if $costForm.already_paid}
        <div>
          <label class="block text-sm text-stone-600 mb-1" for="cash_account_id">Conta</label>
          <select
            id="cash_account_id"
            bind:value={$costForm.cash_account_id}
            class="block w-full border p-2 rounded bg-white"
          >
            <option value="">Selecione…</option>
            {#each cashAccounts as acc}
              <option value={String(acc.id)}>{acc.name}</option>
            {/each}
          </select>
          {#if errors.cash_account_id}
            <p class="text-red-600 text-sm mt-1">{errors.cash_account_id}</p>
          {/if}
        </div>
      {/if}
      <button
        type="submit"
        class="px-4 py-2 bg-stone-800 text-white rounded text-sm"
        disabled={$costForm.processing}
      >
        Adicionar custo
      </button>
    </form>
  </section>
</div>
