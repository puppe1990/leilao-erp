<script>
  import { useForm, router } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import Nav from '@/components/Nav.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

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

  let editingId = null
  let edit = {
    description: '',
    amount: '',
    due_on: '',
  }

  async function settle(id) {
    const ok = await askConfirm({
      title: 'Quitar recebimento',
      message: 'Confirmar quitação deste recebível?',
      detail: 'Será gerada uma entrada no caixa.',
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
    router.post(`/receivables/${id}/settle`, { cash_account_id: String(accountId) })
  }

  async function cancel(id) {
    const ok = await askConfirm({
      title: 'Cancelar recebível',
      message: 'Tem certeza que deseja cancelar este recebível?',
      detail: 'Se estiver ligado a uma venda pendente, a venda também será cancelada.',
      confirmLabel: 'Cancelar recebível',
      tone: 'warning',
      icon: 'cancel',
    })
    if (!ok) return
    router.post(`/receivables/${id}/cancel`)
  }

  async function destroy(id) {
    const ok = await askConfirm({
      title: 'Excluir recebível',
      message: 'Excluir este recebível permanentemente?',
      detail: 'Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/receivables/${id}/delete`)
  }

  function submitCreate() {
    createForm.post('/receivables', {
      onSuccess: () => createForm.reset('description', 'amount'),
    })
  }

  function startEdit(r) {
    editingId = r.id
    edit = {
      description: r.description || '',
      amount: r.amountRaw || '',
      due_on: r.dueOn || '',
    }
  }

  function cancelEdit() {
    editingId = null
  }

  function saveEdit(id) {
    router.post(`/receivables/${id}`, { ...edit }, {
      onSuccess: () => {
        editingId = null
      },
    })
  }
</script>

<AppShell {companyName} active="receivables">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">A receber</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Recebíveis — criar, editar, quitar, cancelar ou excluir.
    </p>
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
    {#if errors.amount}<p class="text-error text-xs mt-2">{errors.amount}</p>{/if}
  </section>

  {#if receivables.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">Nenhuma conta a receber.</div>
  {:else}
    <div class="flex flex-col gap-stack-gap">
      {#each receivables as r (r.id)}
        <div class="ahq-card p-4">
          {#if editingId === r.id}
            <div class="grid gap-2 sm:grid-cols-3">
              <input class="ahq-input h-9 sm:col-span-2 text-sm" bind:value={edit.description} placeholder="Descrição" />
              <input class="ahq-input h-9 font-mono text-sm" bind:value={edit.amount} placeholder="Valor" />
              <input type="date" class="ahq-input h-9 font-mono text-sm" bind:value={edit.due_on} />
              <div class="sm:col-span-2 flex gap-3">
                <button type="button" class="text-secondary text-sm font-medium" on:click={() => saveEdit(r.id)}>
                  Salvar
                </button>
                <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>Cancelar</button>
              </div>
            </div>
          {:else}
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
              <div class="flex gap-2 flex-wrap justify-end">
                {#if r.canEdit}
                  <button type="button" class="ahq-btn-ghost h-10 px-3 text-sm" on:click={() => startEdit(r)}>
                    Editar
                  </button>
                {/if}
                {#if r.canSettle}
                  <button type="button" class="ahq-btn-primary h-10 px-4 text-sm" on:click={() => settle(r.id)}>
                    Quitar
                  </button>
                {/if}
                {#if r.canCancel}
                  <button
                    type="button"
                    class="ahq-btn-ghost h-10 px-3 text-sm text-error border-error"
                    on:click={() => cancel(r.id)}
                  >
                    Cancelar
                  </button>
                {/if}
                {#if r.canDelete}
                  <button type="button" class="ahq-btn-ghost h-10 px-3 text-sm text-error" on:click={() => destroy(r.id)}>
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
</AppShell>
