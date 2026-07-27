<script>
  import { router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'

  export let receivables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  function settle(id) {
    if (!confirm('Quitar este recebimento? Será gerada uma entrada no caixa.')) return
    const accountId = cashAccounts[0]?.id
    if (!accountId) {
      alert('Cadastre uma conta de caixa antes de quitar.')
      return
    }
    router.post(`/receivables/${id}/settle`, {
      cash_account_id: String(accountId),
    })
  }
</script>

<AppShell {companyName} active="receivables">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">A receber</h1>
    <p class="text-on-surface-variant text-body-md mt-1">Marketplace e liberações.</p>
    <div class="mt-3"><Nav active="receivables" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  {#if receivables.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">
      Nenhuma conta a receber.
    </div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each receivables as r}
        <div class="ahq-card p-4">
          <div class="flex justify-between gap-2 items-start">
            <div class="min-w-0">
              <p class="font-semibold text-primary">{r.description}</p>
              <p class="text-on-surface-variant text-sm">
                Vence {r.dueOn}
                {#if r.saleId}<span> · venda #{r.saleId}</span>{/if}
              </p>
            </div>
            <span class={r.canSettle ? 'ahq-badge-pending' : 'ahq-badge-live'}>{r.statusLabel}</span>
          </div>
          <div class="mt-3 flex items-end justify-between">
            <p class="ahq-value text-on-tertiary-container">{r.amount}</p>
            {#if r.canSettle}
              <button type="button" class="ahq-btn-primary h-10 px-4 text-sm" on:click={() => settle(r.id)}>
                Quitar
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</AppShell>
