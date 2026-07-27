<script>
  import { inertia, router } from '@inertiajs/svelte'
  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}

  function settle(id) {
    if (!confirm('Quitar este pagamento? Será gerada uma saída no caixa.')) return
    const accountId = cashAccounts[0]?.id
    if (!accountId) {
      alert('Cadastre uma conta de caixa antes de quitar.')
      return
    }
    router.post(`/payables/${id}/settle`, {
      cash_account_id: String(accountId),
    })
  }
</script>

<div class="max-w-5xl mx-auto p-6">
  <h1 class="text-2xl font-semibold text-stone-800 mb-6">A pagar</h1>

  {#if errors.form}
    <p class="mb-4 text-red-700 text-sm">{errors.form}</p>
  {/if}
  {#if errors.cash_account_id}
    <p class="mb-4 text-red-700 text-sm">{errors.cash_account_id}</p>
  {/if}

  {#if payables.length === 0}
    <div class="border border-dashed border-stone-300 rounded p-8 text-center text-stone-600">
      Nenhuma conta a pagar.
    </div>
  {:else}
    <div class="overflow-x-auto border rounded">
      <table class="w-full text-sm text-left">
        <thead class="bg-stone-100 text-stone-600">
          <tr>
            <th class="p-3 font-medium">Descrição</th>
            <th class="p-3 font-medium">Vencimento</th>
            <th class="p-3 font-medium">Valor</th>
            <th class="p-3 font-medium">Status</th>
            <th class="p-3 font-medium"></th>
          </tr>
        </thead>
        <tbody>
          {#each payables as p}
            <tr class="border-t hover:bg-stone-50">
              <td class="p-3">
                {p.description}
                {#if p.lotId}
                  <span class="text-stone-400 text-xs ml-1">· lote #{p.lotId}</span>
                {/if}
              </td>
              <td class="p-3">{p.dueOn}</td>
              <td class="p-3 font-medium">{p.amount}</td>
              <td class="p-3">{p.statusLabel}</td>
              <td class="p-3 text-right">
                {#if p.canSettle}
                  <button
                    type="button"
                    class="px-3 py-1 bg-stone-800 text-white text-sm rounded"
                    on:click={() => settle(p.id)}
                  >
                    Quitar
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
    <a href="/cash" use:inertia class="underline">Caixa</a>
    <a href="/receivables" use:inertia class="underline">A receber</a>
    <a href="/lots" use:inertia class="underline">Lotes</a>
  </nav>
</div>
