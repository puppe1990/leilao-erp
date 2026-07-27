<script>
  import { router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'

  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

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

<AppShell {companyName} active="payables">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">A pagar</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Títulos e quitações.</p>
    <div class="mt-3"><Nav active="payables" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if payables.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">
      Nenhuma conta a pagar.
    </div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each payables as p}
        <div class="ahq-card p-4">
          <div class="flex justify-between gap-2 items-start">
            <div class="min-w-0">
              <p class="font-semibold text-primary">{p.description}</p>
              <p class="text-on-surface-variant text-sm">
                Vence {p.dueOn}
                {#if p.lotId}<span> · lote #{p.lotId}</span>{/if}
              </p>
            </div>
            <span
              class={p.canSettle ? 'ahq-badge-pending' : 'ahq-badge-sold'}
            >{p.statusLabel}</span>
          </div>
          <div class="mt-3 flex items-end justify-between">
            <p class="ahq-value">{p.amount}</p>
            {#if p.canSettle}
              <button type="button" class="ahq-btn-primary h-10 px-4 text-sm" on:click={() => settle(p.id)}>
                Quitar
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</AppShell>
