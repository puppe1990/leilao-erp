<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let errors = {}
  export let cashAccounts = []
  export let site = {}

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

<div class="max-w-lg mx-auto p-6">
  <h1 class="text-2xl font-semibold text-stone-800 mb-6">Novo lote</h1>

  {#if errors.form}
    <p class="mb-4 text-red-700 text-sm">{errors.form}</p>
  {/if}

  <form on:submit|preventDefault={submit} class="space-y-4">
    <div>
      <label class="block text-sm text-stone-600 mb-1" for="name">Nome do lote</label>
      <input
        id="name"
        type="text"
        bind:value={form.name}
        class="block w-full border p-2 rounded"
        placeholder="Ex: Monitores — leilão Jul/2026"
      />
      {#if errors.name}<p class="text-red-600 text-sm mt-1">{errors.name}</p>{/if}
    </div>

    <div>
      <label class="block text-sm text-stone-600 mb-1" for="purchased_at">Data da compra</label>
      <input
        id="purchased_at"
        type="date"
        bind:value={form.purchased_at}
        class="block w-full border p-2 rounded"
      />
      {#if errors.purchased_at}<p class="text-red-600 text-sm mt-1">{errors.purchased_at}</p>{/if}
    </div>

    <div>
      <label class="block text-sm text-stone-600 mb-1" for="item_title">Título do item</label>
      <input
        id="item_title"
        type="text"
        bind:value={form.item_title}
        class="block w-full border p-2 rounded"
        placeholder="Ex: Monitor"
      />
      {#if errors.item_title}<p class="text-red-600 text-sm mt-1">{errors.item_title}</p>{/if}
    </div>

    <div>
      <label class="block text-sm text-stone-600 mb-1" for="item_qty">Quantidade</label>
      <input
        id="item_qty"
        type="number"
        min="1"
        bind:value={form.item_qty}
        class="block w-full border p-2 rounded"
      />
      {#if errors.item_qty}<p class="text-red-600 text-sm mt-1">{errors.item_qty}</p>{/if}
    </div>

    <div class="border-t pt-4">
      <h2 class="font-medium text-stone-700 mb-3">Custo (arremate)</h2>
      <div class="mb-3">
        <label class="block text-sm text-stone-600 mb-1" for="cost_label">Rótulo</label>
        <input
          id="cost_label"
          type="text"
          bind:value={form.cost_label}
          class="block w-full border p-2 rounded"
        />
      </div>
      <div class="mb-3">
        <label class="block text-sm text-stone-600 mb-1" for="cost_amount">Valor (R$)</label>
        <input
          id="cost_amount"
          type="text"
          bind:value={form.cost_amount}
          class="block w-full border p-2 rounded"
          placeholder="603,00"
        />
        {#if errors.cost_amount}<p class="text-red-600 text-sm mt-1">{errors.cost_amount}</p>{/if}
      </div>
      <label class="flex items-center gap-2 text-sm text-stone-700">
        <input type="checkbox" bind:checked={form.already_paid} />
        Já paguei
      </label>
      {#if form.already_paid}
        <div class="mt-3">
          <label class="block text-sm text-stone-600 mb-1" for="cash_account_id">Conta de caixa</label>
          <select
            id="cash_account_id"
            bind:value={form.cash_account_id}
            class="block w-full border p-2 rounded"
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
    </div>

    <div class="flex gap-3 pt-2">
      <button type="submit" class="px-4 py-2 bg-stone-800 text-white rounded" disabled={form.processing}>
        Salvar lote
      </button>
      <a href="/lots" use:inertia class="px-4 py-2 border rounded text-stone-700">Cancelar</a>
    </div>
  </form>
</div>
