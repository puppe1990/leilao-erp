<script>
  import { inertia, router, useForm } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let suppliers = []
  export let errors = {}
  export let site = {}
  export let companyName = 'Pix Eletro'

  const emptySupplier = {
    name: '',
    whatsapp: '',
    notes: '',
    min_order: '',
  }

  let createForm = useForm({ ...emptySupplier })
  let editingId = null
  let edit = { ...emptySupplier }

  $: rows = Array.isArray(suppliers) ? suppliers : []

  function submitCreate() {
    createForm.post('/suppliers', {
      preserveState: (page) => Object.keys(page.props?.errors || {}).length > 0,
    })
  }

  function startEdit(row) {
    editingId = row.id
    edit = {
      name: row.name || '',
      whatsapp: row.whatsappLabel || row.whatsapp || '',
      notes: row.notes || '',
      min_order: row.minOrderRaw || '',
    }
  }

  function cancelEdit() {
    editingId = null
  }

  function saveEdit(id) {
    router.post(`/suppliers/${id}`, { ...edit }, {
      onSuccess: () => {
        editingId = null
      },
    })
  }

  async function destroy(row) {
    const ok = await askConfirm({
      title: 'Descartar fornecedor',
      message: `Descartar “${row.name}”?`,
      detail: 'A conversa salva também some. Essa ação não pode ser desfeita.',
      confirmLabel: 'Descartar',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/suppliers/${row.id}/delete`)
  }
</script>

<AppShell {companyName} active="suppliers">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Fornecedores</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      Atacado da 25 e da Santa Efigênia — nome, WhatsApp e a conversa salva.
    </p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="ahq-card p-4 mb-section-padding">
    <h2 class="font-semibold text-primary mb-3">Novo fornecedor</h2>
    <form on:submit|preventDefault={submitCreate} class="grid gap-3">
      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label class="ahq-label block mb-1" for="s_name">Nome</label>
          <input id="s_name" class="ahq-input h-10" bind:value={createForm.name} required placeholder="PMCELL" />
          {#if errors.name}<p class="text-error text-xs mt-1">{errors.name}</p>{/if}
        </div>
        <div>
          <label class="ahq-label block mb-1" for="s_whatsapp">WhatsApp</label>
          <input
            id="s_whatsapp"
            class="ahq-input h-10 font-mono text-sm"
            bind:value={createForm.whatsapp}
            required
            placeholder="(11) 91130-4693"
          />
          {#if errors.whatsapp}<p class="text-error text-xs mt-1">{errors.whatsapp}</p>{/if}
        </div>
      </div>
      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label class="ahq-label block mb-1" for="s_notes">O que vende</label>
          <input
            id="s_notes"
            class="ahq-input h-10"
            bind:value={createForm.notes}
            placeholder="power bank, carregador 25W, fone TWS"
          />
        </div>
        <div>
          <label class="ahq-label block mb-1" for="s_min">Pedido mínimo</label>
          <input
            id="s_min"
            class="ahq-input h-10 font-mono"
            bind:value={createForm.min_order}
            placeholder="300"
          />
          {#if errors.min_order}<p class="text-error text-xs mt-1">{errors.min_order}</p>{/if}
        </div>
      </div>
      <div>
        <button type="submit" class="ahq-btn-primary h-10 px-5" disabled={createForm.processing}>
          Salvar fornecedor
        </button>
      </div>
    </form>
  </section>

  <p class="text-sm text-on-surface-variant mb-4">
    {rows.length}
    {rows.length === 1 ? 'fornecedor' : 'fornecedores'}
  </p>

  {#if rows.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">
      Nenhum fornecedor ainda. Coloca o nome e o WhatsApp do atacado.
    </div>
  {:else}
    <div class="ahq-card overflow-x-auto">
      <table class="w-full text-sm text-left min-w-[720px]">
        <thead>
          <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase">Nome</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase w-40">WhatsApp</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase">Última mensagem</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-28">Mínimo</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-44">Ações</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant">
          {#each rows as row (row.id)}
            {#if editingId === row.id}
              <tr class="bg-secondary-container/20">
                <td class="px-3 py-2" colspan="4">
                  <div class="grid sm:grid-cols-4 gap-2">
                    <input class="ahq-input h-9 text-sm" bind:value={edit.name} placeholder="Nome" />
                    <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.whatsapp} placeholder="WhatsApp" />
                    <input class="ahq-input h-9 text-sm" bind:value={edit.notes} placeholder="O que vende" />
                    <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.min_order} placeholder="Mínimo" />
                  </div>
                </td>
                <td class="px-3 py-2 text-right whitespace-nowrap align-top">
                  <button type="button" class="text-secondary font-medium text-sm mr-2" on:click={() => saveEdit(row.id)}>
                    Salvar
                  </button>
                  <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>Cancelar</button>
                </td>
              </tr>
            {:else}
              <tr class="hover:bg-surface-container-low/80">
                <td class="px-3 py-2.5 min-w-0">
                  <a href={`/suppliers/${row.id}`} use:inertia class="font-semibold text-primary hover:underline">
                    {row.name}
                  </a>
                  {#if row.notes}
                    <p class="text-xs text-on-surface-variant truncate mt-0.5">{row.notes}</p>
                  {/if}
                </td>
                <td class="px-3 py-2.5 font-mono whitespace-nowrap text-xs">{row.whatsappLabel}</td>
                <td class="px-3 py-2.5 min-w-0">
                  {#if row.lastBody}
                    <p class="truncate">{row.lastFromMe ? 'Você: ' : ''}{row.lastBody}</p>
                    {#if row.lastAt}
                      <p class="text-[11px] text-on-surface-variant mt-0.5">{row.lastAt}</p>
                    {/if}
                  {:else}
                    <span class="text-on-surface-variant">Sem conversa ainda</span>
                  {/if}
                </td>
                <td class="px-3 py-2.5 text-right font-mono whitespace-nowrap">{row.minOrder}</td>
                <td class="px-3 py-2.5 text-right whitespace-nowrap">
                  <button type="button" class="text-secondary font-medium text-sm mr-2" on:click={() => startEdit(row)}>
                    Editar
                  </button>
                  <button
                    type="button"
                    class="ahq-btn-ghost h-8 px-3 text-error border-error hover:border-error hover:text-error"
                    on:click={() => destroy(row)}
                  >
                    Descartar
                  </button>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</AppShell>
