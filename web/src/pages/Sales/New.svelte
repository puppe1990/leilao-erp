<script>
  import { useForm, inertia } from '@inertiajs/svelte'
  export let errors = {}
  export let items = []
  export let cashAccounts = []
  export let channels = []
  export let site = {}

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
    $form.post('/sales')
  }
</script>

<div class="max-w-lg mx-auto p-6">
  <h1 class="text-2xl font-semibold text-stone-800 mb-6">Nova venda</h1>

  {#if errors.form}
    <p class="mb-4 text-red-700 text-sm">{errors.form}</p>
  {/if}

  {#if items.length === 0}
    <div class="border border-dashed border-stone-300 rounded p-6 text-stone-600 text-sm">
      <p class="mb-2">Nenhum item em estoque para vender.</p>
      <a href="/lots" use:inertia class="underline text-amber-900">Ver lotes</a>
    </div>
  {:else}
    <form on:submit|preventDefault={submit} class="space-y-4">
      <div>
        <label class="block text-sm text-stone-600 mb-1" for="item_id">Item</label>
        <select
          id="item_id"
          bind:value={$form.item_id}
          class="block w-full border p-2 rounded"
        >
          <option value="">Selecione…</option>
          {#each items as it}
            <option value={String(it.id)}>
              #{it.id} — {it.title} (custo {it.unitCost})
            </option>
          {/each}
        </select>
        {#if errors.item_id}<p class="text-red-600 text-sm mt-1">{errors.item_id}</p>{/if}
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="channel">Canal</label>
        <select
          id="channel"
          bind:value={$form.channel}
          class="block w-full border p-2 rounded"
        >
          {#each channels as ch}
            <option value={ch.value}>{ch.label}</option>
          {/each}
        </select>
        {#if errors.channel}<p class="text-red-600 text-sm mt-1">{errors.channel}</p>{/if}
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="sold_at">Data da venda</label>
        <input
          id="sold_at"
          type="date"
          bind:value={$form.sold_at}
          class="block w-full border p-2 rounded"
        />
      </div>

      <div>
        <label class="block text-sm text-stone-600 mb-1" for="gross">Valor bruto (R$)</label>
        <input
          id="gross"
          type="text"
          bind:value={$form.gross}
          class="block w-full border p-2 rounded"
          placeholder="150,00"
        />
        {#if errors.gross}<p class="text-red-600 text-sm mt-1">{errors.gross}</p>{/if}
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-sm text-stone-600 mb-1" for="fee">Taxa (R$)</label>
          <input
            id="fee"
            type="text"
            bind:value={$form.fee}
            class="block w-full border p-2 rounded"
            placeholder="0,00"
          />
          {#if errors.fee}<p class="text-red-600 text-sm mt-1">{errors.fee}</p>{/if}
        </div>
        <div>
          <label class="block text-sm text-stone-600 mb-1" for="shipping">Frete (R$)</label>
          <input
            id="shipping"
            type="text"
            bind:value={$form.shipping}
            class="block w-full border p-2 rounded"
            placeholder="0,00"
          />
          {#if errors.shipping}<p class="text-red-600 text-sm mt-1">{errors.shipping}</p>{/if}
        </div>
      </div>

      <fieldset class="border rounded p-3 space-y-2">
        <legend class="text-sm text-stone-600 px-1">Pagamento</legend>
        <label class="flex items-center gap-2 text-sm text-stone-700">
          <input type="radio" bind:group={$form.payment_status} value="received" />
          Recebi agora
        </label>
        <label class="flex items-center gap-2 text-sm text-stone-700">
          <input type="radio" bind:group={$form.payment_status} value="pending" />
          A receber
        </label>
        {#if errors.payment_status}
          <p class="text-red-600 text-sm">{errors.payment_status}</p>
        {/if}

        {#if $form.payment_status === 'received'}
          <div class="pt-2">
            <label class="block text-sm text-stone-600 mb-1" for="cash_account_id">Conta de caixa</label>
            <select
              id="cash_account_id"
              bind:value={$form.cash_account_id}
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

        {#if $form.payment_status === 'pending'}
          <div class="pt-2">
            <label class="block text-sm text-stone-600 mb-1" for="due_on">Vencimento</label>
            <input
              id="due_on"
              type="date"
              bind:value={$form.due_on}
              class="block w-full border p-2 rounded"
            />
            {#if errors.due_on}<p class="text-red-600 text-sm mt-1">{errors.due_on}</p>{/if}
          </div>
        {/if}
      </fieldset>

      <div class="flex gap-3 pt-2">
        <button
          type="submit"
          class="px-4 py-2 bg-stone-800 text-white rounded"
          disabled={$form.processing}
        >
          Salvar venda
        </button>
        <a href="/sales" use:inertia class="px-4 py-2 border rounded text-stone-700">Cancelar</a>
      </div>
    </form>
  {/if}
</div>
