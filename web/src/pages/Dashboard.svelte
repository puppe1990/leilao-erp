<script>
  import { inertia } from '@inertiajs/svelte'
  import Nav from '@/components/Nav.svelte'
  export let balances = []
  export let totalCashFormatted = 'R$ 0,00'
  export let openPayablesFormatted = 'R$ 0,00'
  export let openReceivablesFormatted = 'R$ 0,00'
  export let monthProfitFormatted = 'R$ 0,00'
  export let overduePayables = 0
  export let overdueReceivables = 0
  export let lotCount = 0
  export let ctaLot = false
  export let env = ''
  export let site = {}
</script>

<div class="max-w-5xl mx-auto p-6">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-semibold text-stone-800">Dashboard</h1>
  </div>

  {#if ctaLot}
    <div class="mb-6 border border-dashed border-amber-400 bg-amber-50 rounded p-6 text-center">
      <p class="text-stone-700 mb-3">Nenhum lote cadastrado ainda.</p>
      <a
        href="/lots/new"
        use:inertia
        class="inline-block px-4 py-2 bg-stone-800 text-white text-sm rounded"
      >
        Registrar primeira compra de leilão
      </a>
    </div>
  {/if}

  <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-8">
    <div class="border rounded p-4 bg-stone-50">
      <p class="text-sm text-stone-600">Saldo total</p>
      <p class="text-2xl font-semibold text-stone-800 mt-1">{totalCashFormatted}</p>
      {#if balances.length > 0}
        <ul class="mt-3 space-y-1 text-xs text-stone-500">
          {#each balances as b}
            <li class="flex justify-between gap-2">
              <span>{b.name}</span>
              <span class="font-medium text-stone-700">{b.formatted}</span>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <div class="border rounded p-4 bg-stone-50">
      <p class="text-sm text-stone-600">A receber</p>
      <p class="text-2xl font-semibold text-green-800 mt-1">{openReceivablesFormatted}</p>
      <a href="/receivables" use:inertia class="text-xs underline text-stone-500 mt-2 inline-block"
        >Ver recebíveis</a
      >
    </div>

    <div class="border rounded p-4 bg-stone-50">
      <p class="text-sm text-stone-600">A pagar</p>
      <p class="text-2xl font-semibold text-red-800 mt-1">{openPayablesFormatted}</p>
      <a href="/payables" use:inertia class="text-xs underline text-stone-500 mt-2 inline-block"
        >Ver pagáveis</a
      >
    </div>

    <div class="border rounded p-4 bg-stone-50">
      <p class="text-sm text-stone-600">Lucro do mês</p>
      <p class="text-2xl font-semibold text-stone-800 mt-1">{monthProfitFormatted}</p>
      <p class="text-xs text-stone-500 mt-2">Vendas líquidas − custo unitário</p>
    </div>
  </div>

  <section class="mb-8 border rounded p-4">
    <h2 class="text-sm font-medium text-stone-600 mb-3">Alertas vencidos</h2>
    {#if overduePayables === 0 && overdueReceivables === 0}
      <p class="text-sm text-stone-500">Nenhum título vencido.</p>
    {:else}
      <ul class="text-sm space-y-2">
        {#if overduePayables > 0}
          <li class="text-red-700">
            {overduePayables}
            {overduePayables === 1 ? 'conta a pagar vencida' : 'contas a pagar vencidas'}
            — <a href="/payables" use:inertia class="underline">A pagar</a>
          </li>
        {/if}
        {#if overdueReceivables > 0}
          <li class="text-amber-800">
            {overdueReceivables}
            {overdueReceivables === 1 ? 'recebível vencido' : 'recebíveis vencidos'}
            — <a href="/receivables" use:inertia class="underline">A receber</a>
          </li>
        {/if}
      </ul>
    {/if}
  </section>

  <p class="text-xs text-stone-400 mb-4">
    Lotes: {lotCount}{#if env} · {env}{/if}
  </p>

  <Nav active="dashboard" showLogout={true} />
</div>
