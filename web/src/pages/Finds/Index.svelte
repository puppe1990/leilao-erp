<script>
  import { router, useForm } from '@inertiajs/svelte'
  import AppShell from '@/components/AppShell.svelte'
  import { askConfirm } from '@/lib/confirmDialog.js'

  export let finds = []
  export let source = ''
  export let sources = []
  export let errors = {}
  export let site = {}
  export let companyName = 'AuctionHQ'

  const emptyFind = {
    name: '',
    url: '',
    source: 'olx',
    buy: '',
    sell: '',
  }

  let createForm = useForm({ ...emptyFind })
  let editingId = null
  let edit = { ...emptyFind }

  $: rows = Array.isArray(finds) ? finds : []
  $: sourceOptions = Array.isArray(sources) ? sources : []

  function filterSource(value) {
    router.get(value ? `/finds?source=${encodeURIComponent(value)}` : '/finds', {}, { preserveState: true })
  }

  function submitCreate() {
    createForm.post('/finds', {
      preserveState: (page) => Object.keys(page.props?.errors || {}).length > 0,
    })
  }

  function startEdit(row) {
    editingId = row.id
    edit = {
      name: row.name || '',
      url: row.url || '',
      source: row.source || 'olx',
      buy: row.buyRaw || '',
      sell: row.sellRaw || '',
    }
  }

  function cancelEdit() {
    editingId = null
  }

  function saveEdit(id) {
    router.post(`/finds/${id}`, { ...edit }, {
      onSuccess: () => {
        editingId = null
      },
    })
  }

  async function destroy(row) {
    const ok = await askConfirm({
      title: 'Excluir achado',
      message: `Excluir “${row.name}”?`,
      detail: 'Essa ação não pode ser desfeita.',
      confirmLabel: 'Excluir',
      tone: 'danger',
    })
    if (!ok) return
    router.post(`/finds/${row.id}/delete`)
  }

  function marginClass(margin) {
    const text = String(margin || '')
    if (text.startsWith('-')) return 'text-error'
    if (text === 'R$ 0,00') return 'text-on-surface-variant'
    return 'text-secondary'
  }
</script>

