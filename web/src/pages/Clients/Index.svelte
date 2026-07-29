<script>
  import { router, useForm } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let clients = []
  export let query = ''
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  let search = query || ''

  let createForm = useForm({
    name: '',
    phone: '',
    email: '',
    document: '',
    notes: '',
  })

  let editingId = null
  let edit = {
    name: '',
    phone: '',
    email: '',
    document: '',
    notes: '',
  }

  function submitCreate() {
    createForm.post('/clients', {
      onSuccess: () => createForm.reset(),
    })
  }

  function startEdit(c) {
    editingId = c.id
    edit = {
      name: c.name || '',
      phone: c.phone || '',
      email: c.email || '',
      document: c.document || '',
      notes: c.notes || '',
    }
  }

  function cancelEdit() {
    editingId = null
  }

  function saveEdit(c) {
    router.post(`/clients/${c.id}`, { ...edit }, {
      onSuccess: () => {
        editingId = null
      },
    })
  }

  async function destroy(c) {
    const ok = await askConfirm({
      title: 'Excluir cliente',
      message: `Tem certeza que deseja excluir “${c.name}”?`,
      detail: 'Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir',
      tone: 'danger',
      icon: 'person_remove',
    })
    if (!ok) return
    router.post(`/clients/${c.id}/delete`)
  }

  function runSearch() {
    const q = search.trim()
    router.get(q ? `/clients?q=${encodeURIComponent(q)}` : '/clients', {}, { preserveState: true })
  }
</script>

<AppShell {companyName} active="clients">
  <div class="flex items-start justify-between gap-3 mb-section-padding flex-wrap">
    <div>
      <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Clientes</h1>
      <p class="text-on-surface-variant text-body-md mt-1">
        Cadastro de compradores — nome, telefone, e-mail e documento.
      </p>
    </div>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="ahq-card p-4 mb-section-padding">
    <h2 class="font-semibold text-primary mb-3">Novo cliente</h2>
    <form on:submit|preventDefault={submitCreate} class="grid gap-3 sm:grid-cols-2">
      <div class="sm:col-span-2">
        <label class="ahq-label block mb-1" for="c_name">Nome *</label>
        <input id="c_name" class="ahq-input h-10" bind:value={createForm.name} required placeholder="Nome completo" />
      </div>
      <div>
        <label class="ahq-label block mb-1" for="c_phone">Telefone</label>
        <input id="c_phone" class="ahq-input h-10 font-mono" bind:value={createForm.phone} placeholder="(11) 99999-0000" />
      </div>
      <div>
        <label class="ahq-label block mb-1" for="c_email">E-mail</label>
        <input id="c_email" type="email" class="ahq-input h-10" bind:value={createForm.email} placeholder="email@exemplo.com" />
      </div>
      <div>
        <label class="ahq-label block mb-1" for="c_doc">CPF/CNPJ</label>
        <input id="c_doc" class="ahq-input h-10 font-mono" bind:value={createForm.document} placeholder="Opcional" />
      </div>
      <div>
        <label class="ahq-label block mb-1" for="c_notes">Notas</label>
        <input id="c_notes" class="ahq-input h-10" bind:value={createForm.notes} placeholder="Preferências, etc." />
      </div>
      <div class="sm:col-span-2">
        <button type="submit" class="ahq-btn-primary h-10 px-5" disabled={createForm.processing}>
          Salvar cliente
        </button>
      </div>
    </form>
  </section>

  <div class="ahq-card overflow-hidden">
    <div class="p-3 border-b border-outline-variant flex flex-col sm:flex-row gap-2 sm:items-center sm:justify-between">
      <form
        class="relative flex-1 max-w-md flex gap-2"
        on:submit|preventDefault={runSearch}
      >
        <div class="relative flex-1">
          <span
            class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-[20px]"
          >
            search
          </span>
          <input
            type="search"
            class="ahq-input h-10 pl-10 w-full"
            placeholder="Buscar nome, telefone, e-mail…"
            bind:value={search}
            aria-label="Buscar clientes"
          />
        </div>
        <button type="submit" class="ahq-btn-ghost h-10 px-3 text-sm">Buscar</button>
      </form>
      <p class="text-sm text-on-surface-variant">
        {(clients || []).length}
        {(clients || []).length === 1 ? 'cliente' : 'clientes'}
      </p>
    </div>

    {#if (clients || []).length === 0}
      <div class="p-10 text-center text-on-surface-variant border-dashed">
        Nenhum cliente cadastrado ainda.
      </div>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left min-w-[640px]">
          <thead>
            <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase">Nome</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase w-36">Telefone</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase">E-mail</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase w-32">Documento</th>
              <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-36">Ações</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline-variant">
            {#each clients as c (c.id)}
              {#if editingId === c.id}
                <tr class="bg-secondary-container/20">
                  <td class="px-3 py-2" colspan="4">
                    <div class="grid sm:grid-cols-2 gap-2">
                      <input class="ahq-input h-9 text-sm" bind:value={edit.name} placeholder="Nome" />
                      <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.phone} placeholder="Telefone" />
                      <input class="ahq-input h-9 text-sm" bind:value={edit.email} placeholder="E-mail" />
                      <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.document} placeholder="CPF/CNPJ" />
                      <input class="ahq-input h-9 text-sm sm:col-span-2" bind:value={edit.notes} placeholder="Notas" />
                    </div>
                  </td>
                  <td class="px-3 py-2 text-right whitespace-nowrap align-top">
                    <button type="button" class="text-secondary font-medium text-sm mr-2" on:click={() => saveEdit(c)}>
                      Salvar
                    </button>
                    <button type="button" class="text-on-surface-variant text-sm" on:click={cancelEdit}>
                      Cancelar
                    </button>
                  </td>
                </tr>
              {:else}
                <tr class="hover:bg-surface-container-low/80">
                  <td class="px-3 py-2.5">
                    <p class="font-medium text-primary">{c.name}</p>
                    {#if c.notes}
                      <p class="text-[11px] text-on-surface-variant mt-0.5">{c.notes}</p>
                    {/if}
                  </td>
                  <td class="px-3 py-2.5 font-mono text-on-surface-variant">{c.phone || '—'}</td>
                  <td class="px-3 py-2.5 text-on-surface-variant break-all">{c.email || '—'}</td>
                  <td class="px-3 py-2.5 font-mono text-on-surface-variant">{c.document || '—'}</td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <button
                      type="button"
                      class="text-on-surface-variant hover:text-secondary text-sm font-medium mr-2"
                      on:click={() => startEdit(c)}
                    >
                      Editar
                    </button>
                    <button
                      type="button"
                      class="text-error text-sm font-medium"
                      on:click={() => destroy(c)}
                    >
                      Excluir
                    </button>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</AppShell>
