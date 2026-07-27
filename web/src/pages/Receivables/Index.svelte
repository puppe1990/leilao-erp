<script>
  import { useForm, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'

  export let receivables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let createForm = useForm({
    description: '',
    amount: '',
    due_on: new Date().toISOString().slice(0, 10),
  })

  function settle(id) {
    if (!confirm('Quitar este recebimento? Será gerada uma entrada no caixa.')) return
    const accountId = cashAccounts[0]?.id
    if (!accountId) {
      alert('Cadastre uma conta de caixa antes de quitar.')
      return
    }
    router.post(`/receivables/${id}/settle`, { cash_account_id: String(accountId) })
  }

  function cancel(id) {
    if (!confirm('Cancelar este recebível? Se ligado a venda pendente, a venda também será cancelada.')) return
    router.post(`/receivables/${id}/cancel`)
  }

  function submitCreate() {
    createForm.post('/receivables', {
      onSuccess: () => createForm.reset('description', 'amount'),
    })
  }
</script>

<AppShell {companyName} active="receivables">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">A receber</h1>
    <p class="text-on-surface-variant text-body-md mt-1">CRUD de recebíveis e quitações.</p>
    <div class="mt-3"><Nav active="receivables" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="ahq-card p-4 mb-section-padding">
    <h2 class="font-semibold text-primary mb-3">Novo recebível</h2>
    <form on:submit|preventDefault={submitCreate} class="grid gap-3 sm:grid-cols-3">
      <input class="ahq-input h-10 sm:col-span-2" placeholder="Descrição" bind:value={createForm.description} />
      <input class="ahq-input h-10 font-mono" placeholder="Valor R$" bind:value={createForm.amount} />
      <input type="date" class="ahq-input h-10 font-mono" bind:value={createForm.due_on} />
      <button type="submit" class="ahq-btn-primary h-10 sm:col-span-2" disabled={createForm.processing}>Criar</button>
    </form>
  </section>

  {#if receivables.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">Nenhuma conta a receber.</div>
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
          <div class="mt-3 flex items-end justify-between gap-2 flex-wrap">
            <p class="ahq-value text-on-tertiary-container">{r.amount}</p>
            <div class="flex gap-2">
              {#if r.canSettle}
                <button type="button" class="ahq-btn-primary h-10 px-4 text-sm" on:click={() => settle(r.id)}>Quitar</button>
              {/if}
              {#if r.canCancel}
                <button type="button" class="ahq-btn-ghost h-10 px-4 text-sm text-error border-error" on:click={() => cancel(r.id)}>
                  Cancelar
                </button>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</AppShell>
