<script>
  import { inertia } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'

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

<AppShell active="dashboard">
  <section class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Dashboard</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Visão financeira do seu negócio de leilão.</p>
  </section>

  {#if ctaLot}
    <div class="ahq-card p-6 mb-section-padding text-center border-dashed">
      <span class="material-symbols-outlined text-4xl text-secondary mb-2">gavel</span>
      <p class="text-on-surface mb-4">Nenhum lote cadastrado ainda.</p>
      <a href="/lots/new" use:inertia class="ahq-btn-primary">Registrar primeira compra de leilão</a>
    </div>
  {/if}

  <!-- Metric cards -->
  <section class="grid grid-cols-2 md:grid-cols-3 gap-stack-gap mb-section-padding">
    <div class="ahq-card p-4 flex flex-col gap-1">
      <span class="ahq-label">Lotes</span>
      <span class="ahq-value text-primary">{lotCount}</span>
      <div class="mt-2 flex items-center gap-1 text-[10px] text-on-surface-variant">
        <span class="material-symbols-outlined text-[14px]">inventory_2</span>
        <span>cadastrados</span>
      </div>
    </div>

    <div class="ahq-card p-4 flex flex-col gap-1">
      <span class="ahq-label">Saldo total</span>
      <span class="ahq-value">{totalCashFormatted}</span>
      {#if balances.length > 0}
        <div class="mt-2 space-y-0.5">
          {#each balances as b}
            <div class="flex justify-between text-[10px] text-on-surface-variant">
              <span>{b.name}</span>
              <span class="font-mono">{b.formatted}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <div class="col-span-2 md:col-span-1 ahq-card p-4 flex flex-col gap-1">
      <span class="ahq-label">Lucro do mês</span>
      <span class="ahq-value">{monthProfitFormatted}</span>
      <div class="mt-2 flex items-center gap-1 text-[10px] text-secondary">
        <span class="material-symbols-outlined text-[14px]">payments</span>
        <span>líquido − custo unitário</span>
      </div>
    </div>
  </section>

  <section class="grid grid-cols-2 gap-stack-gap mb-section-padding">
    <a href="/receivables" use:inertia class="ahq-card p-4 block hover:border-secondary transition-colors">
      <span class="ahq-label">A receber</span>
      <p class="ahq-value text-on-tertiary-container mt-1">{openReceivablesFormatted}</p>
      <span class="text-[10px] text-secondary mt-2 inline-block">Ver recebíveis →</span>
    </a>
    <a href="/payables" use:inertia class="ahq-card p-4 block hover:border-secondary transition-colors">
      <span class="ahq-label">A pagar</span>
      <p class="ahq-value text-error mt-1">{openPayablesFormatted}</p>
      <span class="text-[10px] text-secondary mt-2 inline-block">Ver pagáveis →</span>
    </a>
  </section>

  <!-- Alerts as activity-style list -->
  <section class="mb-section-padding">
    <h2 class="font-headline-md text-headline-md text-primary mb-stack-gap">Alertas</h2>
    <div class="ahq-card divide-y divide-outline-variant">
      {#if overduePayables === 0 && overdueReceivables === 0}
        <div class="p-4 flex gap-4 items-center">
          <div class="w-10 h-10 rounded-full bg-tertiary-fixed/20 flex items-center justify-center shrink-0">
            <span class="material-symbols-outlined text-on-tertiary-container">check_circle</span>
          </div>
          <div>
            <p class="font-semibold text-body-md">Tudo em dia</p>
            <p class="text-on-surface-variant text-body-md">Nenhum título vencido.</p>
          </div>
        </div>
      {:else}
        {#if overduePayables > 0}
          <a href="/payables" use:inertia class="p-4 flex gap-4 hover:bg-surface-container-low">
            <div class="w-10 h-10 rounded-full bg-error-container flex items-center justify-center shrink-0">
              <span class="material-symbols-outlined text-error">warning</span>
            </div>
            <div class="flex-1">
              <p class="font-semibold text-body-md">Contas a pagar vencidas</p>
              <p class="text-on-surface-variant text-body-md">
                {overduePayables}
                {overduePayables === 1 ? 'título precisa de atenção' : 'títulos precisam de atenção'}
              </p>
            </div>
          </a>
        {/if}
        {#if overdueReceivables > 0}
          <a href="/receivables" use:inertia class="p-4 flex gap-4 hover:bg-surface-container-low">
            <div class="w-10 h-10 rounded-full bg-pending/15 flex items-center justify-center shrink-0">
              <span class="material-symbols-outlined text-pending">schedule</span>
            </div>
            <div class="flex-1">
              <p class="font-semibold text-body-md">Recebíveis vencidos</p>
              <p class="text-on-surface-variant text-body-md">
                {overdueReceivables}
                {overdueReceivables === 1 ? 'valor aguardando liberação' : 'valores aguardando liberação'}
              </p>
            </div>
          </a>
        {/if}
      {/if}
    </div>
  </section>

  <section class="grid grid-cols-2 gap-stack-gap">
    <a href="/lots/new" use:inertia class="ahq-btn-dark w-full text-sm">+ Novo lote</a>
    <a href="/sales/new" use:inertia class="ahq-btn-primary w-full text-sm">+ Nova venda</a>
  </section>

  {#if env}
    <p class="text-[10px] text-on-surface-variant mt-6 uppercase tracking-wider font-mono">{env}</p>
  {/if}
</AppShell>
