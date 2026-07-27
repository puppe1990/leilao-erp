<script>
  import { inertia, router } from '@inertiajs/svelte'
  export let sales = []
  export let errors = {}
  export let site = {}

  function cancelSale(id) {
    if (!confirm('Cancelar esta venda pendente? O item voltará ao estoque.')) return
    router.post(`/sales/${id}/cancel`)
  }
</script>

<div class="max-w-5xl mx-auto p-6">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-semibold text-stone-800">Vendas</h1>
    <a href="/sales/new" use:inertia class="px-4 py-2 bg-stone-800 text-white text-sm rounded">
      Nova venda
    </a>
  </div>

  {#if errors.form}
    <p class="mb-4 text-red-700 text-sm">{errors.form}</p>
  {/if}

  {#if sales.length === 0}
    <div class="border border-dashed border-stone-300 rounded p-8 text-center text-stone-600">
      <p class="mb-4">Nenhuma venda registrada.</p>
      <a href="/sales/new" use:inertia class="underline text-amber-900">
        Registrar primeira venda
      </a>
    </div>
  {:else}
    <div class="overflow-x-auto border rounded">
      <table class="w-full text-sm text-left">
        <thead class="bg-stone-100 text-stone-600">
          <tr>
            <th class="p-3 font-medium">Item</th>
            <th class="p-3 font-medium">Data</th>
            <th class="p-3 font-medium">Canal</th>
            <th class="p-3 font-medium">Bruto</th>
            <th class="p-3 font-medium">Taxa</th>
            <th class="p-3 font-medium">Líquido</th>
            <th class="p-3 font-medium">Pagamento</th>
            <th class="p-3 font-medium"></th>
          </tr>
        </thead>
        <tbody>
          {#each sales as sale}
            <tr class="border-t hover:bg-stone-50">
              <td class="p-3">{sale.itemTitle || `#${sale.itemId}`}</td>
              <td class="p-3">{sale.soldAt?.slice?.(0, 10) || sale.soldAt}</td>
              <td class="p-3">{sale.channelLabel}</td>
              <td class="p-3">{sale.gross}</td>
              <td class="p-3">{sale.fee}</td>
              <td class="p-3">{sale.net}</td>
              <td class="p-3">{sale.paymentLabel}</td>
              <td class="p-3 text-right">
                {#if sale.canCancel}
                  <button
                    type="button"
                    class="text-sm text-red-700 underline"
                    on:click={() => cancelSale(sale.id)}
                  >
                    Cancelar
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <nav class="mt-6 text-sm space-x-4">
    <a href="/dashboard" use:inertia class="underline">Dashboard</a>
    <a href="/lots" use:inertia class="underline">Lotes</a>
    <a href="/cash" use:inertia class="underline">Caixa</a>
    <a href="/payables" use:inertia class="underline">A pagar</a>
    <a href="/receivables" use:inertia class="underline">A receber</a>
  </nav>
</div>
