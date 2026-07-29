<script>
  import { useForm, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'
  import SearchableSelect from '@/components/SearchableSelect.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let payables = []
  export let cashAccounts = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let createForm = useForm({
    description: '',
    amount: '',
    due_on: new Date().toISOString().slice(0, 10),
  })

  let editingId = null
  let edit = {
    description: '',
    amount: '',
    due_on: '',
  }

  $: cashAccountOptions = (Array.isArray(cashAccounts) ? cashAccounts : []).map((a) => ({
    value: String(a.id),
    label: a.name,
  }))

  async function settle(id) {
    const ok = await askConfirm({
      title: 'Quitar pagamento',
      message: 'Confirmar quitação deste título a pagar?',
      detail: 'Será gerada uma saída no caixa.',
      confirmLabel: 'Quitar',
      tone: 'primary',
      icon: 'payments',
    })
    if (!ok) return
    const accountId = cashAccounts[0]?.id
    if (!accountId) {
      alert('Cadastre uma conta de caixa antes de quitar.')
      return
    }
    router.post(`/payables/${id}/settle`, { cash_account_id: String(accountId) })
  }

  async function cancel(id) {
    const ok = await askConfirm({
      title: 'Cancelar título',
      message: 'Tem certeza que deseja cancelar este título a pagar?',
      confirmLabel: 'Cancelar título',
      tone: 'warning',
      icon: 'cancel',
    })
    if (!ok) return
    router.post(`/payables/${id}/cancel`)
  }

  async function destroy(id) {
    const ok = await askConfirm({
      title: 'Excluir título',
      message: 'Excluir este título permanentemente?',
      detail: 'Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/payables/${id}/delete`)
  }

  function submitCreate() {
    createForm.post('/payables', {
      onSuccess: () => createForm.reset('description', 'amount'),
    })
  }

  function startEdit(p) {
    editingId = p.id
    edit = {
      description: p.description || '',
      amount: p.amountRaw || '',
      due_on: p.dueOn || '',
    }
  }

  function cancelEdit() {
    editingId = null
  }

  function saveEdit(id) {
    router.post(`/payables/${id}`, { ...edit }, {
      onSuccess: () => {
        editingId = null
      },
    })
  }
</script>

<AppShell {companyName} active="payables">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">A pagar</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Títulos a pagar — criar, editar, quitar, cancelar ou excluir.
    </p>
    <div class="mt-3"><Nav active="payables" /></div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="ahq-card p-4 mb-section-padding">
    <h2 class="font-semibold text-primary mb-3">Novo título</h2>
    <form on:submit|preventDefault={submitCreate} class="grid gap-3 sm:grid-cols-3">
      <input class="ahq-input h-10 sm:col-span-2" placeholder="Descrição" bind:value={createForm.description} />
      <input class="ahq-input h-10 font-mono" placeholder="Valor R$" bind:value={createForm.amount} />
      <input type="date" class="ahq-input h-10 font-mono" bind:value={createForm.due_on} />
      <button type="submit" class="ahq-btn-primary h-10 sm:col-span-2" disabled={createForm.processing}>Criar</button>
    </form>
    {#if errors.amount}<p class="text-error text-xs mt-2">{errors.amount}</p>{/if}
  </section>

  {#if payables.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">Nenhuma conta a pagar.</div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each payables as p (p.id)}
        <div class="ahq-card p-4">
          {#if editingId === p.id}
            <div class="grid gap-2 sm:grid-cols-3">
              <input class="ahq-input h-9 sm:col-span-2 text-sm" bind:value={edit.description} placeholder="Descrição" />
              <input class="ahq-input h-9 font-mono text-sm" bind:value={edit.amount} placeholder="Valor" />
              <input type="date" class="ahq-input h-9 font-mono text-sm" bind:value={edit.due_on} />
              <div class="sm:col-span-2 flex gap-3">
                <button type="button" class="text-secondary text-sm font-medium" on:click={() => saveEdit(p.id)}>
                  Salvar
                </button>
                <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>Cancelar</button>
              </div>
            </div>
          {:else}
            <div class="flex justify-between gap-2 items-start">
              <div class="min-w-0">
                <p class="font-semibold text-primary">{p.description}</p>
                <p class="text-on-surface-variant text-sm">
                  Vence {p.dueOn}
                  {#if p.lotId}<span> · lote #{p.lotId}</span>{/if}
                </p>
              </div>
              <span class={p.canSettle ? 'ahq-badge-pending' : 'ahq-badge-sold'}>{p.statusLabel}</span>
            </div>
            <div class="mt-3 flex items-end justify-between gap-2 flex-wrap">
              <p class="ahq-value">{p.amount}</p>
              <div class="flex gap-2 flex-wrap justify-end">
                {#if p.canEdit}
                  <button type="button" class="ahq-btn-ghost h-10 px-3 text-sm" on:click={() => startEdit(p)}>
                    Editar
                  </button>
                {/if}
                {#if p.canSettle}
                  <button type="button" class="ahq-btn-primary h-10 px-4 text-sm" on:click={() => settle(p.id)}>
                    Quitar
                  </button>
                {/if}
                {#if p.canCancel}
                  <button
                    type="button"
                    class="ahq-btn-ghost h-10 px-3 text-sm text-error border-error"
                    on:click={() => cancel(p.id)}
                  >
                    Cancelar
                  </button>
                {/if}
                {#if p.canDelete}
                  <button type="button" class="ahq-btn-ghost h-10 px-3 text-sm text-error" on:click={() => destroy(p.id)}>
                    Excluir
                  </button>
                {/if}
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  {#if cashAccountOptions.length === 0}
    <p class="mt-4 text-sm text-on-surface-variant">Cadastre uma conta em Caixa para poder quitar títulos.</p>
  {/if}
</AppShell>