<AppShell {companyName} active="finds">
  <div class="mb-section-padding">
    <h1 class="font-headline-lg text-headline-lg-mobile text-primary">Achados</h1>
    <p class="text-on-surface-variant text-body-md mt-1">
      OLX, Facebook e Kwara — nome, link, quanto pagar e quanto dá para vender.
    </p>
  </div>

  {#if errors.form}
    <p class="mb-4 text-error text-sm ahq-card p-3 bg-error-container/30">{errors.form}</p>
  {/if}

  <section class="ahq-card p-4 mb-section-padding">
    <h2 class="font-semibold text-primary mb-3">Novo achado</h2>
    <form on:submit|preventDefault={submitCreate} class="grid gap-3">
      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label class="ahq-label block mb-1" for="f_name">Nome</label>
          <input id="f_name" class="ahq-input h-10" bind:value={createForm.name} required placeholder="Dell P2016t" />
        </div>
        <div>
          <label class="ahq-label block mb-1" for="f_url">Link</label>
          <input
            id="f_url"
            class="ahq-input h-10 font-mono text-sm"
            bind:value={createForm.url}
            required
            type="url"
            placeholder="https://www.olx.com.br/vi/…"
          />
        </div>
      </div>
      <div class="grid gap-3 sm:grid-cols-3">
        <div>
          <label class="ahq-label block mb-1" for="f_source">Fonte</label>
          <select id="f_source" class="ahq-input h-10" bind:value={createForm.source}>
            {#each sourceOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>
        <div>
          <label class="ahq-label block mb-1" for="f_buy">Comprar</label>
          <input id="f_buy" class="ahq-input h-10 font-mono" bind:value={createForm.buy} required placeholder="175" />
          {#if errors.buy}<p class="text-error text-xs mt-1">{errors.buy}</p>{/if}
        </div>
        <div>
          <label class="ahq-label block mb-1" for="f_sell">Vender</label>
          <input id="f_sell" class="ahq-input h-10 font-mono" bind:value={createForm.sell} required placeholder="250" />
          {#if errors.sell}<p class="text-error text-xs mt-1">{errors.sell}</p>{/if}
        </div>
      </div>
      <div>
        <button type="submit" class="ahq-btn-primary h-10 px-5" disabled={createForm.processing}>Salvar achado</button>
      </div>
    </form>
  </section>

  <div class="flex flex-wrap items-center gap-2 mb-4">
    <button
      type="button"
      class="px-3 py-1.5 rounded-full border text-sm {source === ''
        ? 'bg-secondary text-on-secondary border-transparent font-semibold'
        : 'bg-surface-container-lowest text-on-surface-variant border-outline-variant'}"
      on:click={() => filterSource('')}
    >
      Todos
    </button>
    {#each sourceOptions as opt}
      <button
        type="button"
        class="px-3 py-1.5 rounded-full border text-sm {source === opt.value
          ? 'bg-secondary text-on-secondary border-transparent font-semibold'
          : 'bg-surface-container-lowest text-on-surface-variant border-outline-variant'}"
        on:click={() => filterSource(opt.value)}
      >
        {opt.label}
      </button>
    {/each}
    <p class="text-sm text-on-surface-variant ml-auto">
      {rows.length}
      {rows.length === 1 ? 'achado' : 'achados'}
    </p>
  </div>

  {#if rows.length === 0}
    <div class="ahq-card p-10 text-center text-on-surface-variant border-dashed">
      Nenhum achado ainda. Cole o link da OLX, do Facebook ou da Kwara.
    </div>
  {:else}
    <div class="ahq-card overflow-x-auto">
      <table class="w-full text-sm text-left min-w-[720px]">
        <thead>
          <tr class="bg-surface-container-low text-on-surface-variant border-b border-outline-variant">
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase w-28">Fonte</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase">Nome</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-28">Comprar</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-28">Vender</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-28">Margem</th>
            <th class="px-3 py-2.5 font-medium text-[11px] uppercase text-right w-36">Ações</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant">
          {#each rows as row (row.id)}
            {#if editingId === row.id}
              <tr class="bg-secondary-container/20">
                <td class="px-3 py-2" colspan="5">
                  <div class="grid sm:grid-cols-6 gap-2">
                    <input class="ahq-input h-9 text-sm sm:col-span-2" bind:value={edit.name} placeholder="Nome" />
                    <input class="ahq-input h-9 text-sm font-mono sm:col-span-2" bind:value={edit.url} placeholder="Link" />
                    <select class="ahq-input h-9 text-sm" bind:value={edit.source}>
                      {#each sourceOptions as opt}
                        <option value={opt.value}>{opt.label}</option>
                      {/each}
                    </select>
                    <div class="grid grid-cols-2 gap-2 sm:contents">
                      <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.buy} placeholder="Comprar" />
                      <input class="ahq-input h-9 text-sm font-mono" bind:value={edit.sell} placeholder="Vender" />
                    </div>
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
                <td class="px-3 py-2.5">
                  <span class="ahq-badge-pending">{row.sourceLabel}</span>
                </td>
                <td class="px-3 py-2.5 min-w-0">
                  <a href={row.url} class="font-semibold text-primary hover:underline" target="_blank" rel="noreferrer">
                    {row.name}
                  </a>
                </td>
                <td class="px-3 py-2.5 text-right font-mono whitespace-nowrap">{row.buy}</td>
                <td class="px-3 py-2.5 text-right font-mono whitespace-nowrap">{row.sell}</td>
                <td class="px-3 py-2.5 text-right font-mono whitespace-nowrap font-semibold {marginClass(row.margin)}">
                  {row.margin}
                </td>
                <td class="px-3 py-2.5 text-right whitespace-nowrap">
                  <button type="button" class="text-secondary font-medium text-sm mr-2" on:click={() => startEdit(row)}>
                    Editar
                  </button>
                  <button type="button" class="text-error text-sm" on:click={() => destroy(row)}>Excluir</button>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</AppShell>
