<script>
  import { inertia } from '@inertiajs/svelte'
  export let lots = []
  export let site = {}
</script>

<div class="max-w-4xl mx-auto p-6">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-semibold text-stone-800">Lotes</h1>
    <a href="/lots/new" use:inertia class="px-4 py-2 bg-stone-800 text-white text-sm rounded">
      Novo lote
    </a>
  </div>

  {#if lots.length === 0}
    <div class="border border-dashed border-stone-300 rounded p-8 text-center text-stone-600">
      <p class="mb-4">Nenhum lote cadastrado.</p>
      <a href="/lots/new" use:inertia class="underline text-amber-900">
        Registrar primeira compra de leilão
      </a>
    </div>
  {:else}
    <div class="overflow-x-auto border rounded">
      <table class="w-full text-sm text-left">
        <thead class="bg-stone-100 text-stone-600">
          <tr>
            <th class="p-3 font-medium">Nome</th>
            <th class="p-3 font-medium">Data</th>
            <th class="p-3 font-medium">Custo total</th>
            <th class="p-3 font-medium">Itens</th>
            <th class="p-3 font-medium">Status</th>
          </tr>
        </thead>
        <tbody>
          {#each lots as lot}
            <tr class="border-t hover:bg-stone-50">
              <td class="p-3">
                <a href={`/lots/${lot.id}`} use:inertia class="underline text-stone-800">
                  {lot.name}
                </a>
              </td>
              <td class="p-3">{lot.purchasedAt}</td>
              <td class="p-3">{lot.totalCost}</td>
              <td class="p-3">{lot.itemCount}</td>
              <td class="p-3">{lot.statusLabel}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <nav class="mt-6 text-sm space-x-4">
    <a href="/dashboard" use:inertia class="underline">Dashboard</a>
    <a href="/sales" use:inertia class="underline">Vendas</a>
    <a href="/cash" use:inertia class="underline">Caixa</a>
    <a href="/payables" use:inertia class="underline">A pagar</a>
    <a href="/receivables" use:inertia class="underline">A receber</a>
  </nav>
</div>
